package bolt

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestMemory_smoke(t *testing.T) {
	dbpath, err := ioutil.TempFile("", "joe-memory-test-")
	assert.NilError(t, err)
	defer os.Remove(dbpath.Name())
	m, err := NewMemory(dbpath.Name())
	assert.NilError(t, err)
	defer m.Close()

	var (
		key   = "_bucket_/_key_"
		value = []byte(time.Now().String())
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

func TestGetMissingKey(t *testing.T) {
	dbpath, err := ioutil.TempFile("", "joe-memory-test-")
	assert.NilError(t, err)
	defer os.Remove(dbpath.Name())
	m, err := NewMemory(dbpath.Name())
	assert.NilError(t, err)
	defer m.Close()

	_, ok, err := m.Get("_key_")
	assert.Assert(t, !ok)
}
