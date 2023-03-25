package form3client

import (
	"errors"
	"fmt"
	"net/http"
)

type RequestError struct {
	StatusCode int
	Err        error
}

func (re RequestError) Error() string {
	return fmt.Sprintf("status:%d, error:'%s'", re.StatusCode, re.Err.Error())
}

var (
	// ErrUnmarshalInvalidValue signals that the recieved JSON value
	// could not be converted to its selected interface using json.Unmarshal.
	ErrUnmarshalInvalidValue = errors.New("invalid unmarshal JSON value")

	// ErrRetryRequest signals the failure to execute the request and the
	// limit of retry attempts was reached.
	ErrRetryRequest = errors.New("unable to execute request, retry attempts reached")

	ErrSerializeRequest = errors.New("an error happend while trying to serialize")

	ErrRequiredID = RequestError{
		StatusCode: http.StatusBadRequest,
		Err:        errors.New("an id must be provided and can't contain only blanks"),
	}

	ErrRecordNotFound = errors.New("record does not exist")

	// ErrInvalidQueryParametersCount signals that Get didn't receive an even
	// number of parameters (0 is even), those parameters are to generate the
	// query on the GET url so they need to have the form: key1, value1, key2,
	// value2, key3, value3
	ErrInvalidQueryParametersCount = errors.New("invalid size of query parameters, only an even number is accepted")
)
