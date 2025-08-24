package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DatabaseConfig PostgreSQL数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDatabaseConfig 从环境变量获取PostgreSQL数据库配置
func GetDatabaseConfig() *DatabaseConfig {
	config := &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "admin123"),
		DBName:   getEnv("DB_NAME", "todo_app"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	return config
}

// ConnectDatabase 连接PostgreSQL数据库
func ConnectDatabase(config *DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %v", err)
	}

	log.Printf("Successfully connected to PostgreSQL database")
	return db, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整型环境变量，如果不存在则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// InitializeDatabase 初始化数据库连接和表结构
func InitializeDatabase() (*sql.DB, error) {
	config := GetDatabaseConfig()

	db, err := ConnectDatabase(config)
	if err != nil {
		return nil, err
	}

	// 设置全局数据库连接
	// 注意：这里需要在global包中设置

	return db, nil
}
