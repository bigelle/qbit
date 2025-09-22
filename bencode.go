package qbit

import (
	"bytes"
	"errors"
	"strconv"
)

var (
	ErrMalformedInt    = errors.New("malformed integer")
	ErrMalformedString = errors.New("malformed string")
	ErrMalformedList   = errors.New("malformed list")
	ErrMalformedDict   = errors.New("malformed dictionary")
)

func ReadInt(b []byte) (num int, consumed int, err error) {
	idxI := bytes.IndexByte(b, 'i')
	if idxI == -1 || idxI != 0 {
		return 0, 0, ErrMalformedInt
	}

	idxE := bytes.IndexByte(b, 'e')
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

func ReadString(b []byte) (str string, consumed int, err error) {
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

	consumed = idx + 1 + ln

	// from ":", excluding ":", amount of chars defined
	str = string(b[idx+1 : consumed])

	return str, consumed, nil
}

func ReadList(b []byte) (list []any, consumed int, err error) {
	idxL := bytes.IndexByte(b, 'l')
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
	idxL := bytes.IndexByte(b, 'd')
	if idxL == -1 || idxL != 0 {
		return nil, 0, ErrMalformedList
	}
	consumed = 1

	dict = make(map[string]any)

	for {
		if consumed >= len(b) {
			return nil, 0, ErrMalformedDict
		}
		if b[consumed] == 'e' {
			consumed++
			break
		}

		var (
			k    string
			v    any
			read int
		)

		// first read the key and expect a string
		if b[consumed] >= '0' && b[consumed] <= '9' {
			k, read, err = ReadString(b[consumed:])
			if err != nil {
				return nil, 0, err
			}
			consumed += read
		} else {
			return nil, 0, ErrMalformedDict
		}

		// need to check again
		if consumed >= len(b) {
			return nil, 0, ErrMalformedDict
		}

		// then read a value
		switch b[consumed] {
		case 'e':
			return nil, 0, ErrMalformedDict // has a key, but no value
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
				return nil, 0, ErrMalformedDict
			}
		}

		if err != nil {
			return nil, 0, err
		}
		dict[k] = v
		consumed += read
	}

	return dict, consumed, nil
}
