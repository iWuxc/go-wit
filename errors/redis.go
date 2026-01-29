package errors

import (
	"fmt"
	"strings"
)

// RedisError indicates that the given redis command returned error.
type RedisError struct {
	Command string // redis command (e.g. LRANGE, ZADD, etc)
	Err     error  // underlying error
}

func (e *RedisError) Error() string {
	return fmt.Sprintf("redis command error: %s failed: %v", strings.ToUpper(e.Command), e.Err)
}

func (e *RedisError) Unwrap() error { return e.Err }

// IsRedisError reports whether any error in error's chain is of type RedisError.
func IsRedisError(err error) bool {
	var target *RedisError
	return As(err, &target)
}
