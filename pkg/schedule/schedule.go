package schedule

import (
	"github.com/hatcher/common/pkg/logs"
	"github.com/robfig/cron/v3"
	"strconv"
	"time"
)

type Scheduler struct {
	quit chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		quit: make(chan struct{}),
	}
}

type ScheduledConfig struct {
	Enabled bool
	Type    string
	Value   string
}

// AddScheduledTask 添加定时任务
func (worker *Scheduler) AddScheduledTask(name string, config ScheduledConfig, method func()) {
	if config.Enabled {
		scheduledType := config.Type
		scheduledValue := config.Value
		if scheduledValue == "" {
			logs.Errorf("%s 定时任务未配置执行频率，sceduleType:%s", name, scheduledType)
			return
		}
		if scheduledType == "cron" {
			worker.AddCronTask(scheduledValue, method)
		} else if scheduledType == "fixed_delay" {
			interval, err := strconv.ParseInt(scheduledValue, 10, 64)
			if err != nil {
				logs.Errorf("%s 定时任务执行频率错误，仅可为数字，sceduleType:%s, sceduleValue:%s", name, scheduledType, scheduledValue)
				return
			}
			worker.AddFixDelayTask(interval, method)
		} else {
			logs.Errorf("%s 定时任务类型错误，scheduleType: %s , 仅支持（fixed_delay 或者 cron）", name, scheduledType)
		}
	} else {
		logs.Infof("%s 定时任务未启用", name)
	}
}

// AddCronTask 添加cron任务
func (worker *Scheduler) AddCronTask(cronString string, method func()) {
	cronTask := cron.New(cron.WithSeconds())
	_, err := cronTask.AddFunc(cronString, method)
	if err != nil {
		logs.Errorf("定时任务Cron表达式错误: %v", err)
		return
	}
	go func() {
		cronTask.Start()
		defer cronTask.Stop()

		// 等待退出信号
		select {
		case <-worker.quit:
			return
		}
	}()
}

// AddFixDelayTask 添加固定延迟任务
func (worker *Scheduler) AddFixDelayTask(interval int64, method func()) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-worker.quit:
				return
			case <-ticker.C:
				method()
			}
		}
	}()
}
