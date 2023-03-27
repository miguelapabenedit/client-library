package mocks

import (
	"net/http"
)

type doerMock struct {
	doMockFunc func(client http.Client, req *http.Request) (resp *http.Response, err error)
}

// NewDoerMock returns an mock implementation of the interface Doer that allows to
// mock the Do func for testing purposes.
//
// doMockFunc func(client http.Client, req *http.Request) (resp *http.Response, err error) mock function
//
// In case of a nil func, the Do func will return a happy path with status http.StatusOK and nil error
func NewDoerMock(doMockFunc func(client http.Client, req *http.Request) (resp *http.Response, err error)) *doerMock {
	return &doerMock{
		doMockFunc,
	}
}

func (rm *doerMock) Do(client http.Client, req *http.Request) (resp *http.Response, err error) {
	if rm != nil && rm.doMockFunc != nil {
		return rm.doMockFunc(client, req)
	}

	return &http.Response{StatusCode: http.StatusOK}, nil
}
