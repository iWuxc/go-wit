package utils

import (
	"strconv"
	"time"
)

func GetCurrentTime() string {
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02 15:04:05")
}

func GetCurrentUnixTime() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

type DateTime time.Time

func (d DateTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(d).Format("2006-01-02 15:04:05")), nil
}
