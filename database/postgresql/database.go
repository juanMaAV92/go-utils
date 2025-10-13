package postgresql

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

const (
	// Database operation steps for logging
	createStep    = "creating record"
	updateStep    = "updating record"
	findOneStep   = "finding single record"
	findManyStep  = "finding multiple records"
	countStep     = "counting records"
	joinQueryStep = "executing join query"
	rawQueryStep  = "executing raw query"
)

// Create inserts a new record into the database
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to the struct to be created (will be populated with generated values)
func (db *Database) Create(ctx context.Context, destination interface{}) (err error) {
	if err = validateContext(ctx); err != nil {
		return
	}

	if err = validateDestination(destination); err != nil {
		return
	}

	tx := db.instance.WithContext(ctx).Create(destination)
	if tx.Error != nil {
		err = handleDatabaseError(ctx, db.logger, tx.Error, createStep, "Failed to create record")
		return
	}

	return
}

// Update modifies an existing record in the database
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - model: Pointer to the model struct to be updated (identifies the record)
//   - updates: Map of field names to new values
//   - conditions: Additional conditions (map, struct, string with placeholders, or nil)
//   - args: Additional arguments for placeholder values when using string conditions
//
// Returns:
//   - int64: Number of records that were actually updated
//   - error: Any database error that occurred
func (db *Database) Update(ctx context.Context, model interface{}, updates map[string]interface{}, conditions interface{}, args ...interface{}) (rowsAffected int64, err error) {
	if err = validateContext(ctx); err != nil {
		return
	}

	if err = validateModel(model); err != nil {
		return
	}

	if err = validateUpdates(updates); err != nil {
		return
	}

	tx := db.instance.WithContext(ctx).Model(model)

	if conditions != nil {
		if len(args) > 0 {
			tx = tx.Where(conditions, args...)
		} else {
			tx = tx.Where(conditions)
		}
	}

	tx = tx.Updates(updates)
	if tx.Error != nil {
		err = handleDatabaseError(ctx, db.logger, tx.Error, updateStep, "Failed to update record")
		return
	}

	rowsAffected = tx.RowsAffected
	return
}

// FindOne retrieves a single record based on the given conditions
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to struct where the result will be stored
//   - conditions: Query conditions (map, struct, string with placeholders, or nil)
//   - preloads: Optional list of associations to preload
//   - args: Additional arguments for placeholder values when using string conditions
//
// Returns:
//   - bool: true if record was found, false if not found
//   - error: any database error that occurred
func (db *Database) FindOne(ctx context.Context, destination interface{}, conditions interface{}, preloads *[]string, args ...interface{}) (found bool, err error) {
	if err = validateContext(ctx); err != nil {
		return
	}

	if err = validateDestination(destination); err != nil {
		return
	}

	tx := db.instance.WithContext(ctx)

	if preloads != nil {
		for _, preload := range *preloads {
			tx = tx.Preload(preload)
		}
	}

	if conditions != nil {
		if len(args) > 0 {
			tx = tx.Where(conditions, args...)
		} else {
			tx = tx.Where(conditions)
		}
	}

	tx = tx.First(destination)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		err = handleDatabaseError(ctx, db.logger, tx.Error, findOneStep, "Failed to find record")
		return
	}

	found = true
	return
}

// FindMany retrieves multiple records based on the given conditions
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to slice where results will be stored
//   - conditions: Query conditions (map, struct, string with placeholders, or nil)
//   - options: Optional query parameters (pagination, ordering, preloads)
//   - args: Additional arguments for placeholder values when using string conditions
func (db *Database) FindMany(ctx context.Context, destination interface{}, conditions interface{}, options *QueryOptions, args ...interface{}) (err error) {
	if err = validateContext(ctx); err != nil {
		return
	}

	if err = validateDestination(destination); err != nil {
		return
	}

	tx := db.instance.WithContext(ctx)

	if conditions != nil {
		if len(args) > 0 {
			tx = tx.Where(conditions, args...)
		} else {
			tx = tx.Where(conditions)
		}
	}

	if options != nil {
		tx = applyQueryOptions(tx, options)
	}

	if err = tx.Find(destination).Error; err != nil {
		err = handleDatabaseError(ctx, db.logger, err, findManyStep, "Failed to find records")
		return
	}

	return
}

// Count returns the total number of records matching the given conditions
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - model: Model struct or table name to count records from
//   - conditions: Query conditions (map, struct, string with placeholders, or nil)
//   - args: Additional arguments for placeholder values when using string conditions
func (db *Database) Count(ctx context.Context, model interface{}, conditions interface{}, args ...interface{}) (totalRecords int64, err error) {
	if err = validateContext(ctx); err != nil {
		return
	}

	if err = validateModel(model); err != nil {
		return
	}

	tx := db.instance.WithContext(ctx).Model(model)

	if conditions != nil {
		if len(args) > 0 {
			tx = tx.Where(conditions, args...)
		} else {
			tx = tx.Where(conditions)
		}
	}

	if err = tx.Count(&totalRecords).Error; err != nil {
		err = handleDatabaseError(ctx, db.logger, err, countStep, "Failed to count records")
		return
	}

	return
}

