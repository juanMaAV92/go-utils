package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/juanMaAV92/go-utils/log"
	"gorm.io/gorm"
)

const (
	errContextRequired          = "context is required"
	errDestinationRequired      = "destination is required"
	errDestinationMustBePointer = "destination must be a pointer"
	errModelRequired            = "model is required"
	errUpdatesRequired          = "updates are required"
	errQueryRequired            = "query is required"
	errBaseTableRequired        = "base table is required"
	errJoinsRequired            = "at least one join is required"

	// PostgreSQL error codes
	pgUniqueViolation     = "23505"
	pgCheckViolation      = "23514"
	pgForeignKeyViolation = "23503"

	// User-friendly error messages
	msgDuplicateRecord             = "a record with the same values already exists"
	msgConstraintViolation         = "the provided data violates database constraints"
	msgInvalidReference            = "invalid reference in the provided data"
	msgDatabaseError               = "an unexpected database error occurred"
	msgTransactionFunctionRequired = "transaction function is required"
	msgFailedToCreateRecord        = "Failed to create record"
	msgFailedToUpdateRecord        = "Failed to update record"
	msgFailedToFindRecord          = "Failed to find record"
	msgFailedToFindRecords         = "Failed to find records"
	msgFailedToCountRecords        = "Failed to count records"
	msgFailedToExecuteJoinQuery    = "Failed to execute join query"
	msgFailedToExecuteRawQuery     = "Failed to execute raw query"
	msgFailedToBeginTransaction    = "Failed to begin transaction"
	msgFailedToCommitTransaction   = "Failed to commit transaction"
	msgFailedToRollbackTransaction = "Failed to rollback transaction"
	msgTransactionPanic            = "transaction panic: %v"
)

// Error creation functions
func newContextRequiredError() error {
	return errors.New(errContextRequired)
}

func newDestinationRequiredError() error {
	return errors.New(errDestinationRequired)
}

func newDestinationMustBePointerError() error {
	return errors.New(errDestinationMustBePointer)
}

func newModelRequiredError() error {
	return errors.New(errModelRequired)
}

func newUpdatesRequiredError() error {
	return errors.New(errUpdatesRequired)
}

func newQueryRequiredError() error {
	return errors.New(errQueryRequired)
}

func newBaseTableRequiredError() error {
	return errors.New(errBaseTableRequired)
}

func newJoinsRequiredError() error {
	return errors.New(errJoinsRequired)
}

func newTransactionFunctionRequiredError() error {
	return errors.New(msgTransactionFunctionRequired)
}

func newTransactionPanicError(r interface{}) error {
	return fmt.Errorf(msgTransactionPanic, r)
}

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
