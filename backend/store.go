package main

import (
	"time"
)

const (
	STATE_INIT = iota
	STATE_READY
	STATE_UPDATING
	STATE_ERROR
)

type StoreStatus struct {
	LastRefresh time.Time
	LastError   error
	State       int
}

// Helper: stateToString
func stateToString(state int) string {
	switch state {
	case STATE_INIT:
		return "INIT"
	case STATE_READY:
		return "READY"
	case STATE_UPDATING:
		return "UPDATING"
	case STATE_ERROR:
		return "ERROR"
	}
	return "INVALID"
}

type updateStatus string

const(
	Success updateStatus = "Success"
	Failure = "Failure"
	Skipped = "Skipped"
)
