package middleware

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hatcher/common/pkg/logs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unsafe"
)

func AccessLogMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()
		ctx.Next(c)

		status := ctx.Response.StatusCode()
		path := bytesToString(ctx.Request.URI().PathOriginal())
		latency := time.Since(start)
		method := bytesToString(ctx.Request.Header.Method())
		clientIp := ctx.ClientIP()

		handlerPkgPath := strings.Split(ctx.HandlerName(), "/")
		handleName := ""
		if len(handlerPkgPath) > 0 {
			handleName = handlerPkgPath[len(handlerPkgPath)-1]
		}

		baseLog := fmt.Sprintf("| %s | %s | %d | %v | %s | %s | %v | %s ",
			string(ctx.GetRequest().Scheme()), ctx.Host(), status,
			latency, clientIp, method, path, handleName)
		switch {
		case status >= http.StatusInternalServerError:
			logs.CtxErrorf(c, "%s", baseLog)
		case status >= http.StatusBadRequest:
			logs.CtxWarnf(c, "%s", baseLog)
		default:
			urlQuery := ctx.Request.URI().QueryString()
			reqBody := bytesToString(ctx.Request.Body())
			respBody := bytesToString(ctx.Response.Body())
			maxPrintLen := 3 * 1024
			if len(respBody) > maxPrintLen {
				respBody = respBody[:maxPrintLen]
			}
			if len(reqBody) > maxPrintLen {
				reqBody = reqBody[:maxPrintLen]
			}

			if filepath.Ext(path) == "" {
				logs.CtxInfof(c, "%s ", baseLog)
				logs.CtxDebugf(c, "query : %s \nreq : %s \nresp: %s", urlQuery, reqBody, respBody)
			}
		}

	}
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b)) // nolint
}
