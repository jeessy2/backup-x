package util

import (
	"time"
)

// GetDelaySeconds 获取第一次启动的延时时间(秒)
func GetDelaySeconds(startTime int) time.Duration {
	now := time.Now().Truncate(time.Second)
	midNightNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	midNightTom := midNightNow.Add(24 * time.Hour)

	var seconds int
	if now.Hour() >= startTime {
		// tomorrow
		seconds = int(midNightTom.Add(time.Hour * time.Duration(startTime)).Sub(now).Seconds())
	} else {
		seconds = int(midNightNow.Add(time.Hour * time.Duration(startTime)).Sub(now).Seconds())
	}

	return time.Second * time.Duration(seconds)
}
