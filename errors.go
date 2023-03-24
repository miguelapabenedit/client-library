package form3client

import (
	"errors"
	"fmt"
	"net/http"
)

type RequestError struct {
	StatusCode   int
	ErrorMessage string
}

func (re RequestError) Error() string {
	return fmt.Sprintf("status:%d, error:'%s'", re.StatusCode, re.ErrorMessage)
}

var (
	// ErrUnmarshalInvalidValue signals that the recieved JSON value
	// could not be converted to its selected interface using json.Unmarshal.
	ErrUnmarshalInvalidValue = errors.New("invalid unmarshal JSON value")

	// ErrRetryRequest signals the failure to execute the request and the
	// limit of retry attempts was reached.
	ErrRetryRequest = errors.New("unable to execute request, retry attempts reached")

	ErrSerializeRequest = errors.New("an error happend while trying to deserialize")

	ErrRequiredID = RequestError{
		StatusCode:   http.StatusBadRequest,
		ErrorMessage: "an id must be provided and can`t contain only blanks",
	}

	ErrRecordNotFound = "record does not exist"
)
