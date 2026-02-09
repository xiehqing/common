package app

import (
	"github.com/xiehaiqing/common/pkg/logs"
	"sync"
)

var state *State

func init() {
	logs.Infof("Initializing agent state")
	state = NewState()
}

type State struct {
	AppState sync.Map
}

// AddApp 添加app
func AddApp(requestId string, app *App) {
	state.AppState.Store(requestId, app)
}

// RemoveApp 移除app
func RemoveApp(requestId string) {
	state.AppState.Delete(requestId)
}

// ShutdownApp 关闭app
func ShutdownApp(requestId string) bool {
	app, ok := state.AppState.Load(requestId)
	if ok {
		app.(*App).Shutdown()
		RemoveApp(requestId)
		return true
	}
	return false
}

func NewState() *State {
	return &State{
		AppState: sync.Map{},
	}
}
