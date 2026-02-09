package middleware

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

func SetLogIdMW() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		logID := uuid.New().String()
		ctx = context.WithValue(ctx, "log-id", logID)

		c.Header("X-Log-ID", logID)
		c.Next(ctx)
	}
}
