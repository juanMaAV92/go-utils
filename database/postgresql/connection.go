package postgresql

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/juanMaAV92/go-utils/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

const (
	connectionRetries = 2
	retryDelay        = time.Second
	pgURLFormat       = "postgres://%s:%s@%s:%s/%s?sslmode=%s"
	dsnFormat         = "host=%s port=%s user=%s dbname=%s password=%s sslmode=%s"
)

// New creates a Database backed by a new connection pool.
// Call once at service startup and inject the returned Database where needed.
func New(cfg Config, log logger.Logger) (Database, error) {
	gdb, err := connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("postgresql: %w", err)
	}
	return &database{instance: gdb, logger: log}, nil
}

// RunMigrations applies all pending up migrations from migrationsPath.
//
//	RunMigrations(cfg, "file:///app/migrations")
func RunMigrations(cfg Config, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, buildPgURL(cfg))
	if err != nil {
		return fmt.Errorf("postgresql: failed to create migration instance: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("postgresql: failed to apply migrations: %w", err)
	}
	return nil
}

func connect(cfg Config) (*gorm.DB, error) {
	dsn := buildDSN(cfg)
	gormCfg := &gorm.Config{Logger: buildGormLogger(cfg.Verbose)}

	var (
		instance *gorm.DB
		err      error
	)
	for i := 0; i <= connectionRetries; i++ {
		instance, err = gorm.Open(postgres.Open(dsn), gormCfg)
		if err == nil {
			break
		}
		if i < connectionRetries {
			time.Sleep(retryDelay)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open connection after %d attempts: %w", connectionRetries+1, err)
	}

	sqlDB, err := instance.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxPoolSize)
	sqlDB.SetMaxOpenConns(cfg.MaxPoolSize)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifeTime)

	if err := instance.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("failed to enable OTel tracing: %w", err)
	}

	return instance, nil
}

func buildGormLogger(verbose bool) gormLogger.Interface {
	if !verbose {
		return gormLogger.Default.LogMode(gormLogger.Silent)
	}
	return gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormLogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogger.Warn,
			IgnoreRecordNotFoundError: true,
		},
	)
}

func buildDSN(cfg Config) string {
	return fmt.Sprintf(dsnFormat, cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Password, resolveSSLMode(cfg.SSLMode))
}

func buildPgURL(cfg Config) string {
	return fmt.Sprintf(pgURLFormat, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, resolveSSLMode(cfg.SSLMode))
}

func resolveSSLMode(s string) string {
	if s == "" {
		return "require"
	}
	return s
}
