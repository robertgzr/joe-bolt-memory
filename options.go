package bolt // import "github.com/robertgzr/joe-bolt-memory"

import (
	"os"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

type Option func(*memory) error

func WithLogger(logger *zap.Logger) Option {
	return func(m *memory) error {
		m.logger = logger
		return nil
	}
}

func WithFileMode(mode os.FileMode) Option {
	return func(m *memory) error {
		m.mode = mode
		return nil
	}
}

func WithOptions(opts *bolt.Options) Option {
	return func(m *memory) error {
		m.options = opts
		return nil
	}
}
