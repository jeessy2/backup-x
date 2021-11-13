package web

import (
	"backup-x/client"
)

// Run run
func Run() {
	client.RunCycle()
}

// RunOnce run
func RunOnce() {
	client.RunOnce()
}
