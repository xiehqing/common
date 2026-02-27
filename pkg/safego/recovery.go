package safego

import (
	"context"
	"github.com/xiehqing/common/pkg/logs"
	"runtime/debug"
)

// Recovery 捕获panic
func Recovery(ctx context.Context) {
	e := recover()
	if e == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	logs.Errorf("[Recovery] cache panic error = %v \n stacktrace = \n%s", e, string(debug.Stack()))
}
