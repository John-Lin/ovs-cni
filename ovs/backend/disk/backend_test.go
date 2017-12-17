package disk

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var store Store

const (
	ID     = "testing-host"
	IFNAME = "veth"
)

func TestNewStore(t *testing.T) {
	store, err := New("", "./test")
	assert.NotNil(t, store)
	assert.NoError(t, err)
}

func TestReserve(t *testing.T) {
	find, err := store.Reserve(ID, IFNAME)
	assert.True(t, find)
	assert.NoError(t, err)
	find, err = store.Reserve(ID, IFNAME)
	assert.False(t, find)
	assert.NoError(t, err)
}

func TestReleaseByID(t *testing.T) {
	name, err := store.ReleaseByID(ID)
	assert.NoError(t, err)
	assert.Equal(t, IFNAME, name)

	name, err = store.ReleaseByID(ID)
	assert.Error(t, err)
}
