package postgresql

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgconn"
	"github.com/juanMaAV92/go-utils/log"
	"gorm.io/gorm"
)

const (
	// Database operation steps for logging
	createStep      = "creating record"
	updateStep      = "updating record"
	findOneStep     = "finding single record"
	findManyStep    = "finding multiple records"
	countStep       = "counting records"
	joinQueryStep   = "executing join query"
	rawQueryStep    = "executing raw query"
	transactionStep = "transaction"

	// Validation error messages
	errContextRequired          = "context is required"
	errDestinationRequired      = "destination is required"
	errDestinationMustBePointer = "destination must be a pointer"
	errModelRequired            = "model is required"
	errUpdatesRequired          = "updates are required"
	errConditionRequired        = "condition is required"
	errConfigRequired           = "configuration is required"
	errQueryRequired            = "query is required"
	errBaseTableRequired        = "base table is required"
	errJoinsRequired            = "at least one join is required"
	errTransactionFnRequired    = "transaction function is required"

	// PostgreSQL error codes
	pgUniqueViolation     = "23505"
	pgCheckViolation      = "23514"
	pgForeignKeyViolation = "23503"

	// User-friendly error messages
	msgDuplicateRecord     = "a record with the same values already exists"
	msgConstraintViolation = "the provided data violates database constraints"
	msgInvalidReference    = "invalid reference in the provided data"
	msgDatabaseError       = "an unexpected database error occurred"
)

// Create inserts a new record into the database
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to the struct to be created (will be populated with generated values)
func (db *Database) Create(ctx context.Context, destination interface{}) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateDestination(destination); err != nil {
		return err
	}

	tx := db.instance.WithContext(ctx).Create(destination)
	if tx.Error != nil {
		return handleDatabaseError(ctx, db.logger, tx.Error, createStep, "Failed to create record")
	}

	return nil
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
func (db *Database) Update(ctx context.Context, model interface{}, updates map[string]interface{}, conditions interface{}, args ...interface{}) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}

	if err := validateModel(model); err != nil {
		return 0, err
	}

	if err := validateUpdates(updates); err != nil {
		return 0, err
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
		return 0, handleDatabaseError(ctx, db.logger, tx.Error, updateStep, "Failed to update record")
	}

	return tx.RowsAffected, nil
}

// FindOne retrieves a single record based on the given conditions
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to struct where the result will be stored
//   - conditions: Query conditions (map, struct, or string)
//
// Returns:
//   - bool: true if record was found, false if not found
//   - error: any database error that occurred
func (db *Database) FindOne(ctx context.Context, destination interface{}, conditions interface{}, preloads *[]string) (bool, error) {
	if err := validateContext(ctx); err != nil {
		return false, err
	}

	if err := validateDestination(destination); err != nil {
		return false, err
	}

	if err := validateConditions(conditions); err != nil {
		return false, err
	}

	tx := db.instance.WithContext(ctx)

	if preloads != nil {
		for _, preload := range *preloads {
			tx = tx.Preload(preload)
		}
	}

	tx = tx.Where(conditions).First(destination)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, handleDatabaseError(ctx, db.logger, tx.Error, findOneStep, "Failed to find record")
	}

	return true, nil
}

// FindMany retrieves multiple records based on the given conditions
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to slice where results will be stored
//   - conditions: Query conditions (map, struct, or string)
//   - options: Optional query parameters (pagination, ordering, preloads)
func (db *Database) FindMany(ctx context.Context, destination interface{}, conditions interface{}, options *QueryOptions) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateDestination(destination); err != nil {
		return err
	}

	tx := db.instance.WithContext(ctx)

	if conditions != nil {
		tx = tx.Where(conditions)
	}

	if options != nil {
		tx = applyQueryOptions(tx, options)
	}

	if err := tx.Find(destination).Error; err != nil {
		return handleDatabaseError(ctx, db.logger, err, findManyStep, "Failed to find records")
	}

	return nil
}

// Count returns the total number of records matching the given conditions
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - model: Model struct or table name to count records from
//   - conditions: Query conditions (map, struct, string with placeholders, or nil)
//   - args: Additional arguments for placeholder values when using string conditions
func (db *Database) Count(ctx context.Context, model interface{}, conditions interface{}, args ...interface{}) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}

	if err := validateModel(model); err != nil {
		return 0, err
	}

	tx := db.instance.WithContext(ctx).Model(model)

	if conditions != nil {
		if len(args) > 0 {
			tx = tx.Where(conditions, args...)
		} else {
			tx = tx.Where(conditions)
		}
	}

	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, handleDatabaseError(ctx, db.logger, err, countStep, "Failed to count records")
	}

	return count, nil
}

