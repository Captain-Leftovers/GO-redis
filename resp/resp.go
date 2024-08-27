package resp

import (
	"fmt"
	"strings"
)

var store = make(map[string]string)

var commands = map[string]func(...BulkString) ([]byte, error){
	"ECHO": handleEcho,
	"PING": handlePing,
	"SET":  handleSet,
	"GET":  handleGet,
}

func handleSet(args ...BulkString) ([]byte, error) {
	key := *args[1].Value
	value := *args[2].Value

	store[key] = value
	return []byte("+OK\r\n"), nil
}

func handleGet(args ...BulkString) ([]byte, error) {
	key := *args[1].Value

	if value, exists := store[key]; exists {
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)), nil
	}
	return []byte("$-1\r\n"), nil
}

func handlePing(args ...BulkString) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}

func handleEcho(args ...BulkString) ([]byte, error) {
	if len(args) < 2 {
		return []byte(""), fmt.Errorf("missing argument for ECHO")
	}

	return args[1].serialize()
}

// parses the RESP data checks for commands and returns serialized response string and error if any
func ExecuteRespData(data []byte) ([]byte, error) {
	respData, _, err := ParseByteDataToResp(data)
	if err != nil {
		return []byte(""), err
	}

	val, ok := respData.(Array)
	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	// TODO : see if we can use more than 2 elements
	if len(*val.Elements) < 1 {
		return []byte(""), fmt.Errorf("invalid command")
	}

	cmd, ok := (*val.Elements)[0].(BulkString)

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	cmdStr := strings.ToUpper(*cmd.Value)

	commandFunc, ok := commands[cmdStr]

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	if len(*val.Elements) < 2 {
		return commandFunc()
	}


	args := make([]BulkString, len(*val.Elements))

	for i, elem := range *val.Elements {
		arg, ok := elem.(BulkString)

		if !ok {
			return []byte(""), fmt.Errorf("invalid command")
		}

		args[i] = arg
	}


	// arg, ok := (*val.Elements)[1].(BulkString)

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	return commandFunc(args...)

}
