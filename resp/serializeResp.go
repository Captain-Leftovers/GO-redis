package resp

import (
	"fmt"
)

func SerializeRESPDataToBytes(data RESPData) ([]byte, error) {
	switch v := data.(type) {
	case SimpleString:
		return SerializeSimpleString(v)
	case SimpleError:
		return SerializeSimpleError(v)
	case Integer:
		return SerializeInteger(v)
	case BulkString:
		return SerializeBulkString(v)
	case Array:
		return SerializeArray(v)
	default:
		return nil, fmt.Errorf("unknown RESP type")
	}
}

func SerializeSimpleString(s SimpleString) ([]byte, error) {
	return []byte(fmt.Sprintf("+%s\r\n", s.Value)), nil
}

// TODO : add tests and test per function

func SerializeSimpleError(s SimpleError) ([]byte, error) {
	return []byte(fmt.Sprintf("-%s\r\n", s.Message)), nil
}

func SerializeInteger(i Integer) ([]byte, error) {
	return []byte(fmt.Sprintf(":%d\r\n", i.Value)), nil
}

func SerializeBulkString(b BulkString) ([]byte, error) {
	if b.Value == nil {
		return []byte("$-1\r\n"), nil
	}
	str := *b.Value
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)), nil
}

func SerializeArray(a Array) ([]byte, error) {
    if a.Elements == nil {
        return []byte("*-1\r\n"), nil
    }
    var result []byte
    result = append(result, fmt.Sprintf("*%d\r\n", len(*a.Elements))...)
    for _, elem := range *a.Elements {
        b, err := SerializeRESPDataToBytes(elem)
        if err != nil {
            return nil, err
        }
        result = append(result, b...)
    }
    fmt.Println("result is ", string(result))
    return result, nil
}