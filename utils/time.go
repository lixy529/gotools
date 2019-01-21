package utils

import (
	"time"
)

// CurTime returns the current time.
// fmtStr: time format, default is yyyy-mm-dd hh:mi:ss.
func CurTime(fmtStr ...string) string {
	str := "2006-01-02 15:04:05"
	if len(fmtStr) > 0 {
		str = fmtStr[0]
	}

	return time.Now().Format(str)
}

// AddDate returns the current time to increase or decrease the specified time.
func AddDate(years, months, days int) time.Time {
	nTime := time.Now()
	return nTime.AddDate(years, months, days)
}

// StrToTimeStamp returns timestamp.
// strTime: The time in string format.
// fmtTime: time format, default is 2006-01-02 15:04:05.
func StrToTimeStamp(timeStr string, timeFmt ...string) int64 {
	tmpFmt := "2006-01-02 15:04:05"
	if len(timeFmt) > 0 {
		tmpFmt = timeFmt[0]
	}
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(tmpFmt, timeStr, loc)
	return theTime.Unix()
}

// TimeStampToStr returns the time in string format.
// args:
// 1) timestamp: If timestamp is empty, take the current time.
// 2) time format, default is 2006-01-02 15:04:05.
func TimeStampToStr(args ...interface{}) string {
	argc := len(args)
	var timeStamp int64
	var ok bool
	timeFmt := "2006-01-02 15:04:05"
	if argc == 0 {
		return time.Now().Format(timeFmt)
	}
	if argc > 0 {
		timeStamp, ok = args[0].(int64)
		if !ok {
			return ""
		}
	}
	if argc > 1 {
		timeFmt, ok = args[1].(string)
		if !ok {
			return ""
		}
	}

	return time.Unix(timeStamp, 0).Format(timeFmt)
}

// TomrrowRest returns the remaining time from now to tomorrow morning, the unit is seconds
func TomrrowRest() int64 {
	tom := AddDate(0, 0, 1)
	tomStr, _ := time.ParseInLocation("2006-01-02", tom.Format("2006-01-02"), time.Local)

	return tomStr.Unix() - time.Now().Unix()
}
