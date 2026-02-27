package cron

import (
	"fmt"
	"strconv"
	"strings"
)

// CronDescriptor 7位Cron表达式描述器
type CronDescriptor struct {
	seconds string // 秒
	minutes string // 分
	hours   string // 时
	days    string // 日
	months  string // 月
	weeks   string // 周
	years   string // 年
}

// monthNames 月份中文名称
var monthNames = map[int]string{
	1: "1月", 2: "2月", 3: "3月", 4: "4月",
	5: "5月", 6: "6月", 7: "7月", 8: "8月",
	9: "9月", 10: "10月", 11: "11月", 12: "12月",
}

// weekNames 星期中文名称
var weekNames = map[int]string{
	0: "周日", 1: "周一", 2: "周二", 3: "周三",
	4: "周四", 5: "周五", 6: "周六", 7: "周日",
}

// ParseCron 解析cron表达式
// 格式：秒 分 时 日 月 周 年
func ParseCron(cronExpr string) (*CronDescriptor, error) {
	parts := strings.Fields(cronExpr)
	if len(parts) == 5 {
		return &CronDescriptor{
			seconds: "0",
			minutes: parts[0],
			hours:   parts[1],
			days:    parts[2],
			months:  parts[3],
			weeks:   parts[4],
			years:   "",
		}, nil
	}
	if len(parts) == 6 {
		second := parts[0]
		if parts[0] == "?" {
			second = "0"
		}
		return &CronDescriptor{
			seconds: second,
			minutes: parts[1],
			hours:   parts[2],
			days:    parts[3],
			months:  parts[4],
			weeks:   parts[5],
			years:   "*",
		}, nil
	}
	if len(parts) == 7 {
		second := parts[0]
		if parts[0] == "?" {
			second = "0"
		}
		return &CronDescriptor{
			seconds: second,
			minutes: parts[1],
			hours:   parts[2],
			days:    parts[3],
			months:  parts[4],
			weeks:   parts[5],
			years:   parts[6],
		}, nil
	}
	return nil, fmt.Errorf("无效的cron表达式，最多7个字段（秒 分 时 日 月 周 年），最少5个字段（分 时 日 月 周），当前有%d个字段", len(parts))
}

// ToChineseDescription 将cron表达式转换为中文描述
func (cd *CronDescriptor) ToChineseDescription() string {
	var parts []string

	// 处理年份
	yearDesc := cd.describeYear()
	hasYearLimit := cd.years != "*"
	if yearDesc != "" {
		parts = append(parts, yearDesc)
	}

	// 检查是否有日期/星期限制
	hasDayLimit := (cd.days != "*" && cd.days != "?")
	hasWeekLimit := (cd.weeks != "*" && cd.weeks != "?")
	hasMonthLimit := cd.months != "*"

	// 处理月份
	monthDesc := cd.describeMonth()
	if monthDesc != "" && monthDesc != "每月" {
		parts = append(parts, monthDesc)
	}

	// 处理星期（如果指定了具体星期，则日期通常为*）
	weekDesc := cd.describeWeek()

	// 处理日期
	dayDesc := cd.describeDay()

	// 星期和日期的处理
	if weekDesc != "" && dayDesc != "" {
		parts = append(parts, fmt.Sprintf("%s或%s", dayDesc, weekDesc))
	} else if weekDesc != "" {
		// 只有星期限制
		// 对于单个星期添加"每"前缀，对于范围或列表不添加
		if !strings.HasPrefix(weekDesc, "每") && !strings.Contains(weekDesc, "至") && !strings.Contains(weekDesc, "、") {
			// 去掉"周"字，因为weekDesc已经包含了（如"周一"）
			parts = append(parts, "每"+weekDesc)
		} else {
			parts = append(parts, weekDesc)
		}
	} else if dayDesc != "" {
		// 只有日期限制
		// 如果有月份限制，直接添加日期，不需要"每月"前缀
		if hasMonthLimit {
			parts = append(parts, dayDesc)
		} else {
			// 没有月份限制，对于单个日期添加"每月"前缀
			if !strings.HasPrefix(dayDesc, "每") {
				parts = append(parts, "每月"+dayDesc)
			} else {
				parts = append(parts, dayDesc)
			}
		}
	} else if !hasDayLimit && !hasWeekLimit && (hasMonthLimit || hasYearLimit) {
		// 有月份或年份限制但没有日期和星期限制，需要添加"每天"作为上下文
		parts = append(parts, "每天")
	} else if !hasDayLimit && !hasWeekLimit && !hasMonthLimit && !hasYearLimit {
		// 都没有限制，根据时间来判断
		// 特殊情况：不需要添加"每天"的情况
		isEveryHour := cd.seconds == "0" && cd.minutes == "0" && cd.hours == "*" && !strings.Contains(cd.hours, "/")
		isEveryMinute := cd.seconds == "0" && cd.minutes == "*" && cd.hours == "*"
		isEverySecond := cd.seconds == "*" && cd.minutes == "*" && cd.hours == "*"
		hasSecondStep := strings.Contains(cd.seconds, "/")
		hasMinuteStep := strings.Contains(cd.minutes, "/")
		hasHourStep := strings.Contains(cd.hours, "/")

		// 对于有具体时间值的才添加"每天"
		timeHasSpecificValue := (cd.hours != "*" && !hasHourStep) ||
			(cd.minutes != "*" && cd.minutes != "0" && !hasMinuteStep) ||
			(cd.seconds != "*" && cd.seconds != "0" && !hasSecondStep)

		// 对于步长表达式也添加"每天"作为上下文，除了特殊情况
		needDayContext := (hasSecondStep || hasMinuteStep || hasHourStep) && !isEveryHour && !isEveryMinute && !isEverySecond

		if timeHasSpecificValue || needDayContext {
			parts = append(parts, "每天")
		}
	}

	// 处理时分秒
	timeDesc := cd.describeTime()
	if timeDesc != "" {
		parts = append(parts, timeDesc)
	}

	if len(parts) == 0 {
		return "每秒执行"
	}

	return strings.Join(parts, "")
}

