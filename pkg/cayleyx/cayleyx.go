package cayleyx

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"os"
	// 内存后端
	_ "github.com/cayleygraph/cayley/graph/memstore"
	// BoltDB后端
	_ "github.com/cayleygraph/cayley/graph/kv/bolt"
	"github.com/pkg/errors"
	"github.com/xiehqing/common/pkg/logs"
)

type Config struct {
	Type   GraphType `json:"type" yaml:"type" mapstructure:"type"`
	DBPath string    `json:"dbPath" yaml:"db-path" mapstructure:"db-path"`
}

type GraphType string

const (
	GraphTypeMemory GraphType = "memory"
	GraphTypeBolt   GraphType = "bolt"
)

func (c GraphType) Name() string {
	switch c {
	case GraphTypeMemory:
		return "memstore"
	case GraphTypeBolt:
		return "bolt"
	default:
		return "mem"
	}
}

// NewCayleyGraph 创建cayley图
func NewCayleyGraph(cfg *Config) (*cayley.Handle, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if cfg.Type == "" {
		cfg.Type = GraphTypeMemory
	}
	name := cfg.Type.Name()
	if cfg.Type == GraphTypeMemory {
		handle, err := cayley.NewMemoryGraph()
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to create memory graph db: %v", err)
		}
		return handle, nil
	}
	err := graph.InitQuadStore(name, cfg.DBPath, nil)
	if err != nil {
		if errors.Is(err, graph.ErrDatabaseExists) {
			logs.Debug("database already exists")
		} else {
			return nil, errors.WithMessagef(err, "failed to init quad store: %v", err)
		}
	}
	// 打开数据库
	store, err := cayley.NewGraph(name, cfg.DBPath, nil)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to open graph db: %v", err)
	}
	return store, nil
}

// ClearBoltGraph 清空BoltDB图数据库
func ClearBoltGraph(store *cayley.Handle, fullDbPath string) error {
	if store == nil {
		return nil
	}
	store.Close()
	// 删除数据库文件
	if err := os.RemoveAll(fullDbPath); err != nil {
		return errors.WithMessagef(err, "删除数据库文件失败：%v", err)
	}
	return nil
}
