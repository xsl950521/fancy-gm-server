package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"redis_data/api/handlers"
	"redis_data/api/routes"
	"redis_data/api/models"
	"redis_data/config"
	"redis_data/pkg/auth"
	"redis_data/pkg/audit"
	"redis_data/internal/model"
	"redis_data/internal/repository"
	"redis_data/pkg/database"
	"redis_data/pkg/logger"
	"redis_data/service/history"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	configPath := flag.String("config", "", "path to config file (yaml/json/toml)")
	flag.Parse()

	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get working directory: %v", err))
	}

	// 加载配置（使用Viper，支持多格式和环境变量）
	appCfg, err := config.LoadAppConfig(*configPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		logger.Logger.Warn("Failed to load config, using defaults", zap.Error(err))
		defaultCfg := config.DefaultAppConfig()
		appCfg = &defaultCfg
	}

	// 初始化日志系统（先初始化日志，后续错误才能记录）
	if err := logger.Init(appCfg.Log); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Logger.Info("Starting GM Server",
		zap.String("config_file", *configPath),
	)

	// 使用配置中的Web配置
	webCfg := appCfg.Web
	applyWebEnv(&webCfg)

	// 使用配置中的数据库配置
	dbCfg := appCfg.Database

	// 创建GORM logger
	var gormLog gormlogger.Interface
	if appCfg.Log.Level == "debug" {
		gormLog = gormlogger.Default.LogMode(gormlogger.Info)
	} else {
		gormLog = gormlogger.Default.LogMode(gormlogger.Silent)
	}

	db, err := database.Connect(dbCfg, gormLog)
	if err != nil {
		logger.Logger.Fatal("Failed to connect database", zap.Error(err))
	}

	// 执行数据库迁移
	if err := database.Migrate(db,
		&model.Job{},
		&model.JobFile{},
		&model.JobResult{},
		&model.History{},
		&model.HistoryFile{},
		&model.HistoryResult{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.AuditLog{},
	); err != nil {
		logger.Logger.Fatal("Failed to migrate database", zap.Error(err))
	}
	logger.Logger.Info("Database migrated successfully")

	// 初始化默认角色和权限
	if err := initDefaultRolesAndPermissions(db, appCfg); err != nil {
		logger.Logger.Warn("Failed to init default roles and permissions", zap.Error(err))
	}

	// 设置存储目录
	uploadDir := filepath.Join(workDir, "storage", "uploads")
	resultDir := filepath.Join(workDir, "storage", "results")
	historyDir := filepath.Join(workDir, "storage", "history")

	// 确保目录存在
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(resultDir, 0755)
	os.MkdirAll(historyDir, 0755)

	// 创建Repository
	jobRepo := repository.NewJobRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// 创建任务管理器
	jobManager := models.NewJobManager(jobRepo)

	// 创建历史记录管理器
	historyManager := history.NewHistoryManager(historyRepo)

	// 创建认证服务
	jwtConfig := auth.JWTConfig{
		SecretKey:  appCfg.Auth.JWT.SecretKey,
		Expiration: appCfg.Auth.JWT.Expiration,
		RefreshExp: appCfg.Auth.JWT.RefreshExp,
		Issuer:     appCfg.Auth.JWT.Issuer,
	}
	jwtService := auth.NewJWTService(jwtConfig)
	authService := auth.NewAuthService(userRepo, jwtService)

	// 创建审计日志服务
	auditService := audit.NewAuditService(auditRepo)

	// 创建处理器
	authHandler := handlers.NewAuthHandler(authService)
	uploadHandler := handlers.NewUploadHandler(jobManager, uploadDir)
	processHandler := handlers.NewProcessHandler(jobManager, historyManager, resultDir)
	dataHandler := handlers.NewDataHandler(jobManager, resultDir)
	historyHandler := handlers.NewHistoryHandler(historyManager)
	analyticsHandler := handlers.NewAnalyticsHandler(jobManager, resultDir)
	missingListHandler := handlers.NewMissingListHandler(&webCfg)

	// 创建Gin路由
	if appCfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// 设置中间件
	routes.SetupMiddleware(r, appCfg)

	// 创建处理器集合
	metricsHandler := handlers.NewMetricsHandler()
	allHandlers := &routes.Handlers{
		AuthHandler:        authHandler,
		UploadHandler:      uploadHandler,
		ProcessHandler:     processHandler,
		DataHandler:        dataHandler,
		HistoryHandler:     historyHandler,
		AnalyticsHandler:   analyticsHandler,
		MissingListHandler: missingListHandler,
		MetricsHandler:     metricsHandler,
	}

	// 设置路由
	routes.SetupRoutes(r, workDir, appCfg, allHandlers, jwtService, auditService)

	// 启动服务器
	// 优先使用server配置，如果没有则使用web配置
	serverHost := appCfg.Server.Host
	serverPort := appCfg.Server.Port
	if serverHost == "" {
		serverHost = webCfg.Host
	}
	if serverPort <= 0 {
		serverPort = webCfg.Port
	}

	addr := fmt.Sprintf("%s:%d", serverHost, serverPort)
	displayHost := serverHost
	if displayHost == "0.0.0.0" || displayHost == "" {
		displayHost = "localhost"
	}
	logger.Logger.Info("Server starting",
		zap.String("host", displayHost),
		zap.Int("port", serverPort),
		zap.String("address", addr),
		zap.String("db_host", dbCfg.Host),
		zap.String("db_name", dbCfg.DBName),
	)
	if err := r.Run(addr); err != nil {
		logger.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func getEnvInt(name string) *int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &val
}

func applyWebEnv(cfg *config.WebConfig) {
	setEnvIfEmpty("RESEND_BATCH_SIZE", strconv.Itoa(cfg.ResendBatchSize))
	setEnvIfEmpty("RESEND_BATCH_WORKERS", strconv.Itoa(cfg.ResendBatchWorkers))
	setEnvIfEmpty("RESEND_RANK_WORKERS", strconv.Itoa(cfg.ResendRankWorkers))
	setEnvIfEmpty("RESEND_HTTP_TIMEOUT_SEC", strconv.Itoa(cfg.ResendHTTPTimeoutSec))
}

func setEnvIfEmpty(key string, value string) {
	if strings.TrimSpace(os.Getenv(key)) == "" {
		_ = os.Setenv(key, value)
	}
}