// describeYear 描述年份
func (cd *CronDescriptor) describeYear() string {
	if cd.years == "*" || cd.years == "" {
		return ""
	}

	// 范围：2024-2026
	if strings.Contains(cd.years, "-") {
		parts := strings.Split(cd.years, "-")
		if len(parts) == 2 {
			return fmt.Sprintf("在%s年至%s年期间，", parts[0], parts[1])
		}
	}

	// 列表：2024,2025,2026
	if strings.Contains(cd.years, ",") {
		years := strings.Split(cd.years, ",")
		return fmt.Sprintf("在%s年，", strings.Join(years, "、"))
	}

	// 步长：2024/2
	if strings.Contains(cd.years, "/") {
		parts := strings.Split(cd.years, "/")
		if len(parts) == 2 {
			return fmt.Sprintf("从%s年开始每%s年，", parts[0], parts[1])
		}
	}

	// 单个年份
	return fmt.Sprintf("在%s年，", cd.years)
}

// describeMonth 描述月份
func (cd *CronDescriptor) describeMonth() string {
	if cd.months == "*" {
		return "每月"
	}

	// 步长：*/2 或 1/2
	if strings.Contains(cd.months, "/") {
		parts := strings.Split(cd.months, "/")
		if len(parts) == 2 {
			if parts[0] == "*" {
				return fmt.Sprintf("每%s个月", parts[1])
			}
			return fmt.Sprintf("从%s月开始每%s个月", parts[0], parts[1])
		}
	}

	// 范围：1-6
	if strings.Contains(cd.months, "-") {
		parts := strings.Split(cd.months, "-")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			return fmt.Sprintf("%s至%s", monthNames[start], monthNames[end])
		}
	}

	// 列表：1,3,5
	if strings.Contains(cd.months, ",") {
		months := strings.Split(cd.months, ",")
		var names []string
		for _, m := range months {
			if num, err := strconv.Atoi(strings.TrimSpace(m)); err == nil {
				if name, ok := monthNames[num]; ok {
					names = append(names, name)
				}
			}
		}
		if len(names) > 0 {
			return strings.Join(names, "、")
		}
	}

	// 单个月份
	if num, err := strconv.Atoi(cd.months); err == nil {
		if name, ok := monthNames[num]; ok {
			return name
		}
	}

	return "每月"
}

// describeDay 描述日期
func (cd *CronDescriptor) describeDay() string {
	if cd.days == "*" || cd.days == "?" {
		return ""
	}

	// 步长：*/5
	if strings.Contains(cd.days, "/") {
		parts := strings.Split(cd.days, "/")
		if len(parts) == 2 {
			if parts[0] == "*" {
				return fmt.Sprintf("每%s天", parts[1])
			}
			return fmt.Sprintf("从%s号开始每%s天", parts[0], parts[1])
		}
	}

	// 范围：1-15
	if strings.Contains(cd.days, "-") {
		parts := strings.Split(cd.days, "-")
		if len(parts) == 2 {
			return fmt.Sprintf("%s号至%s号", parts[0], parts[1])
		}
	}

	// 列表：1,10,20
	if strings.Contains(cd.days, ",") {
		days := strings.Split(cd.days, ",")
		var formattedDays []string
		for _, d := range days {
			formattedDays = append(formattedDays, strings.TrimSpace(d)+"号")
		}
		return strings.Join(formattedDays, "、")
	}

	// 单个日期
	return fmt.Sprintf("%s号", cd.days)
}

