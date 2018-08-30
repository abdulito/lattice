package sync

import (
	"fmt"
	"sync"
)

// LockGranularity defines the granularity with which to obtain a lock.
type LockGranularity int32

const (
	// LockGranularityIntentionExclusive can be acquired if no
	// other goroutine has acquired the lock with LockGranularityExclusive.
	LockGranularityIntentionExclusive LockGranularity = iota

	// LockGranularityExclusive can be acquired if no other goroutine
	// has acquired the lock with LockGranularityExclusive or LockGranularityIntentionExclusive.
	LockGranularityExclusive
)

// IntentionLock is a non-blocking multi-granular lock whose acquisition semantics
// respect the following matrix:
//      IX  X
// IX | Y | N
//  X | N | N
type IntentionLock struct {
	lock sync.Mutex

	exclusive bool
	intention int32
}

// IntentionLockUnlocker is returned by a successful call to TryLock, and is used
// to later unlock the IntentionLock with the proper granularity.
type IntentionLockUnlocker struct {
	lock        *IntentionLock
	granularity LockGranularity
}

// TryLock will attempt to lock the IntentionLock with the supplied LockGranularity
// without blocking. If it succeeds it returns an IntentionLockUnlocker which must
// later be used to unlock the lock. If it fails it returns nil.
func (l *IntentionLock) TryLock(granularity LockGranularity) *IntentionLockUnlocker {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch granularity {
	case LockGranularityIntentionExclusive:
		if l.exclusive {
			return nil
		}

		l.intention += 1

	case LockGranularityExclusive:
		if l.exclusive || l.intention > 0 {
			return nil
		}

		l.exclusive = true

	default:
		panic(fmt.Sprintf("unrecognized lock granularity: %v", granularity))
	}

	return &IntentionLockUnlocker{
		lock:        l,
		granularity: granularity,
	}
}

// Unlock will unlock the corresponding IntentionLock for the granularity associated with
// the IntentionLockUnlocker.
func (l *IntentionLockUnlocker) Unlock() {
	l.lock.lock.Lock()
	defer l.lock.lock.Unlock()

	switch l.granularity {
	case LockGranularityIntentionExclusive:
		l.lock.intention -= 1

	case LockGranularityExclusive:
		l.lock.exclusive = false

	default:
		panic(fmt.Sprintf("unrecognized lock granularity: %v", l.granularity))
	}
}
