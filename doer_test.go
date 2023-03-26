package form3client_test

import (
	"errors"
	f3Client "form3-client-library"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDo_WhenRequestSuccess_ThenSuccessWithNilErr(t *testing.T) {
	expStatusCode := http.StatusCreated

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expStatusCode)
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	doer := f3Client.NewRetryDoer(0, 0, 0)

	resp, err := doer.Do(http.Client{}, req)

	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, expStatusCode)
}

func TestDo_WhenInitDoerWithEmptyValuesAndRequestErr_ThenReturnsErrWithNoRetry(t *testing.T) {
	expErr := url.Error{Op: "Get", URL: "url", Err: errors.New(`unsupported protocol scheme ""`)}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, "url", nil)
	doer := f3Client.NewRetryDoer(1, 0, 4)

	resp, err := doer.Do(http.Client{}, req)

	assert.EqualError(t, err, expErr.Error())
	assert.Nil(t, resp)
}

func TestDo_WhenRequestErrAndRetryLimitReached_ThenFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	doer := f3Client.NewRetryDoer(2, 250, 300)

	resp, err := doer.Do(http.Client{}, req)

	assert.EqualError(t, err, f3Client.ErrRetryRequest.Error())
	assert.Nil(t, resp)
}

func TestDo_WhenRequestSuccessWithRetry_ThenReturnsResp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	doer := f3Client.NewRetryDoer(2, 250, 300)

	_, err := doer.Do(http.Client{}, req)

	assert.NoError(t, err)
}
