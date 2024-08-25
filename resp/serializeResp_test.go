package resp

import (
	"reflect"
	"testing"
)

func TestSerializeSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    SimpleString
		expected []byte
	}{
		{
			name:     "hello",
			input:    SimpleString{Value: "hello"},
			expected: []byte("+hello\r\n"),
		},
		{
			name:     "world",
			input:    SimpleString{Value: "world"},
			expected: []byte("+world\r\n"),
		},
		{
			name:     "empty string",
			input:    SimpleString{Value: ""},
			expected: []byte("+\r\n"),
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, err := SerializeSimpleString(test.input)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}

func TestSerializeSimpleError(t *testing.T) {
	tests := []struct {
		name     string
		input    SimpleError
		expected []byte
	}{
		{
			name:     "error",
			input:    SimpleError{Message: "error"},
			expected: []byte("-error\r\n"),
		},
		{
			name:     "world",
			input:    SimpleError{Message: "world"},
			expected: []byte("-world\r\n"),
		},
		{
			name:     "empty string",
			input:    SimpleError{Message: ""},
			expected: []byte("-\r\n"),
		},
		{
			name:     "ERR unknown command 'asdf'",
			input:    SimpleError{Message: "ERR unknown command 'asdf'"},
			expected: []byte("-ERR unknown command 'asdf'\r\n"),
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, err := SerializeSimpleError(test.input)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}

func TestSerializeInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    Integer
		expected []byte
	}{
		{
			name:     "1",
			input:    Integer{Value: 1},
			expected: []byte(":1\r\n"),
		},
		{
			name:     "0",
			input:    Integer{Value: 0},
			expected: []byte(":0\r\n"),
		},
		{
			name:     "-1",
			input:    Integer{Value: -1},
			expected: []byte(":-1\r\n"),
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, err := SerializeInteger(test.input)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}

func TestSerializeBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    BulkString
		expected []byte
	}{
		{
			name:     "hello",
			input:    BulkString{Value: stringPtr("hello")},
			expected: []byte("$5\r\nhello\r\n"),
		},
		{
			name:     "world",
			input:    BulkString{Value: stringPtr("world")},
			expected: []byte("$5\r\nworld\r\n"),
		},
		{
			name:     "empty string",
			input:    BulkString{Value: stringPtr("")},
			expected: []byte("$0\r\n\r\n"),
		},
		{
			name:     "nil string",
			input:    BulkString{Value: nil},
			expected: []byte("$-1\r\n"),
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, err := SerializeBulkString(test.input)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}

func TestSerializeArray(t *testing.T) {
	tests := []struct {
		name     string
		input    Array
		expected []byte
	}{
		{
			name:     "empty array",
			input:    Array{Elements: &[]RESPData{}},
			expected: []byte("*0\r\n"),
		},
		{
			name:     "hello world in two bulk strings",
			input:    Array{Elements: &[]RESPData{BulkString{Value: stringPtr("hello")}, BulkString{Value: stringPtr("world")}}},
			expected: []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
		},
		{
			name:     "array with 3 elements",
			input:    Array{Elements: &[]RESPData{SimpleString{Value: "hello"}, SimpleString{Value: "world"}, Integer{Value: 1}}},
			expected: []byte("*3\r\n+hello\r\n+world\r\n:1\r\n"),
		},
		{
			name:     "array with 1 element",
			input:    Array{Elements: &[]RESPData{SimpleString{Value: "hello"}}},
			expected: []byte("*1\r\n+hello\r\n"),
		},
		{
			name:     "nil array",
			input:    Array{Elements: nil},
			expected: []byte("*-1\r\n"),
		},
        {
            name:     "array with nil element between bulk string and simple string",
            input:    Array{Elements: &[]RESPData{BulkString{Value: stringPtr("hello")}, BulkString{Value: nil}, SimpleString{Value: "world"}}},
            expected: []byte("*3\r\n$5\r\nhello\r\n$-1\r\n+world\r\n"),
        },
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, err := SerializeArray(test.input)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}
