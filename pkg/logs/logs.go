package logs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Output string

const (
	Stdout Output = "stdout"
	Stderr Output = "stderr"
	File   Output = "file"
)

type Level int

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelNotice
	LevelWarn
	LevelError
	LevelFatal
)

func GetLevel(level string) Level {
	upperLevel := strings.ToUpper(level)
	switch upperLevel {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "NOTICE":
		return LevelNotice
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

var strs = []string{
	"[TRACE] ",
	"[DEBUG] ",
	"[INFO] ",
	"[NOTICE] ",
	"[WARN] ",
	"[ERROR] ",
	"[FATAL] ",
}

func (lv Level) toString() string {
	if lv >= LevelTrace && lv <= LevelFatal {
		return strs[lv]
	}
	return fmt.Sprintf("[?%d] ", lv)
}

type Logger interface {
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Notice(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
}

type CtxLogger interface {
	CtxTracef(ctx context.Context, format string, v ...interface{})
	CtxDebugf(ctx context.Context, format string, v ...interface{})
	CtxInfof(ctx context.Context, format string, v ...interface{})
	CtxNoticef(ctx context.Context, format string, v ...interface{})
	CtxWarnf(ctx context.Context, format string, v ...interface{})
	CtxErrorf(ctx context.Context, format string, v ...interface{})
	CtxFatalf(ctx context.Context, format string, v ...interface{})
}

type FormatLogger interface {
	Tracef(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Noticef(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

type Control interface {
	SetLevel(Level)
	SetOutput(io.Writer)
}

type FullLogger interface {
	Logger
	FormatLogger
	CtxLogger
	Control
}

type ILog struct {
	stdLog *log.Logger
	level  Level
}

func (il *ILog) SetOutput(w io.Writer) {
	il.stdLog.SetOutput(w)
}

func (il *ILog) SetLevel(lv Level) {
	il.level = lv
}

func (il *ILog) logf(lv Level, format *string, v ...interface{}) {
	if il.level > lv {
		return
	}
	msg := lv.toString()
	if format != nil {
		msg += fmt.Sprintf(*format, v...)
	} else {
		msg += fmt.Sprint(v...)
	}
	il.stdLog.Output(4, msg)
	if lv == LevelFatal {
		os.Exit(1)
	}
}

func (il *ILog) logfCtx(ctx context.Context, lv Level, format *string, v ...interface{}) {
	if il.level > lv {
		return
	}
	msg := lv.toString()
	logID := ctx.Value("log-id")
	if logID != nil {
		msg += fmt.Sprintf("[log-id: %v] ", logID)
	}
	if format != nil {
		msg += fmt.Sprintf(*format, v...)
	} else {
		msg += fmt.Sprint(v...)
	}
	il.stdLog.Output(4, msg)
	if lv == LevelFatal {
		os.Exit(1)
	}
}

func (il *ILog) Fatal(v ...interface{}) {
	il.logf(LevelFatal, nil, v...)
}

func (il *ILog) Error(v ...interface{}) {
	il.logf(LevelError, nil, v...)
}

func (il *ILog) Warn(v ...interface{}) {
	il.logf(LevelWarn, nil, v...)
}

func (il *ILog) Notice(v ...interface{}) {
	il.logf(LevelNotice, nil, v...)
}

func (il *ILog) Info(v ...interface{}) {
	il.logf(LevelInfo, nil, v...)
}

func (il *ILog) Debug(v ...interface{}) {
	il.logf(LevelDebug, nil, v...)
}

func (il *ILog) Trace(v ...interface{}) {
	il.logf(LevelTrace, nil, v...)
}

func (il *ILog) Fatalf(format string, v ...interface{}) {
	il.logf(LevelFatal, &format, v...)
}

func (il *ILog) Errorf(format string, v ...interface{}) {
	il.logf(LevelError, &format, v...)
}

func (il *ILog) Warnf(format string, v ...interface{}) {
	il.logf(LevelWarn, &format, v...)
}

func (il *ILog) Noticef(format string, v ...interface{}) {
	il.logf(LevelNotice, &format, v...)
}

func (il *ILog) Infof(format string, v ...interface{}) {
	il.logf(LevelInfo, &format, v...)
}

func (il *ILog) Debugf(format string, v ...interface{}) {
	il.logf(LevelDebug, &format, v...)
}

func (il *ILog) Tracef(format string, v ...interface{}) {
	il.logf(LevelTrace, &format, v...)
}

func (il *ILog) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	il.logfCtx(ctx, LevelFatal, &format, v...)
}

func (il *ILog) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	il.logfCtx(ctx, LevelError, &format, v...)
}

func (il *ILog) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	il.logfCtx(ctx, LevelWarn, &format, v...)
}

func (il *ILog) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	il.logfCtx(ctx, LevelNotice, &format, v...)
}

func (il *ILog) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	il.logfCtx(ctx, LevelInfo, &format, v...)
}

func (il *ILog) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	il.logfCtx(ctx, LevelDebug, &format, v...)
}

func (il *ILog) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	il.logfCtx(ctx, LevelTrace, &format, v...)
}
