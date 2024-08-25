package resp


type RESPData interface{
	serialize() ([]byte, error)
}

type SimpleString struct {
	Value string
}

func (s SimpleString) serialize() ([]byte, error) {
	return SerializeSimpleString(s)
}

type SimpleError struct {
	Message string
}

func (s SimpleError) serialize() ([]byte, error) {
	return SerializeSimpleError(s)
}


type Integer struct {
	Value int
}

func (i Integer) serialize() ([]byte, error) {
	return SerializeInteger(i)
}



type BulkString struct {
	Value *string
}

func (b BulkString) serialize() ([]byte, error) {
	return SerializeBulkString(b)
}




type Array struct {
	Elements *[]RESPData
}

func (a Array) serialize() ([]byte, error) {
	return SerializeArray(a)
}



