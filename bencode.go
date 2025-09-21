package qbit

import (
	"bytes"
	"errors"
	"strconv"
)

var (
	BeginIntSep = []byte("i")
	EndSep      = []byte("e")
)

var (
	ErrMalformedInt    = errors.New("malformed integer")
	ErrMalformedString = errors.New("malformed string")
)

func ReadInt(b []byte) (num int, consumed int, err error) {
	idxI := bytes.Index(b, BeginIntSep)
	if idxI == -1 {
		return 0, 0, ErrMalformedInt
	}

	idxE := bytes.Index(b, EndSep)
	if idxE == -1 {
		return 0, 0, ErrMalformedInt
	}

	num, err = strconv.Atoi(string(b[idxI+1 : idxE]))
	if err != nil {
		return 0, 0, ErrMalformedInt
	}

	// including beginning and ending separator
	consumed = len(b[idxI : idxE+1])

	return num, consumed, nil
}
