package mcpx

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xiehqing/common/pkg/logs"
	"strconv"
	"time"
)

type Config struct {
	Host                     string `yaml:"host"`
	Port                     int    `yaml:"port"`
	Name                     string `yaml:"name"`
	Version                  string `yaml:"version"`
	WithPromptCapabilities   bool   `yaml:"with-prompt-capabilities"`
	WithResourceCapabilities bool   `yaml:"with-resource-capabilities"`
	WithToolCapabilities     bool   `yaml:"with-tool-capabilities"`
	WithLogging              bool   `yaml:"with-logging"`
	WithRecovery             bool   `yaml:"with-recovery"`
}

func NewMcpServer(cfg *Config, hooks *server.Hooks) *server.MCPServer {
	var opts []server.ServerOption
	if cfg.WithLogging {
		// 启用默认日志中间件
		opts = append(opts, server.WithLogging())
	}
	if cfg.WithRecovery {
		// 启用默认恢复中间件，防止未处理的异常导致服务崩溃
		opts = append(opts, server.WithRecovery())
	}
	if cfg.WithPromptCapabilities {
		// 配置服务端的 Prompt（提示）相关能力,
		// listChanged: 设置为 true 时，当提示列表发生变化（如新增/删除提示）时， 服务端会自动向所有已初始化的客户端发送 prompts/list_changed 通知
		opts = append(opts, server.WithPromptCapabilities(true))
	}
	if cfg.WithResourceCapabilities {
		// mcp-server.WithResourceCapabilities(true, true): 启用服务器资源能力配置
		// 第一个true，允许动态资源分配
		// 第二个true，允许启用资源监控
		opts = append(opts, server.WithResourceCapabilities(true, true))
	}
	if cfg.WithToolCapabilities {
		// 配置服务端的 Tool 相关能力,
		// toolChanged: 设置为 true 时，当工具列表发生变化（如新增/删除工具）时， 服务端会自动向所有已初始化的客户端发送 tools/list_changed 通知
		opts = append(opts, server.WithToolCapabilities(true))
	}
	if hooks == nil {
		// 启用默认日志几率中间件
		// 功能说明:
		//   - 允许在请求处理的关键阶段插入自定义逻辑:
		//     * 请求前/后处理 (Before/AfterRequest)
		//     * 特定方法调用前后 (如 BeforeToolCall)
		//     * 错误返回前的最后处理 (BeforeErrorResponse)
		//   - 典型用例:
		//     1. 请求日志记录
		//     2. 权限校验
		//     3. 指标收集
		//     4. 错误统一格式化
		opts = append(opts, server.WithHooks(hooks))
	}
	mcpServer := server.NewMCPServer(
		cfg.Name,
		cfg.Version,
		opts...,
	)
	return mcpServer
}

func InitSseMcpServer(cfg *Config, mcpServer *server.MCPServer) func() {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	sseServer := server.NewSSEServer(mcpServer)
	go func() {
		logs.Infof("Mcp服务启动成功，监听地址：%s", addr)
		err := sseServer.Start(addr)
		if err != nil {
			panic(err)
		}
	}()
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(1000))
		defer cancel()
		if err := sseServer.Shutdown(ctx); err != nil {
			logs.Errorf("停止Mcp服务失败：%v", err)
		}
		select {
		case <-ctx.Done():
			logs.Infof("Mcp服务停止成功")
		default:
			logs.Infof("Mcp服务停止成功")
		}
	}
}

func GetParamInt64(c mcp.CallToolRequest, key string, defaultValue int64) int64 {
	args := c.GetArguments()
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return defaultValue
}
