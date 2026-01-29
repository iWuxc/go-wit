package errors

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Convert(err error) *status.Status {
	if err == nil {
		return status.New(codes.OK, "OK")
	}

	if se, ok := err.(interface {
		GRPCStatus() *status.Status
	}); ok {
		return se.GRPCStatus()
	}

	switch err {
	case context.DeadlineExceeded:
		return status.New(codes.DeadlineExceeded, err.Error())
	case context.Canceled:
		return status.New(codes.Canceled, err.Error())
	}

	return status.New(codes.Unknown, err.Error())
}

// New .
// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
//
// @Doc https://pkg.go.dev/github.com/pkg/errors#New
func New(text string) error { return errors.New(text) }

// Errorf
// formats according to a format specifier and returns the string as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
//
// @Doc https://pkg.go.dev/github.com/pkg/errors#Errorf
func Errorf(format string, args ...interface{}) error { return errors.Errorf(format, args...) }

// Is reports whether any error in error's chain matches target.
// The chain consists of err itself followed by the sequence of errors obtained by repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or
// if it implements a method Is(error) bool such that Is(target) returns true.
//
// @Doc https://pkg.go.dev/github.com/pkg/errors#Is
func Is(err, target error) bool { return errors.Is(err, target) }

// As finds the first error in error's chain that matches target, and if so, sets target to that error value and returns true.
//
// The chain consists of err itself followed by the sequence of errors obtained by repeatedly calling Unwrap.
// An error matches target if the error's concrete value is assignable to the value pointed to by target, or if the error has a method As(interface{}) bool such that As(target) returns true. In the latter case, the As method is responsible for setting target.
// As will panic if target is not a non-nil pointer to either a type that implements error, or to any interface type. As returns false if err is nil.
//
// @Doc https://pkg.go.dev/github.com/pkg/errors#As
func As(err error, target interface{}) bool { return errors.As(err, target) }

// Unwrap returns the result of calling the Unwrap method on err,
// if error's type contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
//
// @Doc https://pkg.go.dev/github.com/pkg/errors#Unwrap
func Unwrap(err error) error { return errors.Unwrap(err) }

// Wrap . returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error { return errors.Wrap(err, message) }

// WithMessage annotates err with the format specifier.
// If err is nil, WithMessage returns nil.
func WithMessage(err error, message string) error { return errors.WithMessage(err, message) }
