package logs

/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type LogConfig struct {
	Level  string `json:"level" mapstructure:"level"`
	Output Output `json:"output" mapstructure:"output"`
	Path   string `json:"path" mapstructure:"path"`
	File   string `json:"file" mapstructure:"file"`
}

func (cfg *LogConfig) Prepare() {
	if cfg.Output == "" {
		cfg.Output = Stdout
	}
	if cfg.Path == "" {
		cfg.Path = "logs"
	}
}

// CreateFileWriter 构建日志文件写入器
func CreateFileWriter(path, name string) (io.Writer, error) {
	os.MkdirAll(path, 0755)
	var f *os.File
	file := filepath.Join(path, name)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件错误, err: %v", err)
	}
	return f, nil
}

func InitLogger(cfg LogConfig, defaultLogFile string) error {
	cfg.Prepare()
	if cfg.File == "" {
		cfg.File = defaultLogFile
	}
	level := GetLevel(cfg.Level)
	SetLevel(level)
	if cfg.Output == Stdout {
		SetOutput(os.Stdout)
	} else if cfg.Output == Stderr {
		SetOutput(os.Stderr)
	} else if cfg.Output == File {
		writer, err := CreateFileWriter(cfg.Path, cfg.File)
		if err != nil {
			return err
		}
		SetOutput(writer)
	}
	return nil
}

var logger FullLogger = &ILog{
	level:  LevelInfo,
	stdLog: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
}

// SetOutput sets the output of default logs. By default, it is stderr.
func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

// SetLevel sets the level of logs below which logs will not be output.
// The default log level is LevelTrace.
// Note that this method is not concurrent-safe.
func SetLevel(lv Level) {
	logger.SetLevel(lv)
}

// DefaultLogger return the default logs for kitex.
func DefaultLogger() FullLogger {
	return logger
}

// SetLogger sets the default logs.
// Note that this method is not concurrent-safe and must not be called
// after the use of DefaultLogger and global functions in this package.
func SetLogger(v FullLogger) {
	logger = v
}

// Fatal calls the default logs's Fatal method and then os.Exit(1).
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Error calls the default logs's Error method.
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Warn calls the default logs's Warn method.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Notice calls the default logs's Notice method.
func Notice(v ...interface{}) {
	logger.Notice(v...)
}

// Info calls the default logs's Info method.
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Debug calls the default logs's Debug method.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Trace calls the default logs's Trace method.
func Trace(v ...interface{}) {
	logger.Trace(v...)
}

// Fatalf calls the default logs's Fatalf method and then os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Errorf calls the default logs's Errorf method.
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Warnf calls the default logs's Warnf method.
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// Noticef calls the default logs's Noticef method.
func Noticef(format string, v ...interface{}) {
	logger.Noticef(format, v...)
}

// Infof calls the default logs's Infof method.
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Debugf calls the default logs's Debugf method.
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Tracef calls the default logs's Tracef method.
func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

// CtxFatalf calls the default logs's CtxFatalf method and then os.Exit(1).
func CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxFatalf(ctx, format, v...)
}

// CtxErrorf calls the default logs's CtxErrorf method.
func CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxErrorf(ctx, format, v...)
}

// CtxWarnf calls the default logs's CtxWarnf method.
func CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxWarnf(ctx, format, v...)
}

// CtxNoticef calls the default logs's CtxNoticef method.
func CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	logger.CtxNoticef(ctx, format, v...)
}

// CtxInfof calls the default logs's CtxInfof method.
func CtxInfof(ctx context.Context, format string, v ...interface{}) {
	logger.CtxInfof(ctx, format, v...)
}

// CtxDebugf calls the default logs's CtxDebugf method.
func CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxDebugf(ctx, format, v...)
}

// CtxTracef calls the default logs's CtxTracef method.
func CtxTracef(ctx context.Context, format string, v ...interface{}) {
	logger.CtxTracef(ctx, format, v...)
}
