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

type bucketkey struct {
	bucket, key []byte
}

func (bk *bucketkey) String() string {
	return path.Join(string(bk.bucket), string(bk.key))
}

func BucketkeyFromString(in string) *bucketkey {
	bucket, key := path.Split(in)
	if bucket == "" {
		bucket = "_joe"
	}
	return &bucketkey{bucket: []byte(bucket), key: []byte(key)}
}

func (m *memory) Set(key string, value []byte) error {
	bk := BucketkeyFromString(key)
	m.logger.Debug("Database access: Put",
		zap.ByteString("bucket", bk.bucket),
		zap.ByteString("key", bk.key),
	)

	err := m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bk.bucket)
		if err != nil {
			return err
		}
		return b.Put(bk.key, value)
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *memory) Get(key string) ([]byte, bool, error) {
	var value []byte
	bk := BucketkeyFromString(key)
	m.logger.Debug("Database access: Get",
		zap.ByteString("bucket", bk.bucket),
		zap.ByteString("key", bk.key),
	)

	err := m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bk.bucket)
		value = b.Get(bk.key)
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
	bk := BucketkeyFromString(key)
	m.logger.Debug("Database access: Delete",
		zap.ByteString("bucket", bk.bucket),
		zap.ByteString("key", bk.key),
	)

	err := m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bk.bucket)
		return b.Delete(bk.key)
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
				bk := &bucketkey{bucket: bucket, key: key}
				keys = append(keys, bk.String())
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
