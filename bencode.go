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
	if idxI == -1 || idxI != 0 {
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

var StringSep = []byte(":")

func ReadString(b []byte) (str string, consumed int, err error) {
	idx := bytes.Index(b, StringSep)
	if idx == -1 {
		return "", 0, ErrMalformedString
	}

	ln, err := strconv.Atoi(string(b[:idx]))
	if err != nil {
		return "", 0, ErrMalformedString
	}

	consumed = idx + 1 + ln

	// from ":", excluding ":", amount of chars defined
	str = string(b[idx+1 : consumed])

	return str, consumed, nil
}

var ListSep = []byte("l")

var ErrMalformedList = errors.New("malformed list")

func ReadList(b []byte) (list []any, consumed int, err error) {
	idxL := bytes.Index(b, ListSep)
	if idxL == -1 || idxL != 0 {
		return nil, 0, ErrMalformedList
	}

	consumed = 1
loop:
	for {
		if consumed >= len(b) {
			return nil, 0, errors.New("list not terminated")
		}

		var (
			v    any
			read int
		)

		switch b[consumed] {
		case 'e':
			consumed++
			break loop
		case 'i':
			v, read, err = ReadInt(b[consumed:])
		case 'l':
			v, read, err = ReadList(b[consumed:])
		case 'd':
			v, read, err = ReadDictionary(b[consumed:])
		default:
			if b[consumed] >= '0' && b[consumed] <= '9' {
				v, read, err = ReadString(b[consumed:])
			} else {
				return nil, 0, ErrMalformedList
			}
		}

		if err != nil {
			return nil, 0, err
		}
		list = append(list, v)
		consumed += read
	}

	return list, consumed, nil
}

func ReadDictionary(b []byte) (dict map[string]any, consumed int, err error) {
	return nil, 0, errors.New("not implemented yet")
}
