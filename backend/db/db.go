package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"wetalk/config"
)

// DB 全局数据库实例
var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层 sql.DB 失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连通性测试失败: %w", err)
	}

	log.Println("数据库连接成功")
	return nil
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("获取底层 sql.DB 失败: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}
}