// describeWeek 描述星期
func (cd *CronDescriptor) describeWeek() string {
	if cd.weeks == "*" || cd.weeks == "?" {
		return ""
	}

	// 范围：1-5 (周一到周五)
	if strings.Contains(cd.weeks, "-") {
		parts := strings.Split(cd.weeks, "-")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			return fmt.Sprintf("%s至%s", weekNames[start], weekNames[end])
		}
	}

	// 列表：1,3,5
	if strings.Contains(cd.weeks, ",") {
		weeks := strings.Split(cd.weeks, ",")
		var names []string
		for _, w := range weeks {
			if num, err := strconv.Atoi(strings.TrimSpace(w)); err == nil {
				if name, ok := weekNames[num]; ok {
					names = append(names, name)
				}
			}
		}
		if len(names) > 0 {
			return strings.Join(names, "、")
		}
	}

	// 步长：*/2
	if strings.Contains(cd.weeks, "/") {
		parts := strings.Split(cd.weeks, "/")
		if len(parts) == 2 {
			return fmt.Sprintf("每%s天", parts[1])
		}
	}

	// 单个星期
	if num, err := strconv.Atoi(cd.weeks); err == nil {
		if name, ok := weekNames[num]; ok {
			return name
		}
	}

	return ""
}

// describeTime 描述时分秒
func (cd *CronDescriptor) describeTime() string {
	// 判断最小的时间单位
	// 优先级：秒 > 分 > 时

	// 情况1: 秒有步长或特定值（不是 * 或有 /）
	if cd.seconds != "*" {
		if strings.Contains(cd.seconds, "/") {
			// 如 */10 表示每10秒
			return cd.describeSecond()
		} else if cd.seconds != "0" {
			// 有具体的秒数
			return cd.buildCompleteTime()
		}
	}

	// 情况2: 秒是0或*，检查分钟
	if cd.minutes != "*" {
		if strings.Contains(cd.minutes, "/") {
			// 如 */5 表示每5分钟
			parts := strings.Split(cd.minutes, "/")
			if len(parts) == 2 && parts[0] == "*" {
				return fmt.Sprintf("每%s分钟", parts[1])
			}
			return cd.describeMinute()
		} else {
			// 有具体的分钟
			return cd.buildCompleteTime()
		}
	}

	// 情况3: 分钟也是*，检查小时
	if cd.hours != "*" {
		if strings.Contains(cd.hours, "/") {
			// 如 */4 表示每4小时
			parts := strings.Split(cd.hours, "/")
			if len(parts) == 2 && parts[0] == "*" {
				return fmt.Sprintf("每%s小时", parts[1])
			}
			return cd.describeHour()
		} else {
			// 有具体的小时
			return cd.buildCompleteTime()
		}
	}

	// 情况4: 时分秒都是*或者只有秒是0其他是*
	// 0 * * 表示每分钟
	if cd.seconds == "0" && cd.minutes == "*" && cd.hours == "*" {
		return "每分钟"
	}

	// 0 0 * 表示每小时整点
	if cd.seconds == "0" && cd.minutes == "0" && cd.hours == "*" {
		return "每小时整点"
	}

	// * * * 表示每秒
	if cd.seconds == "*" && cd.minutes == "*" && cd.hours == "*" {
		return "每秒执行"
	}

	return cd.buildCompleteTime()
}

