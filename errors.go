package form3client

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type RequestError struct {
	StatusCode int
	Err        error
}

func (re RequestError) Error() string {
	return fmt.Sprintf("status:%d, error:'%s'.", re.StatusCode, re.Err.Error())
}

var (
	// ErrUnmarshalInvalidValue signals that the recieved JSON value
	// could not be converted to its selected interface using json.Unmarshal.
	ErrUnmarshalInvalidValue = errors.New("invalid unmarshal JSON value.")

	// ErrRetryRequest signals the failure to execute the request and the
	// limit of retry attempts was reached.
	ErrRetryRequest = errors.New("unable to execute request, retry attempts reached.")

	// ErrSerializeRequest signals the failure while trying to encode an object with
	// json.Marshal operation.
	ErrSerializeRequest = errors.New("an error happend while trying to serialize.")

	// ErrRequiredID signals that in order to continue the user must provide and id that doesn't
	// contain only blanks.
	ErrRequiredID = RequestError{
		StatusCode: http.StatusBadRequest,
		Err:        errors.New("an id must be provided and can't contain only blanks."),
	}

	// ErrTimeout signals that the request was cancel do to reach the specified timeout,
	// client timeout can be change withe de ClientOption -> Timeout() and
	// for more settings information review client options.
	ErrTimeout = errors.New("request cancel due to context timeout deadline exceeded")

	// ErrRecordNotFound signals that the requested resource is not available or does not
	// exist.
	ErrRecordNotFound = errors.New("record does not exist.")

	// ErrClientInternal signals that during the execution of the request an error
	// was encounter, preventing for any retry attempts.
	ErrClientInternal = errors.New("client doer internal error")
)

func handleClientError(err error) error {
	if os.IsTimeout(err) {
		return ErrTimeout
	}

	return ErrClientInternal
}

func handleResponseError(resp *http.Response) error {
	var respErr ResponseError
	if err := unmarshalBody(resp.Body, &respErr); err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		var (
			errors    = strings.Split(respErr.ErrorMessage, "\n")
			errBuffer bytes.Buffer
		)

		for _, err := range errors {
			if !strings.EqualFold(err, "validation failure list:") {
				errBuffer.WriteString(fmt.Sprintf("%s;", err))
			}
		}

		respErr.ErrorMessage = errBuffer.String()
	case http.StatusNotFound:
		respErr.ErrorMessage = ErrRecordNotFound.Error()
	}

	return RequestError{Err: errors.New(respErr.ErrorMessage), StatusCode: resp.StatusCode}
}
