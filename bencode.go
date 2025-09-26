package qbit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

var (
	ErrMalformedInt    = errors.New("malformed integer")
	ErrMalformedString = errors.New("malformed string")
	ErrMalformedList   = errors.New("malformed list")
	ErrMalformedDict   = errors.New("malformed dictionary")

	ErrNotEnoughData = errors.New("not enough data")
)

type BencodeDecoder struct {
	readPos int
	src     io.Reader
	buf     []byte
}

func NewDecoder(src io.Reader) *BencodeDecoder {
	return &BencodeDecoder{
		readPos: 0,
		src:     src,
		buf:     make([]byte, 4*1024),
	}
}

func (d *BencodeDecoder) Decode(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return errors.New("v must be a pointer")
	}
	if rv.IsNil() {
		return errors.New("v can't be nil")
	}

	err := d.decodeAny(rv)
	if err != nil {
		return err
	}

	return nil
}

func (d *BencodeDecoder) decodeAny(rv reflect.Value) error {
	b, err := d.ReadByte()
	if err != nil {
		return err
	}

	switch b {
	case 'i':
		err = d.decodeInt(rv)
	default:
		err = d.decodeString(rv)
	}

	if err != nil {
		return err
	}
	return nil
}

func (d *BencodeDecoder) decodeString(rv reflect.Value) (err error) {
	var (
		n   int
		str string
	)
	for {
		n, err = d.src.Read(d.buf[d.readPos:])
		if err != nil {
			return err
		}

		str, n, err = ParseString(d.buf[:d.readPos+n])
		if err == nil {
			rv.Elem().SetString(str)
			return nil
		}

		if errors.Is(err, ErrNotEnoughData) {
			continue
		}
		return err
	}
}

func (d *BencodeDecoder) decodeInt(rv reflect.Value) (err error) {
	var (
		n int
		i int
	)
	for {
		n, err = d.src.Read(d.buf[d.readPos:])
		if err != nil {
			return err
		}

		i, n, err = ParseInt(d.buf[:d.readPos+n])
		if err == nil {
			rv.Elem().SetInt(int64(i))
			return nil
		}

		if errors.Is(err, ErrNotEnoughData) {
			continue
		}
		return err
	}
}

func (d *BencodeDecoder) ReadByte() (byte, error) {
	n, err := d.src.Read(d.buf[d.readPos : d.readPos+1])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	d.readPos++

	return d.buf[d.readPos-1], nil
}

// ParseInt returns the integer num and the number of bytes read from b.
// It expects b to begin with the encoded integer itself:
//
// "i42e<remaining bytes>" - will return 42, 4, nil
func ParseInt(b []byte) (num int, read int, err error) {
	if b[0] != 'i' {
		return 0, 0, ErrMalformedInt
	}

	idxE := bytes.IndexByte(b, 'e')
	if idxE == -1 {
		return 0, 0, ErrNotEnoughData
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
func ParseString(b []byte) (str string, read int, err error) {
	if (b[0] < '0' || b[0] > '9') && b[0] != '-' {
		return "", 0, ErrMalformedString
	}

	idx := bytes.IndexByte(b, ':')
	if idx == -1 {
		return "", 0, ErrNotEnoughData
	}

	ln, err := strconv.Atoi(string(b[:idx]))
	if err != nil {
		fmt.Println(string(b))
		return "", 0, ErrMalformedString
	}

	if idx+1+ln > len(b) {
		return "", 0, ErrNotEnoughData
	}

	read = idx + 1 + ln

	// from ":", excluding ":", amount of chars defined
	str = string(b[idx+1 : read])

	return str, read, nil
}

// ParseList returns list and the number of bytes read from b.
// It expects b to begin with the list itself:
//
// "l3:cow3:mooe<remaining bytes>" - will return {"cow", "moo"}, 12, nil
func ParseList(b []byte) (list []any, read int, err error) {
	if b[0] != 'l' {
		return nil, 0, ErrMalformedList
	}

	read = 1
	var (
		v any
		n int
	)

	for {
		if read >= len(b) {
			return nil, 0, ErrNotEnoughData
		}

		v, n, err = readAny(b[read:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				read++
				break
			}
			return nil, 0, err
		}
		list = append(list, v)
		read += n
	}

	return list, read, nil
}

// ParseDictionary returns dict and the number of bytes read from b.
// It expects b to begin with the dictionary itself:
//
// "d3:cow3:mooe<remaining bytes>" - will return {"cow": "moo"}, 12, nil
func ParseDictionary(b []byte) (dict map[string]any, read int, err error) {
	if b[0] != 'd' {
		return nil, 0, ErrMalformedDict
	}

	read = 1
	dict = make(map[string]any)

	var (
		k string
		v any
		n int
	)

	for {
		if read >= len(b) {
			return nil, 0, ErrNotEnoughData
		}
		if b[read] == 'e' {
			read++
			break
		}

		// first read the key and expect a string
		k, n, err = ParseString(b[read:])
		if err != nil {
			return nil, 0, err
		}
		read += n

		// need to check again
		if read >= len(b) {
			return nil, 0, ErrNotEnoughData
		}

		// then read a value
		v, n, err = readAny(b[read:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, 0, ErrMalformedDict
			}
			return nil, 0, err
		}
		dict[k] = v
		read += n
	}

	return dict, read, nil
}

func readAny(b []byte) (any, int, error) {
	switch b[0] {
	case 'e':
		return nil, 0, io.EOF
	case 'i':
		return ParseInt(b)
	case 'l':
		return ParseList(b)
	case 'd':
		return ParseDictionary(b)
	default:
		return ParseString(b)
	}
}