// FindWithJoins executes complex queries with JOIN operations
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to slice where results will be stored
//   - config: Join configuration defining tables, conditions, and options
func (db *Database) FindWithJoins(ctx context.Context, destination interface{}, config JoinConfig) (err error) {
	if err = validateContext(ctx); err != nil {
		return
	}

	if err = validateDestination(destination); err != nil {
		return
	}

	if err = validateJoinConfig(config); err != nil {
		return
	}

	tx := db.instance.WithContext(ctx).Table(config.BaseTable)

	for _, join := range config.Joins {
		tx = applyJoin(tx, join)
	}

	tx = applyJoinOptions(tx, config)

	if err = tx.Find(destination).Error; err != nil {
		err = handleDatabaseError(ctx, db.logger, err, joinQueryStep, "Failed to execute join query")
		return
	}

	return
}

// ExecuteRawQuery executes a raw SQL query with parameters and returns detailed results
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer where results will be stored (nil for non-SELECT queries)
//   - query: SQL query string with placeholders (?)
//   - args: Query arguments to replace placeholders
//
// Returns:
//   - QueryResult: Contains RowsAffected (for INSERT/UPDATE/DELETE) and Found (for SELECT)
//   - error: Any database error that occurred
func (db *Database) ExecuteRawQuery(ctx context.Context, destination interface{}, query string, args ...interface{}) (result QueryResult, err error) {
	if err = validateContext(ctx); err != nil {
		return
	}

	if err = validateQuery(query); err != nil {
		return
	}

	var tx *gorm.DB

	if destination == nil {
		// For non-SELECT queries (INSERT, UPDATE, DELETE)
		tx = db.instance.WithContext(ctx).Exec(query, args...)
		if tx.Error != nil {
			err = handleDatabaseError(ctx, db.logger, tx.Error, rawQueryStep, "Failed to execute raw query")
			return
		}
		result.RowsAffected = tx.RowsAffected
		result.Found = tx.RowsAffected > 0
	} else {
		// For SELECT queries
		if err = validateDestination(destination); err != nil {
			return
		}

		tx = db.instance.WithContext(ctx).Raw(query, args...).Scan(destination)
		if tx.Error != nil {
			err = handleDatabaseError(ctx, db.logger, tx.Error, rawQueryStep, "Failed to execute raw query")
			return
		}
		result.RowsAffected = tx.RowsAffected
		result.Found = tx.RowsAffected > 0
	}

	return
}

// Validation methods for consistent parameter checking

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New(errContextRequired)
	}
	return nil
}

func validateDestination(destination interface{}) error {
	if destination == nil {
		return errors.New(errDestinationRequired)
	}

	if reflect.TypeOf(destination).Kind() != reflect.Ptr {
		return errors.New(errDestinationMustBePointer)
	}

	return nil
}

func validateModel(model interface{}) error {
	if model == nil {
		return errors.New(errModelRequired)
	}
	return nil
}

func validateUpdates(updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New(errUpdatesRequired)
	}
	return nil
}

func validateJoinConfig(config JoinConfig) error {
	if config.BaseTable == "" {
		return errors.New(errBaseTableRequired)
	}

	if len(config.Joins) == 0 {
		return errors.New(errJoinsRequired)
	}

	return nil
}

func validateQuery(query string) error {
	if query == "" {
		return errors.New(errQueryRequired)
	}
	return nil
}

// Helper methods

func applyQueryOptions(tx *gorm.DB, options *QueryOptions) *gorm.DB {
	if options.OrderBy != "" {
		tx = tx.Order(options.OrderBy)
	}

	for _, preload := range options.Preloads {
		tx = tx.Preload(preload)
	}

	if options.Pagination != nil {
		tx = applyPagination(tx, *options.Pagination)
	}

	return tx
}

func applyPagination(tx *gorm.DB, pagination PaginationOptions) *gorm.DB {
	page := pagination.Page
	limit := pagination.Limit

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	return tx.Offset(offset).Limit(limit)
}

func applyJoin(tx *gorm.DB, join JoinClause) *gorm.DB {
	joinSQL := fmt.Sprintf("%s JOIN %s ON %s", join.Type, join.Table, join.On)
	return tx.Joins(joinSQL)
}

func applyJoinOptions(tx *gorm.DB, config JoinConfig) *gorm.DB {
	if len(config.Conditions) > 0 {
		tx = tx.Where(config.Conditions)
	}

	if config.Select != "" {
		tx = tx.Select(config.Select)
	}

	for _, preload := range config.Preloads {
		tx = tx.Preload(preload)
	}

	if config.OrderBy != "" {
		tx = tx.Order(config.OrderBy)
	}

	if config.Limit > 0 {
		tx = tx.Limit(config.Limit)
	}

	if config.Offset > 0 {
		tx = tx.Offset(config.Offset)
	}

	return tx
}
