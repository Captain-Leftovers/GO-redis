package resp

import (
	"reflect"
	"testing"
)

func TestParseSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
		hasError bool
	}{
		{
			name:     "hello",
			input:    []byte("+hello\r\n"),
			expected: "hello",
			hasError: false,
		},
		{
			name:     "world",
			input:    []byte("+world\r\n"),
			expected: "world",
			hasError: false,
		},
		{
			name:     "incorrectformat",
			input:    []byte("+incorrectformat"),
			expected: "",
			hasError: true,
		},
		{
			name:     "empty string return  and trailing",
			input:    []byte("+\r\nempty string trailing"),
			expected: "",
			hasError: false,
		},
		{
			name:     "missing + prefix",
			input:    []byte("hello\r\n"), // Missing "+"
			expected: "",
			hasError: true,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, _, err := parseSimpleString(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("expected an error for input %s, but got none", string(test.input))
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error for input %s, but got %v", string(test.input), err)
				}
				if result.Value != test.expected {
					t.Errorf("expected %s, but got %s", test.expected, result.Value)
				}
			}

		})

	}

}

func TestParseSimpleError(t *testing.T) {

	tests := []struct {
		name     string
		input    []byte
		expected string
		hasError bool
	}{
		{
			name:     "incorrect crlf slash / ",
			input:    []byte("-Error of some kind/r/n"),
			expected: "Error of some kind no crlf err",
			hasError: true,
		},
		{
			name:     "correct message",
			input:    []byte("-Err message goes here XoXo\r\n"),
			expected: "Err message goes here XoXo",
			hasError: false,
		},
		{
			name:     "cant have empty error message considered malformed",
			input:    []byte("-\r\nError of some kind\r\n"),
			expected: "Error of some kind",
			hasError: true,
		},
		{
			name:     "correct message with trailing resp",
			input:    []byte("-first Error here\r\nsecond Error of some kind\r\n"),
			expected: "first Error here",
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, _, err := parseSimpleError(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("expected an error for input %s, but got none", string(test.input))
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error for input %s, but got %v", string(test.input), err)
				}
				if result.Message != test.expected {
					t.Errorf("expected %s, but got %s", test.expected, result.Message)
				}
			}
		})
	}
}

func TestParseIntegers(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int
		hasError bool
	}{
		{
			name:     "passing 333",
			input:    []byte(":333\r\n"),
			expected: 333,
			hasError: false,
		},
		{
			name:     "erroring +-33",
			input:    []byte(":+-33\r\n"),
			expected: 0,
			hasError: true,
		},
		{
			name:     "passing negative -444",
			input:    []byte(":-444\r\n"),
			expected: -444,
			hasError: false,
		},
		{
			name:     "passing +777 with + infront",
			input:    []byte(":+777\r\n"),
			expected: 777,
			hasError: false,
		},
		{
			name:     "exp +999 plus trialing data",
			input:    []byte(":+999\r\n100"),
			expected: 999,
			hasError: false,
		},
		{
			name:     "exp  error double --",
			input:    []byte(":--888\r\n"),
			expected: -888,
			hasError: true,
		},
		{
			name:     "exp 0",
			input:    []byte(":0\r\n"),
			expected: 0,
			hasError: false,
		},
		{
			name:     "malformed no value between:'\r\n'",
			input:    []byte(":\r\n555\r\n"),
			expected: 0,
			hasError: true,
		},
	}

	for _, test := range tests {
		result, _, err := parseInteger(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("expected an error for input %s, but got none", string(test.input))
			}
		} else {
			if err != nil {
				t.Errorf("did not expect an error for input %s, but got %v", string(test.input), err)
			}
			if result.Value != test.expected {
				t.Errorf("expected %d, but got %v", test.expected, result.Value)
			}
		}
	}
}

func TestParseBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected BulkString
		trailing []byte
		hasError bool
	}{
		{
			name:     "incorrect format",
			input:    []byte("+hello\r\n"),
			expected: BulkString{Value: stringPtr("hello")},
			trailing: []byte{},
			hasError: true,
		},
		{
			name:     "missing $ prefix",
			input:    []byte("$world\r\n"),
			expected: BulkString{Value: stringPtr("world")},
			trailing: []byte{},
			hasError: true,
		},
		{
			name:     "incorrect format",
			input:    []byte("+incorrectformat"),
			expected: BulkString{Value: stringPtr("")},
			trailing: []byte{},
			hasError: true,
		},
		{
			name:     "incorrect format",
			input:    []byte("$\r\nincorrectformat"),
			expected: BulkString{Value: stringPtr("")},
			trailing: []byte{},
			hasError: true,
		},
		{
			name:     "valid bulk string",
			input:    []byte("$5\r\nhello\r\n"),
			expected: BulkString{Value: stringPtr("hello")},
			trailing: []byte{},
			hasError: false,
		},
		{
			name:     "valid bulk string 2",
			input:    []byte("$11\r\nhello-world\r\n"),
			expected: BulkString{Value: stringPtr("hello-world")},
			trailing: []byte{},
			hasError: false,
		},
		{
			name:     "valid bulk string3",
			input:    []byte("$12\r\nhello\r\nworld\r\n"),
			expected: BulkString{Value: stringPtr("hello\r\nworld")},
			trailing: []byte{},
			hasError: false,
		},
		{
			name:     "empty bulk string",
			input:    []byte("$0\r\n\r\n"),
			expected: BulkString{Value: stringPtr("")},
			trailing: []byte{},
			hasError: false,
		},
		{
			name:     "null bulk string",
			input:    []byte("$-1\r\n"),
			expected: BulkString{Value: nil},
			trailing: []byte{},
			hasError: false,
		},
		{
			name:     "input from array",
			input:    []byte("$5\r\nhello\r\n$-1\r\n$5\r\nworld\r\n"),
			expected: BulkString{Value: stringPtr("hello")},
			trailing: []byte("$-1\r\n$5\r\nworld\r\n"),
			hasError: false,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, trailingData, err := parseBulkString(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("expected an error for input %s, but got none", string(test.input))
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error for input %s, but got %v", string(test.input), err)
				}

				if result.Value != nil && test.expected.Value != nil && *result.Value != *test.expected.Value {
					t.Errorf("expected %v, but got %v", test.expected, result)
				}

				if result.Value == nil && test.expected.Value != nil || result.Value != nil && test.expected.Value == nil {
					t.Errorf("expected %v, but got %v", test.expected, result)
				}
			}

			if string(trailingData) != string(test.trailing) {
				t.Errorf("expected trailing data %v, but got %v", test.trailing, trailingData)
			}
		})
	}
}
func TestParseArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Array
		hasError bool
	}{

		{
			name:  "1 3 bulk strings in array second is nil",
			input: []byte("*3\r\n$5\r\nhello\r\n$-1\r\n$5\r\nworld\r\n"),
			expected: Array{
				Elements: &[]RESPData{
					BulkString{Value: stringPtr("hello")},
					BulkString{Value: nil}, // Representing a null bulk string
					BulkString{Value: stringPtr("world")},
				},
			},
			hasError: false,
		},
		{
			name:     "2  exp nil array",
			input:    []byte("*-1\r\n"),    // Null array
			expected: Array{Elements: nil}, // Nil array expected
			hasError: false,
		},
		{
			name:  "3 exp empty array",
			input: []byte("*0\r\n"), // Empty array
			expected: Array{
				Elements: &[]RESPData{},
			},
			hasError: false,
		},
		{
			name:     "4 incorrect prefix",
			input:    []byte("+hello\r\n"), // Incorrect prefix
			expected: Array{Elements: nil},
			hasError: true,
		},

		{
			name:     "5 incorrect format",
			input:    []byte("*\r\n"), // Incorrect format
			expected: Array{Elements: nil},
			hasError: true,
		},
		{
			name:     "6 incorrect format",
			input:    []byte("*\r\n$5\r\nhello\r\n"), // Incorrect format
			expected: Array{Elements: nil},
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			result, _, err := parseArray(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("expected an error for input %s, but got none", string(test.input))
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error for input %s, but got %v", string(test.input), err)
				}

				if test.expected.Elements == nil || result.Elements == nil {
					if test.expected.Elements != nil && result.Elements != nil {
						t.Errorf("expected %v, but got %v", test.expected, result)
					} else {
						if test.expected.Elements == nil && result.Elements != nil || test.expected.Elements != nil && result.Elements == nil {
							t.Errorf("expected %v, but got %v", test.expected, result)
						}
					}

				} else {

					if !reflect.DeepEqual(*result.Elements, *test.expected.Elements) {
						t.Errorf("expected %v, but got %v", *test.expected.Elements, *result.Elements)
					}
				}

			}
		})
	}
}

// //Helpers

func stringPtr(s string) *string {
	return &s
}
