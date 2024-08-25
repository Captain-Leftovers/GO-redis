package resp

import (
	"bytes"
	"fmt"
	"strconv"
)

var CRLF = "\r\n"

// RESPData is an interface that all RESP types implement
// returns the type of the RESPData, and the remaining unprocessed data as []byte, and an error if any
func ParseByteDataToResp(data []byte) (RESPData, []byte, error) {

	prefix := string(data[0])

	switch prefix {
	case "+":
		return parseSimpleString(data)
	case "-":
		return parseSimpleError(data)
	case ":":
		return parseInteger(data)
	case "$":
		return parseBulkString(data)
	case "*":
		return parseArray(data)
	default:
		return nil, []byte{}, fmt.Errorf("unknown RESP type: %s", prefix)
	}

}

func parseSimpleString(data []byte) (SimpleString, []byte, error) {
	result := ""
	prefix := "+"

	startIndex := bytes.Index(data, []byte(prefix))

	endIndex := bytes.Index(data, []byte(CRLF))

	if endIndex == -1 {
		return SimpleString{}, []byte{}, fmt.Errorf("incorrect/missing termination of simpleString")
	}

	trailingData := data[endIndex+2:]

	if startIndex != -1 && endIndex != -1 && endIndex > startIndex {

		result = string(data[startIndex+1 : endIndex])

		return SimpleString{Value: result}, trailingData, nil
	} else {

		return SimpleString{}, trailingData, fmt.Errorf("invalid simple string format")
	}

}

func parseSimpleError(data []byte) (SimpleError, []byte, error) {

	prefix := "-"

	startIndex := bytes.Index(data, []byte(prefix))
	endIndex := bytes.Index(data, []byte(CRLF))

	trailingData := data[endIndex+2:]

	if startIndex != -1 && endIndex != -1 && endIndex > startIndex+1 {
		errorMessage := data[startIndex+1 : endIndex]

		return SimpleError{Message: string(errorMessage)}, trailingData, nil
	} else {
		return SimpleError{}, trailingData, fmt.Errorf("invalid 'simple error' format")
	}

}

func parseInteger(data []byte) (Integer, []byte, error) {
	prefix := ":"

	startIndex := bytes.Index(data, []byte(prefix))
	endIndex := bytes.Index(data, []byte(CRLF))

	trailingData := data[endIndex+2:]

	if startIndex != -1 && endIndex != -1 && endIndex > startIndex+1 {
		result := string(data[startIndex+1 : endIndex])
		val, err := strconv.Atoi(result)

		if err != nil {
			return Integer{}, trailingData, fmt.Errorf("error while converting %s to int", result)
		}

		return Integer{Value: val}, trailingData, nil

	} else {
		return Integer{}, trailingData, fmt.Errorf("invalid 'Integers RESP' format")
	}
}

func parseBulkString(data []byte) (BulkString, []byte, error) {

	prefix := "$"

	prefixIndex := bytes.Index(data, []byte(prefix))

	if prefixIndex != 0 {
		return BulkString{}, []byte{}, fmt.Errorf("incorrect prefix got '%v' but expected '%s'", string(data[0]), prefix)
	}

	firstCRLF := bytes.Index(data, []byte(CRLF))

	if prefixIndex == -1 || firstCRLF == -1 || firstCRLF < prefixIndex {
		return BulkString{}, []byte{}, fmt.Errorf("incorrect 'Bulk string' format")
	}

	numBytes, err := strconv.Atoi(string(data[prefixIndex+1 : firstCRLF]))
	if err != nil {
		return BulkString{}, []byte{}, fmt.Errorf("incorrect 'Bulk string' format")
	}

	if numBytes == -1 && firstCRLF != -1 && firstCRLF == 3 {
		return BulkString{Value: nil}, data[5:], nil
	}

	if numBytes == -1 && firstCRLF == -1 {
		return BulkString{}, []byte{}, fmt.Errorf("malformed bulk string")
	}

	remData := data[firstCRLF+2:]
	nextCRLF := bytes.Index(remData[numBytes:], []byte(CRLF)) + numBytes

	if numBytes != nextCRLF {

		return BulkString{}, []byte{}, fmt.Errorf("incorrect 'Bulk string' format")
	}

	trailingData := remData[nextCRLF+2:]

	result := string(remData[:nextCRLF])

	return BulkString{Value: &result}, trailingData, nil

}

func parseArray(data []byte) (Array, []byte, error) {

	prefix := "*"

	prefixIndex := bytes.Index(data, []byte(prefix))

	if prefixIndex != 0 {
		return Array{Elements: nil}, []byte{}, fmt.Errorf("incorrect prefix got '%v' on index '%d' but expected '%s' on index '0'", string(data[0]), prefixIndex, prefix)
	}

	firstCRLF := bytes.Index(data, []byte(CRLF))
	if firstCRLF == -1 {
		return Array{Elements: nil}, []byte{}, fmt.Errorf("incorrect data passed to parse Array")
	}

	strNum := string(data[prefixIndex+1 : firstCRLF])

	num, err := strconv.Atoi(strNum)
	if err != nil {
		return Array{Elements: nil}, []byte{}, fmt.Errorf("cannot convert %s to number", strNum)
	}

	emptyCRLF := bytes.Index(data, []byte(CRLF))

	if emptyCRLF == -1 {
		return Array{Elements: nil}, []byte{}, fmt.Errorf("malformed resp Array")
	}

	if num == 0 && len(data[:emptyCRLF]) == 2 {
		return Array{Elements: &[]RESPData{}}, data[emptyCRLF+2:], nil
	}

	if num == -1 && len(data[:emptyCRLF]) == 3 {
		return Array{Elements: nil}, data[emptyCRLF+2:], nil
	}

	result := Array{
		&[]RESPData{},
	}

	trailingData := []byte{}

	for i, currIndex := 0, firstCRLF; i < num; i++ {

		nextCRLF := bytes.Index(data[currIndex+2:], []byte(CRLF))

		if nextCRLF == -1 {
			return Array{Elements: nil}, []byte{}, fmt.Errorf("malformed array, missing CRLF after element %d", i)
		}

		respElement := data[currIndex+2:]

		res, trailingData, err := ParseByteDataToResp(respElement)


		if err != nil {
			return Array{Elements: nil}, trailingData, fmt.Errorf("%v", err)
		}

		*result.Elements = append(*result.Elements, res)

		currIndex += (len(respElement) - len(trailingData))

	}

	return result, trailingData, nil
}
