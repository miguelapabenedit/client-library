package mocks

import (
	"net/http"
)

type DoerMock struct {
	doMockFn func(client http.Client, req *http.Request) (resp *http.Response, err error)
}

func NewDoerMock(doMockFn func(client http.Client, req *http.Request) (resp *http.Response, err error)) *DoerMock {
	return &DoerMock{
		doMockFn,
	}
}

func (rm *DoerMock) Do(client http.Client, req *http.Request) (resp *http.Response, err error) {
	if rm != nil && rm.doMockFn != nil {
		return rm.doMockFn(client, req)
	}

	return &http.Response{StatusCode: http.StatusOK}, nil
}
