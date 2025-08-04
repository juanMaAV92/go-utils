package database

import (
	"time"

	"github.com/juanMaAV92/go-utils/env"
	"github.com/juanMaAV92/go-utils/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultMaxPoolSize = 2
	defaultMaxLifeTime = "5m"
)

type Database struct {
	instance *gorm.DB
	logger   log.Logger
}

type DBConfig struct {
	Host        string
	LogLevel    logger.LogLevel
	MaxLifeTime time.Duration
	MaxPoolSize int
	Name        string
	Password    string
	Port        string
	User        string
}

type JoinConfig struct {
	BaseTable  string                 `json:"base_table"`
	Joins      []JoinClause           `json:"joins"`
	Conditions map[string]interface{} `json:"conditions"`
	Preloads   []string               `json:"preloads"`
	Select     string                 `json:"select"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	OrderBy    string                 `json:"order_by"`
}

type JoinClause struct {
	Type  string `json:"type"`  // "INNER", "LEFT", "RIGHT", "FULL"
	Table string `json:"table"` // Table to join
	On    string `json:"on"`    // Join condition
}

type PaginationOptions struct {
	Page  int `json:"page"`  // Page number (1-based)
	Limit int `json:"limit"` // Number of records per page
}

type QueryOptions struct {
	Pagination *PaginationOptions `json:"pagination,omitempty"`
	OrderBy    string             `json:"order_by,omitempty"`
	Preloads   []string           `json:"preloads,omitempty"`
}

func GetDBConfig() *DBConfig {
	return &DBConfig{
		Host:        env.GetEnv(env.PostgresHost),
		Password:    env.GetEnv(env.PostgresPassword),
		User:        env.GetEnv(env.PostgresUser),
		Port:        env.GetEnv(env.PostgresPort),
		Name:        env.GetEnv(env.PostgresName),
		LogLevel:    logger.Silent,
		MaxPoolSize: env.GetEnvAsIntWithDefault(env.PostgresMaxPoolSize, defaultMaxPoolSize),
		MaxLifeTime: env.GetEnvAsDurationWithDefault(env.PostgresMaxLifeTime, defaultMaxLifeTime),
	}
}
