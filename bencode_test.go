package qbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadInt(t *testing.T) {
	// Small int:
	buf := []byte("i3e")
	n, read, err := ReadInt(buf)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, read)

	// Big int:
	buf = []byte("i2502859205e")
	n, read, err = ReadInt(buf)
	require.NoError(t, err)
	assert.Equal(t, 2502859205, n)
	assert.Equal(t, 12, read)

	// Negative int:
	buf = []byte("i-3e")
	n, read, err = ReadInt(buf)
	require.NoError(t, err)
	assert.Equal(t, -3, n)
	assert.Equal(t, 4, read)

	// Has leftover bytes:
	buf = []byte("i2502859205e4:spam")
	n, read, err = ReadInt(buf)
	require.NoError(t, err)
	assert.Equal(t, 2502859205, n)
	assert.Equal(t, 12, read)

	// Completely wrong syntax:
	buf = []byte("not a number")
	n, _, err = ReadInt(buf)
	require.Error(t, err)

	// Begins with i, ends with e, but not a number:
	buf = []byte("inot a numbere")
	n, _, err = ReadInt(buf)
	require.Error(t, err)
}
