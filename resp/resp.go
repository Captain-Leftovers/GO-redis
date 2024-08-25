package resp

import (
	"fmt"
	"strings"
)

var commands = map[string]func(...BulkString) ([]byte, error){
	"ECHO": echo,
	"PING": ping,
}

func ping(args ...BulkString) ([]byte, error) {

	return []byte("+PONG\r\n"), nil
}

func echo(args ...BulkString) ([]byte, error) {
	if len(args) < 1 {
		return []byte(""), fmt.Errorf("missing argument for ECHO")
	}

	return args[0].serialize()
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

	f, ok := commands[cmdStr]

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	if len(*val.Elements) < 2 {
		return f()
	}

	arg, ok := (*val.Elements)[1].(BulkString)

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	return f(arg)

	// TODO : loop trough commands like ECHO case insensitive and maybe create a map of commands

}