// buildCompleteTime 构建完整的时分秒描述
func (cd *CronDescriptor) buildCompleteTime() string {
	var parts []string

	// 小时部分
	if cd.hours != "*" {
		hourDesc := cd.describeHour()
		if hourDesc != "" {
			parts = append(parts, hourDesc)
		}
	} else if cd.minutes != "*" {
		// 小时是*但分钟不是*，表示每小时的某分钟
		parts = append(parts, "每小时")
	}

	// 分钟部分
	if cd.minutes != "*" {
		minuteDesc := cd.describeMinute()
		if minuteDesc != "" {
			// 如果前面有"每小时"，添加"的"字（但整点不需要"的"）
			if len(parts) > 0 && parts[len(parts)-1] == "每小时" {
				if minuteDesc == "整点" {
					parts = append(parts, minuteDesc)
				} else {
					parts = append(parts, "的"+minuteDesc)
				}
			} else {
				parts = append(parts, minuteDesc)
			}
		}
	} else if len(parts) > 0 && cd.hours != "*" {
		// 如果有小时但分钟是*，说明是整点
		parts = append(parts, "整点")
	}

	// 秒部分
	if cd.seconds != "*" && cd.seconds != "0" {
		secondDesc := cd.describeSecond()
		if secondDesc != "" {
			parts = append(parts, secondDesc)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "")
}

// describeHour 描述小时
func (cd *CronDescriptor) describeHour() string {
	if cd.hours == "*" {
		return ""
	}

	// 步长：*/4
	if strings.Contains(cd.hours, "/") {
		parts := strings.Split(cd.hours, "/")
		if len(parts) == 2 {
			if parts[0] == "*" {
				return fmt.Sprintf("每%s小时", parts[1])
			}
			return fmt.Sprintf("从%s点开始每%s小时", parts[0], parts[1])
		}
	}

	// 范围：9-17
	if strings.Contains(cd.hours, "-") {
		parts := strings.Split(cd.hours, "-")
		if len(parts) == 2 {
			return fmt.Sprintf("%s点至%s点", parts[0], parts[1])
		}
	}

	// 列表：9,12,18
	if strings.Contains(cd.hours, ",") {
		hours := strings.Split(cd.hours, ",")
		var formattedHours []string
		for _, h := range hours {
			formattedHours = append(formattedHours, strings.TrimSpace(h)+"点")
		}
		return strings.Join(formattedHours, "、")
	}

	// 单个小时
	return fmt.Sprintf("%s点", cd.hours)
}

// describeMinute 描述分钟
func (cd *CronDescriptor) describeMinute() string {
	if cd.minutes == "*" {
		return ""
	}

	// 步长：*/10
	if strings.Contains(cd.minutes, "/") {
		parts := strings.Split(cd.minutes, "/")
		if len(parts) == 2 {
			if parts[0] == "*" {
				return fmt.Sprintf("每%s分钟", parts[1])
			}
			return fmt.Sprintf("从%s分开始每%s分钟", parts[0], parts[1])
		}
	}

	// 范围：0-30
	if strings.Contains(cd.minutes, "-") {
		parts := strings.Split(cd.minutes, "-")
		if len(parts) == 2 {
			return fmt.Sprintf("%s分至%s分", parts[0], parts[1])
		}
	}

	// 列表：0,15,30,45
	if strings.Contains(cd.minutes, ",") {
		minutes := strings.Split(cd.minutes, ",")
		var formattedMinutes []string
		for _, m := range minutes {
			formattedMinutes = append(formattedMinutes, strings.TrimSpace(m)+"分")
		}
		return strings.Join(formattedMinutes, "、")
	}

	// 单个分钟
	if cd.minutes == "0" {
		return "整点"
	}
	return fmt.Sprintf("%s分", cd.minutes)
}

// describeSecond 描述秒
func (cd *CronDescriptor) describeSecond() string {
	if cd.seconds == "*" {
		return ""
	}

	// 步长：*/10
	if strings.Contains(cd.seconds, "/") {
		parts := strings.Split(cd.seconds, "/")
		if len(parts) == 2 {
			if parts[0] == "*" {
				return fmt.Sprintf("每%s秒", parts[1])
			}
			return fmt.Sprintf("从%s秒开始每%s秒", parts[0], parts[1])
		}
	}

	// 范围：0-30
	if strings.Contains(cd.seconds, "-") {
		parts := strings.Split(cd.seconds, "-")
		if len(parts) == 2 {
			return fmt.Sprintf("%s秒至%s秒", parts[0], parts[1])
		}
	}

	// 列表：0,15,30,45
	if strings.Contains(cd.seconds, ",") {
		seconds := strings.Split(cd.seconds, ",")
		var formattedSeconds []string
		for _, s := range seconds {
			formattedSeconds = append(formattedSeconds, strings.TrimSpace(s)+"秒")
		}
		return strings.Join(formattedSeconds, "、")
	}

	// 单个秒
	if cd.seconds == "0" {
		return ""
	}
	return fmt.Sprintf("%s秒", cd.seconds)
}

// CronToDescription 将7位cron表达式转换为中文描述（快捷函数）
func CronToDescription(cronExpr string) (string, error) {
	descriptor, err := ParseCron(cronExpr)
	if err != nil {
		return "", err
	}
	return descriptor.ToChineseDescription(), nil
}
