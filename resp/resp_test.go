package resp

import (
	"reflect"
	"testing"
)

func TestExecuteRespData(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{

		{
			name:     "Echo hey",
			input:    []byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"),
			expected: []byte("$3\r\nhey\r\n"),
		},
	}
	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			result, err := ExecuteRespData(test.input)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if test.expected != nil && !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, but got %v", test.expected, result)
			}

		})
	}
}
