package main

import "time"

//GetDateRange 获取默认合法作业时间范围
func GetDateRange() (startDate time.Time, endDate time.Time) {
	nowDate := time.Now()

	if nowDate.Hour() < 12 {
		startDate = time.Date(nowDate.Year(), nowDate.Month(), nowDate.Day()-1,
			12, 0, 0, 0, nowDate.Location())
		endDate = time.Date(nowDate.Year(), nowDate.Month(), nowDate.Day(),
			7, 30, 0, 0, nowDate.Location())
	} else if nowDate.Hour() > 12 && nowDate.Hour() < 17 {
		startDate = time.Date(nowDate.Year(), nowDate.Month(), nowDate.Day()-1,
			17, 0, 0, 0, nowDate.Location())
		endDate = time.Date(nowDate.Year(), nowDate.Month(), nowDate.Day(),
			12, 30, 0, 0, nowDate.Location())
	} else {
		startDate = time.Date(nowDate.Year(), nowDate.Month(), nowDate.Day()-1,
			21, 30, 0, 0, nowDate.Location())
		endDate = time.Date(nowDate.Year(), nowDate.Month(), nowDate.Day(),
			17, 30, 0, 0, nowDate.Location())
	}
	return
}
