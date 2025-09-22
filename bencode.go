package qbit

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

var (
	ErrMalformedInt    = errors.New("malformed integer")
	ErrMalformedString = errors.New("malformed string")
	ErrMalformedList   = errors.New("malformed list")
	ErrMalformedDict   = errors.New("malformed dictionary")
)

// ReadInt returns the integer num and the number of bytes read from b.
// It expects b to begin with the encoded integer itself:
//
// "i42e<remaining bytes>" - will return 42, 4, nil
func ReadInt(b []byte) (num int, read int, err error) {
	if b[0] != 'i' {
		return 0, 0, ErrMalformedInt
	}

	idxE := bytes.IndexByte(b, 'e')
	if idxE == -1 {
		return 0, 0, ErrMalformedInt
	}

	num, err = strconv.Atoi(string(b[1:idxE]))
	if err != nil {
		return 0, 0, ErrMalformedInt
	}

	// including ending separator
	read = len(b[:idxE+1])

	return num, read, nil
}

// Readstring returns the string str and the number of bytes read from b.
// It expects b to begin with the string itself:
//
// "4:spam<remaining bytes>" - will return "spam", 6, nil
func ReadString(b []byte) (str string, read int, err error) {
	if b[0] <= '0' || b[0] >= '9' {
		return "", 0, ErrMalformedString
	}

	idx := bytes.IndexByte(b, ':')
	if idx == -1 {
		return "", 0, ErrMalformedString
	}

	ln, err := strconv.Atoi(string(b[:idx]))
	if err != nil {
		return "", 0, ErrMalformedString
	}

	read = idx + 1 + ln

	// from ":", excluding ":", amount of chars defined
	str = string(b[idx+1 : read])

	return str, read, nil
}

// ReadList returns list and the number of bytes read from b.
// It expects b to begin with the list itself:
//
// "l3:cow3:mooe<remaining bytes>" - will return {"cow", "moo"}, 12, nil
func ReadList(b []byte) (list []any, read int, err error) {
	if b[0] != 'l' {
		return nil, 0, ErrMalformedList
	}

	read = 1
	var (
		v        any
		readOnce int
	)

	for {
		if read >= len(b) {
			return nil, 0, errors.New("list not terminated")
		}

		v, readOnce, err = readAny(b[read:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				read++
				break
			}
			return nil, 0, err
		}
		list = append(list, v)
		read += readOnce
	}

	return list, read, nil
}

// ReadDictionary returns dict and the number of bytes read from b.
// It expects b to begin with the dictionary itself:
//
// "d3:cow3:mooe<remaining bytes>" - will return {"cow": "moo"}, 12, nil
func ReadDictionary(b []byte) (dict map[string]any, read int, err error) {
	idxL := bytes.IndexByte(b, 'd')
	if idxL == -1 || idxL != 0 {
		return nil, 0, ErrMalformedList
	}
	read = 1

	dict = make(map[string]any)

	var (
		k        string
		v        any
		readOnce int
	)

	for {
		if read >= len(b) {
			return nil, 0, ErrMalformedDict
		}
		if b[read] == 'e' {
			read++
			break
		}

		// first read the key and expect a string
		k, readOnce, err = ReadString(b[read:])
		if err != nil {
			return nil, 0, err
		}
		read += readOnce

		// need to check again
		if read >= len(b) {
			return nil, 0, ErrMalformedDict
		}

		// then read a value
		v, readOnce, err = readAny(b[read:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, 0, ErrMalformedDict
			}
			return nil, 0, err
		}
		dict[k] = v
		read += readOnce
	}

	return dict, read, nil
}

func readAny(b []byte) (any, int, error) {
	switch b[0] {
	case 'e':
		return nil, 0, io.EOF
	case 'i':
		return ReadInt(b)
	case 'l':
		return ReadList(b)
	case 'd':
		return ReadDictionary(b)
	default:
		return ReadString(b)
	}
}
