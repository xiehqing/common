package main

import (
	"context"
	"github.com/hatcher/common/agent/app"
)

func main() {
	a, err := app.New(context.Background(), "", nil, nil, nil, nil)
	if err != nil {
		return
	}
	a.Shutdown()
}
