package test

import (
	"os"
	"path/filepath"
	"redis_data/config"
	"redis_data/pkg/database"
	"redis_data/pkg/logger"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SetupTestDB 设置测试数据库
func SetupTestDB() (*gorm.DB, error) {
	// 加载测试配置
	testConfigPath := filepath.Join("test", "test_config.yaml")
	appCfg, err := config.LoadAppConfig(testConfigPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		defaultCfg := config.DefaultAppConfig()
		appCfg = &defaultCfg
		appCfg.Database.DBName = "gm_server_test"
	}

	// 初始化日志
	if err := logger.Init(appCfg.Log); err != nil {
		return nil, err
	}

	// 创建GORM logger
	gormLog := gormlogger.Default.LogMode(gormlogger.Silent)

	// 连接数据库
	db, err := database.Connect(appCfg.Database, gormLog)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetTestConfig 获取测试配置
func GetTestConfig() (*config.AppConfig, error) {
	testConfigPath := filepath.Join("test", "test_config.yaml")
	appCfg, err := config.LoadAppConfig(testConfigPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		defaultCfg := config.DefaultAppConfig()
		appCfg = &defaultCfg
		appCfg.Database.DBName = "gm_server_test"
	}
	return appCfg, nil
}

// GetTestWorkDir 获取测试工作目录
func GetTestWorkDir() string {
	workDir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return workDir
}
