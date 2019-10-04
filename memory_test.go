package bolt

import (
	"os"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestMemory_smoke(t *testing.T) {
	now := time.Now().String()
	dbpath := "/tmp/joe_memory_test_" + now

	m, err := NewMemory("/tmp/joe_memory_test_" + now)
	assert.NilError(t, err)
	defer m.Close()
	defer os.Remove(dbpath)

	var (
		key   = "_bucket_/_key_"
		value = []byte(now)
	)

	err = m.Set(key, value)
	assert.NilError(t, err)

	result, ok, err := m.Get(key)
	assert.NilError(t, err)
	assert.Assert(t, ok == true)
	assert.DeepEqual(t, value, result)

	keys, err := m.Keys()
	assert.NilError(t, err)
	assert.Assert(t, cmp.Contains(keys, key))

	deleted, err := m.Delete(key)
	assert.NilError(t, err)
	assert.Assert(t, deleted == true)

	keys2, err := m.Keys()
	assert.NilError(t, err)
	assert.Assert(t, !cmp.Contains(keys2, key)().Success())
}
