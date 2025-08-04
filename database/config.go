package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	logger "github.com/juanMaAV92/go-utils/log"
	"github.com/juanMaAV92/go-utils/path"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const (
	connectionRetries = 2
	retryDelay        = time.Second
	pgURLFormat       = "postgres://%s:%s@%s:%s/%s?sslmode=disable"
	dsnFormat         = "host=%s port=%s user=%s dbname=%s password=%s sslmode=disable"
)

var Logger = gormLogger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags),
	gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	},
)

var (
	singleton    *Database
	onceDB       sync.Once
	singletonErr error
)

func New(cfg DBConfig, logger logger.Logger) (*Database, error) {
	onceDB.Do(func() {
		gdb, err := connect(cfg)
		if err != nil {
			singletonErr = fmt.Errorf("failed to create database connection: %w", err)
			return
		}
		singleton = &Database{
			instance: gdb,
			logger:   logger,
		}
	})

	return singleton, singletonErr
}

func connect(cfg DBConfig) (*gorm.DB, error) {
	dsn := generateDSN(cfg)

	instance, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: Logger.LogMode(gormLogger.Error),
	})

	for i := 0; err != nil && i <= connectionRetries; i++ {
		instance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: Logger.LogMode(gormLogger.Error),
		})
		if err == nil {
			break
		}
		time.Sleep(retryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection after %d attempts: %w", connectionRetries, err)
	}

	db, err := instance.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	db.SetMaxIdleConns(cfg.MaxPoolSize)
	db.SetMaxOpenConns(cfg.MaxPoolSize)
	db.SetConnMaxLifetime(cfg.MaxLifeTime)

	if err := applyMigrations(cfg); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return instance, nil
}

func applyMigrations(cfg DBConfig) error {
	currentPath := fmt.Sprintf("file:///%s/migration", path.GetMainPath())
	url := generatePgURL(cfg)

	m, err := migrate.New(currentPath, url)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func generatePgURL(cfg DBConfig) string {
	return fmt.Sprintf(pgURLFormat, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func generateDSN(cfg DBConfig) string {
	return fmt.Sprintf(dsnFormat, cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Password)
}
