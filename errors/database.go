package errors

import "gorm.io/gorm"

var (
	DBErrRecordNotFound      = gorm.ErrRecordNotFound
	DBErrInvalidTransaction  = gorm.ErrInvalidTransaction
	DBErrMissingWhereClause  = gorm.ErrMissingWhereClause
	DBErrUnsupportedRelation = gorm.ErrUnsupportedRelation
	DBErrPrimaryKeyRequired  = gorm.ErrPrimaryKeyRequired
	DBErrModelValueRequired  = gorm.ErrModelValueRequired
	DBErrUnsupportedDriver   = gorm.ErrUnsupportedDriver
	DBErrRegistered          = gorm.ErrRegistered
	DBErrInvalidField        = gorm.ErrInvalidField
	DBErrInvalidDB           = gorm.ErrInvalidDB
)
