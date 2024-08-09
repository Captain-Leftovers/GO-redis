package resp

type RESPData interface{}

type SimpleString struct {
	Value string
}

type BulkString struct {
	Value string
}

type Integer struct {
	Value int
}

type Array struct {
	Elements []RESPData
}

type SimpleError struct {
	Message string
}
