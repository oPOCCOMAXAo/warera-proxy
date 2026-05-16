package runtimeutils

import (
	"strconv"
	"time"
)

//nolint:gochecknoglobals
var startTime = time.Now().Unix()

func GetStartTime() int64 {
	return startTime
}

func GetStartTimeString() string {
	return strconv.FormatInt(startTime, 10)
}
