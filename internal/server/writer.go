package server

import (
	"fmt"
	"io"
	"strconv"
)

type RespWriter struct {
	writer io.Writer
}

func NewRespWriter(writer io.Writer) *RespWriter {
	return &RespWriter{
		writer: writer,
	}
}

func (w *RespWriter) Write(v RespValue) error {
	result := v.Marshal()

	_, err := w.writer.Write(result)
	if err != nil {
		return fmt.Errorf("writer: failed to write result: %w", err)
	}

	return nil
}

func (v RespValue) Marshal() []byte {
	switch v.typ {
	case RespTypeString:
		return v.marshalString()
	case RespTypeBulk:
		return v.marshalBulk()
	case RespTypeArray:
		return v.marshalArray()
	case RespTypeError:
		return v.marshalError()
	case RespTypeNull:
		return marshalNull()
	}

	return nil
}

func (v RespValue) marshalString() []byte {
	var result []byte

	result = append(result, byte(RespTypeString))
	result = append(result, v.str...)
	result = append(result, "\r\n"...)

	return result
}

func (v RespValue) marshalBulk() []byte {
	var result []byte

	result = append(result, byte(RespTypeBulk))
	result = append(result, strconv.Itoa(len(v.bulk))...)
	result = append(result, "\r\n"...)
	result = append(result, v.bulk...)
	result = append(result, "\r\n"...)

	return result
}

func (v RespValue) marshalArray() []byte {
	var result []byte

	result = append(result, byte(RespTypeArray))
	result = append(result, strconv.Itoa(len(v.array))...)
	result = append(result, "\r\n"...)

	for i := range len(v.array) {
		result = append(result, v.array[i].Marshal()...)
	}

	return result
}

func (v RespValue) marshalError() []byte {
	var result []byte

	result = append(result, byte(RespTypeError))
	result = append(result, v.str...)
	result = append(result, "\r\n"...)

	return result
}

func marshalNull() []byte {
	return []byte("$-1\r\n")
}
