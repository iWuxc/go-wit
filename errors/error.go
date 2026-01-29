package errors

import (
	"fmt"
	"github.com/iWuxc/go-wit/metrics/stat"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type ErrorInterface interface {
	error
	WithMetadata(map[string]string) ErrorInterface
	WithMessage(string) ErrorInterface
	WithLocation(level string) ErrorInterface
	WithLevel(level string) ErrorInterface
	GinError(ctx *gin.Context)
}

const (
	// UnknownReason is unknown reason for error info.
	UnknownReason = "未知的错误"
	// SupportPackageIsVersion1 this constant should not be referenced by any other code.
	SupportPackageIsVersion1 = true
)

var _ ErrorInterface = (*Error)(nil)

type errKey string

var errs = map[errKey]*Error{}

// Register 注册错误信息
func Register(egoError *Error) {
	errs[errKey(egoError.Reason)] = egoError
}

func (x *Error) Error() string {
	if x.GetLevel() == "debug" {
		return x.details()
	}
	
	return x.GetMessage()
}

func (x *Error) details() string {
	return fmt.Sprintf("msg: %s, code: %d, location: %s, level: %s, metadata: %v", x.GetMessage(), x.GetCode(), x.GetLocation(), x.GetLevel(), x.GetMetadata())
}

// Is 判断是否为根因错误
func (x *Error) Is(err error) bool {
	egoErr, flag := err.(*Error)
	if !flag {
		return false
	}
	return x.GetCode() == egoErr.GetCode()
}

// GRPCStatus returns the Status represented by se.
func (x *Error) GRPCStatus() *status.Status {
	s, _ := status.New(codes.Code(x.Code), x.Message).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   x.GetReason(),
			Metadata: x.GetMetadata(),
		})
	return s
}

// WithMetadata with an MD formed by the mapping of key, value.
func (x *Error) WithMetadata(md map[string]string) ErrorInterface {
	err := proto.Clone(x).(*Error)
	err.Metadata = md
	return err
}

// WithMessage set message to current Error
func (x *Error) WithMessage(msg string) ErrorInterface {
	err := proto.Clone(x).(*Error)
	err.Message = msg
	return err
}

// WithLevel set error level to current Error
func (x *Error) WithLevel(level string) ErrorInterface {
	err := proto.Clone(x).(*Error)
	err.Level = level
	return err
}

// WithLocation set error location to current Error
func (x *Error) WithLocation(location string) ErrorInterface {
	err := proto.Clone(x).(*Error)
	err.Location = location
	return err
}

func (x *Error) GinError(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": x.GetCode(),
		"msg":  x.GetMessage(),
		"data": x.GetMetadata(),
	})
}

// NewError returns an error object for the code, message.
func NewError(code int32, reason, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Reason:  reason,
	}
}

// NewStatError return a stat error object for the location, msg, level
func NewStatError(location, msg, level string) *Error {
	stat.APPErrorCount.With("location", location, "level", level, "info", msg).Inc()
	return &Error{
		Location: location,
		Message:  msg,
		Level:    level,
	}
}

// FromError try to convert an error to *Error.
// It supports wrapped errors.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if ok {
		for _, detail := range gs.Details() {
			switch d := detail.(type) {
			case *errdetails.ErrorInfo:
				e, ok := errs[errKey(d.Reason)]
				if ok {
					return e.WithMessage(gs.Message()).WithMetadata(d.Metadata).(*Error)
				}
				return NewError(
					int32(gs.Code()),
					d.Reason,
					gs.Message(),
				).WithMetadata(d.Metadata).(*Error)
			}
		}
	}
	return NewError(int32(codes.Unknown), UnknownReason, err.Error())
}
