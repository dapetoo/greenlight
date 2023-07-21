package data

import (
	"fmt"
	"strconv"
)

// Runtime Custom Runtime type with underlying type int32
type Runtime int32

// MarshalJSON Implement a MarshalJSON method on the Runtime type
func (r Runtime) MarshalJSON() ([]byte, error) {
	//Generate a string containing the movie runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)

	//strconv.Quote function on the string to wrap it in double quotes.
	quotedJSONValue := strconv.Quote(jsonValue)

	//Convert the quoted string value to a byte slice and return it
	return []byte(quotedJSONValue), nil
}
