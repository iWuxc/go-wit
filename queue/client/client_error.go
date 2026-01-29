package client

import "github.com/iWuxc/go-wit/errors"

// ErrDuplicateTask indicates that the given task could not be enqueued since it's a duplicate of another task.
//
// ErrDuplicateTask error only applies to task enqueued with a Unique option.
var ErrDuplicateTask = errors.New("task already exists")

// ErrTaskIDConflict indicates that the given task could not be enqueued since its task ID already exists.
//
// ErrTaskIDConflict error only applies to task enqueued with a TaskID option.
var ErrTaskIDConflict = errors.New("task ID conflicts with another task")