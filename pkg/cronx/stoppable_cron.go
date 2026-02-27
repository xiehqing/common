package cron

import (
	"github.com/robfig/cron/v3"
	"golang.org/x/net/context"
	"sync"
)

type StoppableCron struct {
	cron     *cron.Cron
	mu       sync.Mutex
	running  bool
	stopChan chan struct{}
}

func NewStoppableCron() *StoppableCron {
	return &StoppableCron{
		cron:     cron.New(cron.WithSeconds()),
		stopChan: make(chan struct{}),
		running:  false,
	}
}

func (sc *StoppableCron) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return sc.cron.AddFunc(spec, cmd)
}

func (sc *StoppableCron) Entry(id cron.EntryID) cron.Entry {
	for _, entry := range sc.cron.Entries() {
		if id == entry.ID {
			return entry
		}
	}
	return cron.Entry{}
}

func (sc *StoppableCron) Start() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.running {
		sc.running = true
		sc.cron.Start()
	}
}

func (sc *StoppableCron) Stop() context.Context {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.running {
		sc.running = false
		close(sc.stopChan)
		return sc.cron.Stop()
	}
	return context.Background()
}

func (sc *StoppableCron) ImmediateStop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.running {
		sc.running = false
		// 立即停止，不等待当前执行的任务完成
		sc.cron.Stop()
	}
}
