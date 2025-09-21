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

func TestReadString(t *testing.T) {
	// A 4 character string:
	buf := []byte("4:spam")
	str, read, err := ReadString(buf)
	require.NoError(t, err)
	assert.Equal(t, "spam", str)
	assert.Equal(t, 6, read)

	// A long string containing spaces:
	buf = []byte("444:Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum")
	str, read, err = ReadString(buf)
	require.NoError(t, err)
	assert.Equal(t, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum", str)
	assert.Equal(t, 448, read)

	// A string with leftover bytes:
	buf = []byte("4:spam5:ligma")
	str, read, err = ReadString(buf)
	require.NoError(t, err)
	assert.Equal(t, "spam", str)
	assert.Equal(t, 6, read)

	// Wrong format:
	buf = []byte("wrong format")
	str, read, err = ReadString(buf)
	require.Error(t, err)
}