// FindWithJoins executes complex queries with JOIN operations
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer to slice where results will be stored
//   - config: Join configuration defining tables, conditions, and options
func (db *Database) FindWithJoins(ctx context.Context, destination interface{}, config JoinConfig) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateDestination(destination); err != nil {
		return err
	}

	if err := validateJoinConfig(config); err != nil {
		return err
	}

	tx := db.instance.WithContext(ctx).Table(config.BaseTable)

	for _, join := range config.Joins {
		tx = applyJoin(tx, join)
	}

	tx = applyJoinOptions(tx, config)

	if err := tx.Find(destination).Error; err != nil {
		return handleDatabaseError(ctx, db.logger, err, joinQueryStep, "Failed to execute join query")
	}

	return nil
}

// ExecuteRawQuery executes a raw SQL query with parameters
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - destination: Pointer where results will be stored
//   - query: SQL query string with placeholders (?)
//   - args: Query arguments to replace placeholders
func (db *Database) ExecuteRawQuery(ctx context.Context, destination interface{}, query string, args ...interface{}) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateDestination(destination); err != nil {
		return err
	}

	if err := validateQuery(query); err != nil {
		return err
	}

	var tx *gorm.DB

	if destination == nil {
		tx = db.instance.WithContext(ctx).Exec(query, args...)
	} else {
		tx = db.instance.WithContext(ctx).Raw(query, args...).Scan(destination)
	}

	if tx.Error != nil {
		return handleDatabaseError(ctx, db.logger, tx.Error, rawQueryStep, "Failed to execute raw query")
	}

	return nil
}

// WithTransaction executes a function within a database transaction
// The transaction is automatically committed if the function returns nil,
// or rolled back if the function returns an error or panics.
//
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - fn: Function to execute within the transaction. Receives a Database instance
//         that operates within the transaction context
//
// Returns:
//   - error: Any error from beginning the transaction, executing the function,
//           or committing the transaction

func (db *Database) WithTransaction(ctx context.Context, fn func(*Database) error) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if fn == nil {
		return errors.New(errTransactionFnRequired)
	}

	db.logger.Debug(ctx, transactionStep, "Starting database transaction")

	tx := db.instance.WithContext(ctx).Begin()
	if tx.Error != nil {
		return handleDatabaseError(ctx, db.logger, tx.Error, transactionStep, "Failed to begin transaction")
	}

	txDB := &Database{
		instance: tx,
		logger:   db.logger,
	}

	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				db.logger.Error(ctx, transactionStep, "Failed to rollback after panic",
					log.Field("panic", r), log.Field("rollback_error", rollbackErr.Error()))
			} else {
				db.logger.Debug(ctx, transactionStep, "Transaction rolled back due to panic",
					log.Field("panic", r))
			}
			panic(r)
		}
	}()

	if err := fn(txDB); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			db.logger.Error(ctx, transactionStep, "Failed to rollback transaction",
				log.Field("original_error", err.Error()), log.Field("rollback_error", rollbackErr.Error()))
		} else {
			db.logger.Debug(ctx, transactionStep, "Transaction rolled back due to error",
				log.Field("error", err.Error()))
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return handleDatabaseError(ctx, db.logger, err, transactionStep, "Failed to commit transaction")
	}

	db.logger.Debug(ctx, transactionStep, "Transaction committed successfully")
	return nil
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

func validateConditions(conditions interface{}) error {
	if conditions == nil {
		return errors.New(errConditionRequired)
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

// Error handling
func handleDatabaseError(ctx context.Context, logger log.Logger, err error, step, message string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	logger.Error(ctx, step, message, log.Field("error", err.Error()))

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case pgUniqueViolation:
			return errors.New(msgDuplicateRecord)
		case pgCheckViolation:
			return errors.New(msgConstraintViolation)
		case pgForeignKeyViolation:
			return errors.New(msgInvalidReference)
		}
	}

	return errors.New(msgDatabaseError)
}
