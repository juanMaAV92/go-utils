package postgresql

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/juanmaAV/go-utils/logger"
	"gorm.io/gorm"
)

const (
	pgUniqueViolation     = "23505"
	pgCheckViolation      = "23514"
	pgForeignKeyViolation = "23503"

	msgFailedToCreate     = "failed to create record"
	msgFailedToCreateMany = "failed to create records"
	msgFailedToFind       = "failed to find record"
	msgFailedToUpdate     = "failed to update record"
	msgFailedToDelete     = "failed to delete record"
	msgFailedToCount      = "failed to count records"
	msgFailedToExec       = "failed to execute query"

	stepCreate     = "db.create"
	stepCreateMany = "db.create_many"
	stepFind       = "db.find"
	stepFindMany   = "db.find_many"
	stepUpdate     = "db.update"
	stepDelete     = "db.delete"
	stepCount      = "db.count"
	stepExec       = "db.exec"
)

func (db *database) Create(ctx context.Context, model any) (int64, error) {
	if err := validate(ctx, model); err != nil {
		return 0, err
	}
	tx := db.instance.WithContext(ctx).Create(model)
	if tx.Error != nil {
		return 0, handleDBError(ctx, db.logger, tx.Error, stepCreate, msgFailedToCreate)
	}
	return tx.RowsAffected, nil
}

func (db *database) CreateMany(ctx context.Context, models any, batchSize int) (int64, error) {
	if err := validate(ctx, models); err != nil {
		return 0, err
	}
	if err := validateSlice(models); err != nil {
		return 0, err
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	tx := db.instance.WithContext(ctx).CreateInBatches(models, batchSize)
	if tx.Error != nil {
		return 0, handleDBError(ctx, db.logger, tx.Error, stepCreateMany, msgFailedToCreateMany)
	}
	return tx.RowsAffected, nil
}

// Find retrieves the first record matching conditions into model.
// Returns found=false (no error) when no record matches.
// preloads lists relations to eager-load (e.g. []string{"Profile", "Orders"}).
func (db *database) Find(ctx context.Context, model any, preloads []string, conditions any, args ...any) (bool, error) {
	if err := validate(ctx, model); err != nil {
		return false, err
	}
	tx := db.instance.WithContext(ctx)
	for _, p := range preloads {
		tx = tx.Preload(p)
	}
	if conditions != nil {
		tx = tx.Where(conditions, args...)
	}
	if err := tx.First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, handleDBError(ctx, db.logger, err, stepFind, msgFailedToFind)
	}
	return true, nil
}

// FindMany retrieves all records matching conditions into model (must be a pointer to a slice).
// Returns found=false (no error) when the result set is empty.
func (db *database) FindMany(ctx context.Context, model any, options *QueryOptions, conditions any, args ...any) (bool, error) {
	if err := validate(ctx, model); err != nil {
		return false, err
	}
	tx := db.instance.WithContext(ctx)
	if conditions != nil {
		tx = tx.Where(conditions, args...)
	}
	if options != nil {
		tx = applyQueryOptions(tx, options)
	}
	if err := tx.Find(model).Error; err != nil {
		return false, handleDBError(ctx, db.logger, err, stepFindMany, msgFailedToFind)
	}
	return tx.RowsAffected > 0, nil
}

// UpdateWhere updates fields on model's table for rows matching conditions.
//
// Pass updates as map[string]any to update specific fields including zero values.
// Passing a struct skips zero-value fields — use map when zero values matter.
func (db *database) UpdateWhere(ctx context.Context, model any, updates any, conditions any, args ...any) (int64, error) {
	if err := validate(ctx, model); err != nil {
		return 0, err
	}
	if err := validateUpdates(updates); err != nil {
		return 0, err
	}
	tx := db.instance.WithContext(ctx).Model(model)
	if conditions != nil {
		tx = tx.Where(conditions, args...)
	}
	if err := tx.Updates(updates).Error; err != nil {
		return 0, handleDBError(ctx, db.logger, err, stepUpdate, msgFailedToUpdate)
	}
	return tx.RowsAffected, nil
}

