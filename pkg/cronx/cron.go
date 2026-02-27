package cron

import (
	"github.com/robfig/cron/v3"
	"strconv"
	"strings"
	"time"
)

type CronParser struct {
	parser cron.Parser
}

var DefaultCronParser = NewCronParser()

func NewCronParser() *CronParser {
	return &CronParser{
		parser: cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
	}
}

// ParseExpression 解析cron表达式
func (cp *CronParser) ParseExpression(expr string) (cron.Schedule, string, bool, string, error) {
	parts := strings.Fields(expr)
	if len(parts) == 7 {
		// 前6位是标准cron表达式
		actualExp := strings.Join(parts[:6], " ")
		// 第7位是年份
		yearRange := parts[6]
		schedule, err := cp.parser.Parse(actualExp)
		return schedule, actualExp, true, yearRange, err
	} else {
		schedule, err := cp.parser.Parse(expr)
		return schedule, expr, false, "", err
	}
}

// ValidateExpression 验证cron表达式是否有效
func (cp *CronParser) ValidateExpression(expr string) (bool, string, string, error) {
	_, actualExp, hasYear, year, err := cp.ParseExpression(expr)
	return hasYear, actualExp, year, err
}

// GetNextNSchedules 获取接下来N次执行时间
func (cp *CronParser) GetNextNSchedules(expr string, from time.Time, n int) ([]time.Time, error) {
	schedule, _, hasYear, year, err := cp.ParseExpression(expr)
	if err != nil {
		return nil, err
	}
	// 先判断是否满足年度
	yearInt := time.Now().Year()
	if hasYear {
		if !MatchYear(year, yearInt) {
			return nil, nil
		}
	}

	var schedules []time.Time
	current := from
	for i := 0; i < n; i++ {
		next := schedule.Next(current)
		schedules = append(schedules, next)
		current = next
	}

	return schedules, nil
}

// CronExpressionInfo cron表达式信息
type CronExpressionInfo struct {
	Expression string   `json:"expression"`
	IsValid    bool     `json:"is_valid"`
	Error      string   `json:"error,omitempty"`
	NextRuns   []string `json:"next_runs,omitempty"`
	HasYear    bool     `json:"has_year,omitempty"`
	Year       string   `json:"year,omitempty"`
}

// AnalyzeCronExpression 分析cron表达式
func (cp *CronParser) AnalyzeCronExpression(expr string) *CronExpressionInfo {
	info := &CronExpressionInfo{
		Expression: expr,
	}

	if hasYear, actualExp, year, err := cp.ValidateExpression(expr); err != nil {
		info.IsValid = false
		info.Error = err.Error()
		info.HasYear = hasYear
		info.Year = year
		info.Expression = actualExp
		return info
	} else {
		info.HasYear = hasYear
		info.Year = year
		info.Expression = actualExp
	}

	info.IsValid = true

	// 获取接下来5次执行时间
	schedules, err := cp.GetNextNSchedules(expr, time.Now(), 5)
	if err != nil {
		info.Error = err.Error()
		return info
	}

	for _, schedule := range schedules {
		info.NextRuns = append(info.NextRuns, schedule.Format("2006-01-02 15:04:05"))
	}

	return info
}

// MatchYear 匹配年份
func MatchYear(yearStr string, year int) bool {
	if yearStr == "*" || yearStr == "" || yearStr == "?" {
		return true
	}
	// 支持范围：2024-2026
	if strings.Contains(yearStr, "-") {
		parts := strings.Split(yearStr, "-")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			return year >= start && year <= end
		}
	}

	// 支持列表：2024,2025,2026
	if strings.Contains(yearStr, ",") {
		years := strings.Split(yearStr, ",")
		yearStrVal := strconv.Itoa(year)
		for _, y := range years {
			if strings.TrimSpace(y) == yearStrVal {
				return true
			}
		}
		return false
	}

	// 支持步长：2024/2 表示 2024, 2026, 2028...
	if strings.Contains(yearStr, "/") {
		parts := strings.Split(yearStr, "/")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			step, _ := strconv.Atoi(parts[1])
			if step > 0 && (year-start)%step == 0 && year >= start {
				return true
			}
		}
		return false
	}

	// 单个年份
	targetYear, _ := strconv.Atoi(yearStr)
	return year == targetYear
}
