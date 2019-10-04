package bolt // import "github.com/robertgzr/joe-bolt-memory"

import (
	"os"
	"path"

	"github.com/go-joe/joe"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

type memory struct {
	path    string
	mode    os.FileMode
	options *bolt.Options
	logger  *zap.Logger

	db *bolt.DB
}

func Memory(path string) joe.Module {
	return joe.ModuleFunc(func(conf *joe.Config) error {
		mem, err := NewMemory(path, WithLogger(conf.Logger("memory")))
		if err != nil {
			return err
		}

		conf.SetMemory(mem)
		return nil
	})
}

func NewMemory(path string, opts ...Option) (joe.Memory, error) {
	m := &memory{
		path:    path,
		mode:    0666,
		options: nil,
	}

	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	if m.logger == nil {
		m.logger = zap.NewNop()
	}

	m.logger.Debug("Opening database", zap.String("path", m.path))
	db, err := bolt.Open(m.path, m.mode, m.options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}
	m.db = db

	m.logger.Info("Memory initialized successfully",
		zap.String("path", m.path),
	)
	return m, nil
}

// pathkey is the structure used to encode/decode a bolt bucket and key.
// It will extract the `bucket` name from the root of the path and use the
// remaining components as the key.
type pathkey struct {
	bucket, key []byte
}

func (bk *pathkey) String() string {
	return path.Join(string(bk.bucket), string(bk.key))
}

func pathkeyFromString(in string) *pathkey {
	bucket, key := path.Split(in)
	if bucket == "" {
		bucket = "_joe"
	}
	return &pathkey{bucket: []byte(bucket), key: []byte(key)}
}

func (m *memory) Set(key string, value []byte) error {
	pk := pathkeyFromString(key)
	m.logger.Debug("Database access: Put",
		zap.ByteString("bucket", pk.bucket),
		zap.ByteString("key", pk.key),
	)

	err := m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(pk.bucket)
		if err != nil {
			return err
		}
		return b.Put(pk.key, value)
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *memory) Get(key string) ([]byte, bool, error) {
	var value []byte
	pk := pathkeyFromString(key)
	m.logger.Debug("Database access: Get",
		zap.ByteString("bucket", pk.bucket),
		zap.ByteString("key", pk.key),
	)

	err := m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(pk.bucket)
		value = b.Get(pk.key)
		return nil
	})
	if err != nil {
		if err == bolt.ErrBucketNotFound || os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, true, err
	}
	return value, true, nil
}

func (m *memory) Delete(key string) (bool, error) {
	pk := pathkeyFromString(key)
	m.logger.Debug("Database access: Delete",
		zap.ByteString("bucket", pk.bucket),
		zap.ByteString("key", pk.key),
	)

	err := m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(pk.bucket)
		return b.Delete(pk.key)
	})
	if err != nil {
		if err == bolt.ErrBucketNotFound || os.IsNotExist(err) {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

func (m *memory) Keys() ([]string, error) {
	var keys []string
	err := m.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(bucket []byte, b *bolt.Bucket) error {
			return b.ForEach(func(key, _ []byte) error {
				pk := &pathkey{bucket: bucket, key: key}
				keys = append(keys, pk.String())
				return nil
			})
		})
	})
	if err != nil {
		return keys, err
	}
	return keys, nil
}

func (m *memory) Close() error {
	m.logger.Debug("Closing database", zap.String("path", m.path))
	return m.db.Close()
}