// Delete removes records from model's table matching conditions.
// If model has a DeletedAt field (gorm.DeletedAt), GORM performs a soft delete automatically.
// Passing nil conditions deletes by the model's primary key value.
func (db *database) Delete(ctx context.Context, model any, conditions any, args ...any) (int64, error) {
	if err := validate(ctx, model); err != nil {
		return 0, err
	}
	tx := db.instance.WithContext(ctx)
	if conditions != nil {
		tx = tx.Where(conditions, args...)
	}
	result := tx.Delete(model)
	if result.Error != nil {
		return 0, handleDBError(ctx, db.logger, result.Error, stepDelete, msgFailedToDelete)
	}
	return result.RowsAffected, nil
}

// Count returns the number of records in model's table matching conditions.
func (db *database) Count(ctx context.Context, model any, options *QueryOptions, conditions any, args ...any) (int64, error) {
	if err := validate(ctx, model); err != nil {
		return 0, err
	}
	tx := db.instance.WithContext(ctx).Model(model)
	if options != nil {
		for _, join := range options.Joins {
			tx = tx.Joins(join)
		}
	}
	if conditions != nil {
		tx = tx.Where(conditions, args...)
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, handleDBError(ctx, db.logger, err, stepCount, msgFailedToCount)
	}
	return count, nil
}

// Exec runs a raw SQL statement.
//
// Pass model=nil for non-SELECT statements (INSERT, UPDATE, DELETE).
// Pass a pointer to a slice for SELECT statements — results are scanned into it.
func (db *database) Exec(ctx context.Context, model any, sql string, args ...any) (QueryResult, error) {
	if ctx == nil {
		return QueryResult{}, errors.New("context is required")
	}
	if model != nil && reflect.TypeOf(model).Kind() != reflect.Ptr {
		return QueryResult{}, errors.New("model must be a pointer")
	}

	var tx *gorm.DB
	if model == nil {
		tx = db.instance.WithContext(ctx).Exec(sql, args...)
	} else {
		tx = db.instance.WithContext(ctx).Raw(sql, args...).Scan(model)
	}
	if tx.Error != nil {
		return QueryResult{}, handleDBError(ctx, db.logger, tx.Error, stepExec, msgFailedToExec)
	}
	return QueryResult{RowsAffected: tx.RowsAffected, Found: tx.RowsAffected > 0}, nil
}

// WithTransaction runs fn inside a database transaction.
// fn receives a Database scoped to the transaction.
// Return nil to commit, return an error to rollback.
func (db *database) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if fn == nil {
		return errors.New("transaction function is required")
	}
	return db.instance.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&database{instance: tx, logger: db.logger})
	})
}

// --- internal helpers ---

func validate(ctx context.Context, model any) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if model == nil {
		return errors.New("model is required")
	}
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		return errors.New("model must be a pointer")
	}
	return nil
}

func validateSlice(model any) error {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		return errors.New("model must be a pointer to a slice")
	}
	return nil
}

func validateUpdates(updates any) error {
	if updates == nil {
		return errors.New("updates are required")
	}
	v := reflect.ValueOf(updates)
	if v.Kind() == reflect.Map && v.Len() == 0 {
		return errors.New("updates map must not be empty")
	}
	return nil
}

func applyQueryOptions(tx *gorm.DB, o *QueryOptions) *gorm.DB {
	if o.OrderBy != "" {
		tx = tx.Order(o.OrderBy)
	}
	for _, p := range o.Preloads {
		tx = tx.Preload(p)
	}
	for _, j := range o.Joins {
		tx = tx.Joins(j)
	}
	if o.Pagination != nil {
		page, limit := o.Pagination.Page, o.Pagination.Limit
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		tx = tx.Offset((page - 1) * limit).Limit(limit)
	}
	return tx
}

func handleDBError(ctx context.Context, log logger.Logger, err error, step, message string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	if log != nil {
		log.Error(ctx, step, message, "error", err.Error())
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return fmt.Errorf("%w: %s", ErrDuplicateRecord, pgErr.Detail)
		case pgCheckViolation:
			return fmt.Errorf("%w: %s", ErrConstraintViolation, pgErr.ConstraintName)
		case pgForeignKeyViolation:
			return fmt.Errorf("%w: %s", ErrInvalidReference, pgErr.Detail)
		}
	}

	return errors.New(message)
}
