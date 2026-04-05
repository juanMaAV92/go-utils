package postgresql

import (
	"context"

	"github.com/juanmaAV/go-utils/logger"
	"gorm.io/gorm"
)

// Database is the interface for all database operations.
type Database interface {
	Create(ctx context.Context, model any) (affectedRows int64, err error)
	CreateMany(ctx context.Context, models any, batchSize int) (affectedRows int64, err error)
	Find(ctx context.Context, model any, preloads []string, conditions any, args ...any) (found bool, err error)
	FindMany(ctx context.Context, model any, options *QueryOptions, conditions any, args ...any) (found bool, err error)
	UpdateWhere(ctx context.Context, model any, updates any, conditions any, args ...any) (affectedRows int64, err error)
	Delete(ctx context.Context, model any, conditions any, args ...any) (affectedRows int64, err error)
	Count(ctx context.Context, model any, options *QueryOptions, conditions any, args ...any) (count int64, err error)
	Exec(ctx context.Context, model any, sql string, args ...any) (result QueryResult, err error)
	WithTransaction(ctx context.Context, fn TransactionFunc) error
}

type database struct {
	instance *gorm.DB
	logger   logger.Logger
}

// PaginationOptions defines 1-based page/limit pagination.
type PaginationOptions struct {
	Page  int // 1-based; defaults to 1
	Limit int // records per page; defaults to 10
}

// QueryOptions controls ordering, preloading, joining, and pagination for FindMany and Count.
type QueryOptions struct {
	Pagination *PaginationOptions
	OrderBy    string
	Preloads   []string
	Joins      []string
}

// QueryResult holds the outcome of an Exec call.
type QueryResult struct {
	RowsAffected int64
	Found        bool
}

// TransactionFunc is the callback passed to WithTransaction.
// Return nil to commit, return an error to rollback.
type TransactionFunc func(tx Database) error
