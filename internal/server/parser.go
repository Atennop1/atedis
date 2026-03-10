package server

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type RespType byte

const (
	RespTypeString  RespType = '+'
	RespTypeError   RespType = '-'
	RespTypeInteger RespType = ':'
	RespTypeBulk    RespType = '$'
	RespTypeArray   RespType = '*'
)

type RespValue struct {
	typ RespType

	str   string
	num   int
	bulk  string
	array []RespValue
}

type RespReader struct {
	reader *bufio.Reader
}

func NewRespReader(rd io.Reader) *RespReader {
	return &RespReader{
		reader: bufio.NewReader(rd),
	}
}

func (r *RespReader) Read() (RespValue, error) {
	typ, err := r.reader.ReadByte()
	if err != nil {
		return RespValue{}, fmt.Errorf("parser: failed to read type byte: %w", err)
	}

	switch RespType(typ) {
	case RespTypeArray:
		return r.readArray()
	case RespTypeBulk:
		return r.readBulk()
	}

	return RespValue{}, fmt.Errorf("parser: failed to parse unknown type %b", typ)
}

func (r *RespReader) readArray() (RespValue, error) {
	length, _, err := r.readInteger()
	if err != nil {
		return RespValue{}, fmt.Errorf("parser: failed to read size of an array: %w", err)
	}

	result := RespValue{
		typ:   RespTypeArray,
		array: make([]RespValue, length),
	}

	for i := range length {
		result.array[i], err = r.Read()
		if err != nil {
			return RespValue{}, fmt.Errorf("parser: failed to read array element: %w", err)
		}
	}

	return result, nil
}

func (r *RespReader) readBulk() (RespValue, error) {
	length, _, err := r.readInteger()
	if err != nil {
		return RespValue{}, fmt.Errorf("parser: failed to read size of bulk: %w", err)
	}

	bulk := make([]byte, length)
	_, err = r.reader.Read(bulk)
	if err != nil {
		return RespValue{}, fmt.Errorf("parser: failed to read bulk: %w", err)
	}

	_, _, err = r.readLine()
	if err != nil {
		return RespValue{}, fmt.Errorf("parser: failed to read traling CRLF: %w", err)
	}

	return RespValue{
		typ:  RespTypeBulk,
		bulk: string(bulk),
	}, nil
}

func (r *RespReader) readInteger() (int, int, error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, fmt.Errorf("parser: failed to read line: %w", err)
	}

	result, err := strconv.Atoi(line)
	if err != nil {
		return 0, 0, fmt.Errorf("parser: failed to parse integer: %w", err)
	}

	return result, n, nil
}

func (r *RespReader) readLine() (string, int, error) {
	var n int
	result := make([]byte, 0)

	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return "", 0, fmt.Errorf("parser: failed to read byte: %w", err)
		}

		n++
		result = append(result, b)
		if len(result) >= 2 && result[len(result)-2] == '\r' {
			break
		}
	}

	return string(result[:len(result)-2]), n, nil
}
