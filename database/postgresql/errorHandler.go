package postgresql

import (
	"context"
	"errors"

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
	msgDuplicateRecord     = "a record with the same values already exists"
	msgConstraintViolation = "the provided data violates database constraints"
	msgInvalidReference    = "invalid reference in the provided data"
	msgDatabaseError       = "an unexpected database error occurred"
)

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
