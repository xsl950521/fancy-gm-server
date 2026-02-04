package routes

import (
	"path/filepath"
	"redis_data/api/handlers"
	"redis_data/api/middleware"
	"redis_data/config"
	"redis_data/pkg/auth"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, workDir string, appCfg *config.AppConfig, handlers *Handlers, jwtService *auth.JWTService, auditService interface{}) {
	// 静态文件服务
	r.Static("/static", filepath.Join(workDir, "web", "static"))
	r.StaticFile("/", filepath.Join(workDir, "web", "static", "index.html"))
	r.StaticFile("/history.html", filepath.Join(workDir, "web", "static", "history.html"))
	r.StaticFile("/analytics.html", filepath.Join(workDir, "web", "static", "analytics.html"))
	r.StaticFile("/missinglist.html", filepath.Join(workDir, "web", "static", "missinglist.html"))

	// API路由
	api := r.Group("/api")
	{
		// 认证相关路由（无需认证）
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", handlers.AuthHandler.Login)
			authGroup.POST("/register", handlers.AuthHandler.Register)
			authGroup.POST("/refresh", handlers.AuthHandler.RefreshToken)
		}

		// 需要认证的路由组
		authenticated := api.Group("")
		authenticated.Use(middleware.Auth(jwtService))
		{
			// 用户信息
			authenticated.GET("/auth/profile", handlers.AuthHandler.GetProfile)

			// 文件上传（需要file.upload权限）
			authenticated.POST("/upload",
				middleware.RequirePermission(auth.PermissionFileUpload),
				handlers.UploadHandler.UploadFiles)

			// 处理任务（需要data.process权限）
			authenticated.POST("/process/:jobId",
				middleware.RequirePermission(auth.PermissionDataProcess),
				handlers.ProcessHandler.ProcessJob)
			authenticated.GET("/status/:jobId", handlers.ProcessHandler.GetJobStatus)

			// 数据查询（需要data.view权限）
			authenticated.GET("/data",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.DataHandler.GetData)

			// 文件下载（需要file.download权限）
			authenticated.GET("/download/:jobId",
				middleware.RequirePermission(auth.PermissionFileDownload),
				handlers.DataHandler.DownloadFile)

			// 历史记录
			authenticated.GET("/history",
				middleware.RequirePermission(auth.PermissionHistoryView),
				handlers.HistoryHandler.GetHistories)
			authenticated.GET("/history/:id",
				middleware.RequirePermission(auth.PermissionHistoryView),
				handlers.HistoryHandler.GetHistory)
			authenticated.DELETE("/history/:id",
				middleware.RequirePermission(auth.PermissionHistoryDelete),
				handlers.HistoryHandler.DeleteHistory)

			// 数据分析
			authenticated.GET("/analytics/statistics",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.AnalyticsHandler.GetStatistics)
			authenticated.GET("/analytics/sheets",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.AnalyticsHandler.GetSheetList)

			// 漏发名单
			authenticated.POST("/missinglist/fetch",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.MissingListHandler.FetchMissingList)
			authenticated.POST("/missinglist/filter",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.MissingListHandler.FilterMissingList)
			authenticated.POST("/missinglist/group",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.MissingListHandler.GroupMissingList)
			authenticated.POST("/missinglist/parse-excel",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.MissingListHandler.ParseExcel)
			// 补发操作（高危操作，需要resend.mail权限）
			authenticated.POST("/missinglist/resend",
				middleware.RequirePermission(auth.PermissionResendMail),
				handlers.MissingListHandler.ResendMail)
			authenticated.POST("/missinglist/parse-lua",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.MissingListHandler.ParseLuaConfig)
			authenticated.POST("/missinglist/load-default-config",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.MissingListHandler.LoadDefaultConfig)
			authenticated.GET("/missinglist/item-map",
				middleware.RequirePermission(auth.PermissionDataView),
				handlers.MissingListHandler.GetItemMap)

			// 性能监控（可选，需要admin权限）
			if appCfg.Middleware.PerformanceMonitor.Enabled {
				authenticated.GET("/metrics",
					middleware.RequirePermission(auth.PermissionConfigView),
					handlers.MetricsHandler.GetMetrics)
				authenticated.POST("/metrics/reset",
					middleware.RequirePermission(auth.PermissionConfigModify),
					handlers.MetricsHandler.ResetMetrics)
			}
		}
	}
}

// Handlers 所有处理器集合
type Handlers struct {
	AuthHandler        *handlers.AuthHandler
	UploadHandler      *handlers.UploadHandler
	ProcessHandler     *handlers.ProcessHandler
	DataHandler        *handlers.DataHandler
	HistoryHandler     *handlers.HistoryHandler
	AnalyticsHandler   *handlers.AnalyticsHandler
	MissingListHandler *handlers.MissingListHandler
	MetricsHandler     *handlers.MetricsHandler
}

// SetupMiddleware 设置中间件
func SetupMiddleware(r *gin.Engine, appCfg *config.AppConfig) {
	// 添加中间件（顺序很重要）
	r.Use(middleware.Recovery())           // 恢复panic
	r.Use(middleware.RequestID())          // 请求ID

	// 安全中间件（在请求处理前检查）
	securityCfg := middleware.DefaultSecurityConfig()
	if appCfg.Middleware.Security.MaxRequestBodySize > 0 {
		securityCfg.MaxRequestBodySize = appCfg.Middleware.Security.MaxRequestBodySize
	}
	if appCfg.Middleware.Security.MaxHeaderSize > 0 {
		securityCfg.MaxHeaderSize = appCfg.Middleware.Security.MaxHeaderSize
	}
	r.Use(middleware.Security(securityCfg))

	// 性能监控中间件
	if appCfg.Middleware.PerformanceMonitor.Enabled {
		if appCfg.Middleware.PerformanceMonitor.SlowThreshold > 0 {
			middleware.SetSlowThreshold(appCfg.Middleware.PerformanceMonitor.SlowThreshold)
		}
		if appCfg.Middleware.PerformanceMonitor.MaxSlowRequests > 0 {
			middleware.SetMaxSlowRequests(appCfg.Middleware.PerformanceMonitor.MaxSlowRequests)
		}
		r.Use(middleware.PerformanceMonitor())
	}

	// 请求日志中间件（记录请求和响应）
	if appCfg.Middleware.RequestLogger.Enabled {
		r.Use(middleware.RequestLogger())
	}

	r.Use(middleware.ErrorHandler())       // 错误处理
	r.Use(middleware.CORS())               // CORS
}
