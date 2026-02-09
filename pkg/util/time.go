package util

import (
	"fmt"
	"github.com/hatcher/common/pkg/logs"
	"strconv"
	"time"
)

func ParseTime(layout, value string) (time.Time, error) {
	return ParseTimeWithTimeZone(layout, value, "Asia/Shanghai")
}

func ParseTimeWithTimeZone(layout, value string, tz string) (time.Time, error) {
	if tz == "" {
		return time.Parse(layout, value)
	}
	// 使用 time.LoadLocation 设置全局时区
	loc, err := time.LoadLocation(tz)
	if err != nil {
		logs.Error("Error loading location", err)
		return time.Parse(layout, value)
	}
	return time.ParseInLocation(layout, value, loc)
}

type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

// AnalysisTimeRange 分析时间范围
func AnalysisTimeRange(current time.Time, timeWindow int64, timeUnit string) TimeRange {
	var actualTimeUnit string
	if timeUnit == "" {
		actualTimeUnit = "min"
	}
	actualTimeUnit = timeUnit
	if timeUnit == "分钟" {
		actualTimeUnit = "minute"
	}
	if timeUnit == "小时" {
		actualTimeUnit = "hour"
	}
	if timeUnit == "天" {
		actualTimeUnit = "day"
	}
	if timeUnit == "周" {
		actualTimeUnit = "week"
	}
	if timeUnit == "月" {
		actualTimeUnit = "month"
	}
	if timeUnit == "年" {
		actualTimeUnit = "year"
	}
	var startTime, endTime time.Time
	endTime = current
	switch actualTimeUnit {
	case "min":
	case "minute":
		startTime = current.Add(time.Duration(-timeWindow) * time.Minute)
	case "hour":
		startTime = current.Add(time.Duration(-timeWindow) * time.Hour)
	case "day":
		startTime = current.Add(time.Duration(-timeWindow) * 24 * time.Hour)
	case "week":
		startTime = current.Add(time.Duration(-timeWindow) * 7 * 24 * time.Hour)
	case "month":
		startTime = current.Add(time.Duration(-timeWindow) * 30 * 24 * time.Hour)
	case "year":
		startTime = current.Add(time.Duration(-timeWindow) * 365 * 24 * time.Hour)
	default:
		startTime = current.Add(time.Duration(-timeWindow) * time.Minute)
	}
	return TimeRange{
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// CompareTimeStrings 比较两个时间字符串的大小
// 返回: -1: t1 < t2, 0: t1 == t2, 1: t1 > t2, error: 解析失败
func CompareTimeStrings(t1Str, t2Str string) (int, error) {
	t1, err := ParseTimeString(t1Str)
	if err != nil {
		return 0, fmt.Errorf("解析第一个时间失败: %v", err)
	}

	t2, err := ParseTimeString(t2Str)
	if err != nil {
		return 0, fmt.Errorf("解析第二个时间失败: %v", err)
	}

	return CompareTimes(t1, t2), nil
}

// ParseTimeString 解析多种格式的时间字符串
func ParseTimeString(timeStr string) (time.Time, error) {
	// 尝试多种常见的时间格式
	formats := []string{
		// RFC3339 格式 (ISO8601)
		time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
		time.RFC3339,     // "2006-01-02T15:04:05Z07:00"

		// 常见格式
		"2006-01-02T15:04:05.999+0800",
		"2006-01-02T15:04:05.999Z", // 你的格式
		"2006-01-02T15:04:05Z",     // 不带毫秒
		"2006-01-02 15:04:05.999",  // 空格分隔
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"15:04:05",

		// 中文格式
		"2006年01月02日 15时04分05秒",
		"2006年01月02日",

		// Unix 时间戳
		"1136239445", // 秒级时间戳
	}

	// 先尝试解析为Unix时间戳
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		// 判断是秒还是毫秒
		if timestamp > 10000000000 { // 大于 2001-09-09 的时间戳可能是毫秒
			return time.Unix(0, timestamp*int64(time.Millisecond)), nil
		}
		return time.Unix(timestamp, 0), nil
	}

	// 尝试所有格式
	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("无法解析时间字符串 '%s': %v", timeStr, lastErr)
}

// CompareTimes 比较两个time.Time对象
func CompareTimes(t1, t2 time.Time) int {
	if t1.Before(t2) {
		return -1
	}
	if t1.After(t2) {
		return 1
	}
	return 0
}

// TimeComparison 时间比较的完整功能
type TimeComparison struct{}

// NewTimeComparison 创建时间比较器
func NewTimeComparison() *TimeComparison {
	return &TimeComparison{}
}

// Compare 比较两个时间
func (tc *TimeComparison) Compare(t1Str, t2Str string) (int, error) {
	return CompareTimeStrings(t1Str, t2Str)
}

// IsBefore 判断t1是否在t2之前
func (tc *TimeComparison) IsBefore(t1Str, t2Str string) (bool, error) {
	result, err := CompareTimeStrings(t1Str, t2Str)
	if err != nil {
		return false, err
	}
	return result == -1, nil
}

// IsAfter 判断t1是否在t2之后
func (tc *TimeComparison) IsAfter(t1Str, t2Str string) (bool, error) {
	result, err := CompareTimeStrings(t1Str, t2Str)
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// IsEqual 判断两个时间是否相等
func (tc *TimeComparison) IsEqual(t1Str, t2Str string) (bool, error) {
	result, err := CompareTimeStrings(t1Str, t2Str)
	if err != nil {
		return false, err
	}
	return result == 0, nil
}

// IsBetween 判断时间是否在某个范围内 [start, end]
func (tc *TimeComparison) IsBetween(checkStr, startStr, endStr string) (bool, error) {
	checkTime, err := ParseTimeString(checkStr)
	if err != nil {
		return false, err
	}

	startTime, err := ParseTimeString(startStr)
	if err != nil {
		return false, err
	}

	endTime, err := ParseTimeString(endStr)
	if err != nil {
		return false, err
	}

	return (checkTime.Equal(startTime) || checkTime.After(startTime)) &&
		(checkTime.Equal(endTime) || checkTime.Before(endTime)), nil
}

// GetTimeDifference 获取两个时间的差值
func (tc *TimeComparison) GetTimeDifference(t1Str, t2Str string) (time.Duration, error) {
	t1, err := ParseTimeString(t1Str)
	if err != nil {
		return 0, err
	}

	t2, err := ParseTimeString(t2Str)
	if err != nil {
		return 0, err
	}

	return t2.Sub(t1), nil
}

// FormatTimeDifference 格式化时间差为可读字符串
func (tc *TimeComparison) FormatTimeDifference(t1Str, t2Str string) (string, error) {
	duration, err := tc.GetTimeDifference(t1Str, t2Str)
	if err != nil {
		return "", err
	}

	return FormatDuration(duration), nil
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	// 取绝对值
	if d < 0 {
		d = -d
	}

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second
	d -= seconds * time.Second

	milliseconds := d / time.Millisecond

	var result string
	if days > 0 {
		result += fmt.Sprintf("%d天", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%d小时", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%d分钟", minutes)
	}
	if seconds > 0 {
		result += fmt.Sprintf("%d秒", seconds)
	}
	if milliseconds > 0 && days == 0 && hours == 0 && minutes == 0 {
		result += fmt.Sprintf("%d毫秒", milliseconds)
	}

	if result == "" {
		return "0秒"
	}
	return result
}

// GetDaysInMonth 获取指定月份的天数
func GetDaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// GetAllDaysInCurrentMonth 获取当前月的所有日期
func GetAllDaysInCurrentMonth() []time.Time {
	now := time.Now()
	year, month, _ := now.Date()
	return GetAllDaysByMonth(year, month)
}

// GetAllDaysByMonth 获取指定月份的所有日期
func GetAllDaysByMonth(year int, month time.Month) []time.Time {
	totalDays := GetDaysInMonth(year, month)
	var dates []time.Time
	for day := 1; day <= totalDays; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		dates = append(dates, date)
	}
	return dates
}

// GetBeginTime 获取开始时间
func GetBeginTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

// GetEndTime 获取结束时间
func GetEndTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999, date.Location())
}

// NowTime 获取当前时间，转换时区
func NowTime() time.Time {
	now := time.Now()
	parseTime, _ := ParseTime("2006-01-02 15:04:05", now.Format("2006-01-02 15:04:05"))
	return parseTime
}
