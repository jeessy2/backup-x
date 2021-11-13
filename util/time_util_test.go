package util

import (
	"testing"
	"time"
)

// TestGetDelaySeconds
func TestGetDelaySeconds(t *testing.T) {
	hour := 1
	runTime := time.Now().Add(GetDelaySeconds(hour))
	if runTime.Hour() != hour || runTime.Minute() != 0 {
		t.Error("GetDelaySeconds not correct")
	}
}
