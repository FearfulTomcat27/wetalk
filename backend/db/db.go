package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"wetalk/config"
)

// DB 全局数据库实例
var DB *sql.DB

// Init 初始化数据库连接
func Init(cfg config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset,
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 配置连接池
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("数据库连通性测试失败: %w", err)
	}

	log.Println("数据库连接成功")
	return nil
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}
}
