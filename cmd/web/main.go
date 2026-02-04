package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"redis_data/api/handlers"
	"redis_data/api/middleware"
	"redis_data/api/models"
	"redis_data/config"
	"redis_data/service/history"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "", "path to web config json")
	flag.Parse()

	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	webCfg := config.DefaultWebConfig()
	if strings.TrimSpace(*configPath) != "" {
		loadedCfg, err := config.LoadWebConfig(*configPath)
		if err != nil {
			log.Fatal("Failed to load web config:", err)
		}
		webCfg = *loadedCfg
	} else {
		config.ApplyWebEnvOverrides(&webCfg)
	}

	applyWebEnv(&webCfg)

	// 设置存储目录
	uploadDir := filepath.Join(workDir, "storage", "uploads")
	resultDir := filepath.Join(workDir, "storage", "results")
	historyDir := filepath.Join(workDir, "storage", "history")

	// 确保目录存在
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(resultDir, 0755)
	os.MkdirAll(historyDir, 0755)

	// 创建任务管理器
	jobManager := models.NewJobManager()

	// 创建历史记录管理器
	historyManager, err := history.NewHistoryManager(historyDir)
	if err != nil {
		log.Printf("Warning: Failed to create history manager: %v", err)
		historyManager = nil
	}

	// 创建处理器
	uploadHandler := handlers.NewUploadHandler(jobManager, uploadDir)
	processHandler := handlers.NewProcessHandler(jobManager, historyManager, resultDir)
	dataHandler := handlers.NewDataHandler(jobManager, resultDir)
	historyHandler := handlers.NewHistoryHandler(historyManager)
	analyticsHandler := handlers.NewAnalyticsHandler(jobManager, resultDir)
	missingListHandler := handlers.NewMissingListHandler(&webCfg)

	// 创建Gin路由
	r := gin.Default()

	// 添加CORS中间件
	r.Use(middleware.CORS())

	// 静态文件服务
	r.Static("/static", filepath.Join(workDir, "web", "static"))
	r.StaticFile("/", filepath.Join(workDir, "web", "static", "index.html"))
	r.StaticFile("/history.html", filepath.Join(workDir, "web", "static", "history.html"))
	r.StaticFile("/analytics.html", filepath.Join(workDir, "web", "static", "analytics.html"))
	r.StaticFile("/missinglist.html", filepath.Join(workDir, "web", "static", "missinglist.html"))

	// API路由
	api := r.Group("/api")
	{
		// 文件上传
		api.POST("/upload", uploadHandler.UploadFiles)

		// 处理任务
		api.POST("/process/:jobId", processHandler.ProcessJob)
		api.GET("/status/:jobId", processHandler.GetJobStatus)

		// 数据查询
		api.GET("/data", dataHandler.GetData)

		// 文件下载
		api.GET("/download/:jobId", dataHandler.DownloadFile)

		// 历史记录
		api.GET("/history", historyHandler.GetHistories)
		api.GET("/history/:id", historyHandler.GetHistory)
		api.DELETE("/history/:id", historyHandler.DeleteHistory)

		// 数据分析
		api.GET("/analytics/statistics", analyticsHandler.GetStatistics)
		api.GET("/analytics/sheets", analyticsHandler.GetSheetList)

		// 漏发名单
		api.POST("/missinglist/fetch", missingListHandler.FetchMissingList)
		api.POST("/missinglist/filter", missingListHandler.FilterMissingList)
		api.POST("/missinglist/group", missingListHandler.GroupMissingList)
		api.POST("/missinglist/parse-excel", missingListHandler.ParseExcel)
		api.POST("/missinglist/resend", missingListHandler.ResendMail)
		api.POST("/missinglist/parse-lua", missingListHandler.ParseLuaConfig)
		api.POST("/missinglist/load-default-config", missingListHandler.LoadDefaultConfig)
		api.GET("/missinglist/item-map", missingListHandler.GetItemMap)
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", webCfg.Host, webCfg.Port)
	displayHost := webCfg.Host
	if displayHost == "0.0.0.0" || displayHost == "" {
		displayHost = "localhost"
	}
	fmt.Printf("Server starting on http://%s:%d\n", displayHost, webCfg.Port)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
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
