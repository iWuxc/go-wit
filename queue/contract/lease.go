package contract

import (
	"sync"
	"time"
)

// Lease is a time bound lease for worker to process task.
// It provides a communication channel between lessor and lessee about lease expiration.
type Lease struct {
	once     sync.Once
	ch       chan struct{}
	mu       sync.Mutex
	expireAt time.Time // guarded by mu
}

func NewLease(expirationTime time.Time) *Lease {
	return &Lease{
		ch:       make(chan struct{}),
		expireAt: expirationTime,
	}
}

// Reset changes the lease to expire at the given time.
// It returns true if the lease is still valid and reset operation was successful, false if the lease had been expired.
func (l *Lease) Reset(expirationTime time.Time) bool {
	if !l.IsValid() {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireAt = expirationTime
	return true
}

// NotifyExpiration
// Sends a notification to lessee about expired lease
// Returns true if notification was sent, returns false if the lease is still valid and notification was not sent.
func (l *Lease) NotifyExpiration() bool {
	if l.IsValid() {
		return false
	}
	l.once.Do(l.closeCh)
	return true
}

func (l *Lease) closeCh() {
	close(l.ch)
}

// Done returns a communication channel from which the lessee can read to get notified when lessor notifies about lease expiration.
func (l *Lease) Done() <-chan struct{} {
	return l.ch
}

// Deadline returns the expiration time of the lease.
func (l *Lease) Deadline() time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.expireAt
}

// IsValid returns true if the lease's expiration time is in the future or equals to the current time,
// returns false otherwise.
func (l *Lease) IsValid() bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.expireAt.After(now) || l.expireAt.Equal(now)
}
