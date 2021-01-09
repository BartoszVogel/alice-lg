package birdwatcherV2

import (
	"strings"
	"sync"
)

/*
Helper functions for dealing with birdwatcher API data
*/
type LockMap struct {
	locks *sync.Map
}

func NewLockMap() *LockMap {
	return &LockMap{
		locks: &sync.Map{},
	}
}

func (l *LockMap) Lock(key string) {
	mutex, _ := l.locks.LoadOrStore(key, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
}

func (l *LockMap) Unlock(key string) {
	mutex, ok := l.locks.Load(key)
	if !ok {
		return // Nothing to unlock
	}
	mutex.(*sync.Mutex).Unlock()
}

func isProtocolUp(protocol string) bool {
	protocol = strings.ToLower(protocol)
	return protocol == "up"
}
