package form3client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "http://localhost:8080/v1/organisation/accounts"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
)

// Doer provides the main intention to send an HTTP request and returns an HTTP response
// client. A mock library can be found under the file ./mocks with an implementation of
// a ClientOption configuration.
type Doer interface {
	Do(client http.Client, req *http.Request) (resp *http.Response, err error)
}

type Client struct {
	client http.Client
	doer   Doer
}

func NewClient(options ...ClientOption) Client {
	c := Client{
		client: http.Client{},
	}

	var defaultOptions = []ClientOption{
		Timeout(2 * time.Second),
		Retries(2, 2250, 150),
	}

	options = append(defaultOptions, options...)

	for _, option := range options {
		c = option(c)
	}

	return c
}

func (c *Client) Fetch(ctx context.Context, id string) (Account, error) {
	if containsOnlyBlanks(id) {
		return Account{}, ErrRequiredID
	}

	path := fmt.Sprintf("%s/%s", baseURL, id)
	req, err := makeJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return Account{}, err
	}

	resp, err := c.doer.Do(c.client, req)
	if err != nil {
		return Account{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Account{}, normalizeResponseError(resp)
	}

	var rData AccountResponse
	if err := unmarshalBody(resp.Body, &rData); err != nil {
		return Account{}, err
	}

	return rData.Account, nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	if containsOnlyBlanks(id) {
		return ErrRequiredID
	}

	path := fmt.Sprintf("%s/%s?version=0", baseURL, id)
	req, err := makeJSONRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	resp, err := c.doer.Do(c.client, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return normalizeResponseError(resp)
	}

	return nil
}

func (c *Client) Create(ctx context.Context, account AccountRequest) (Account, error) {
	req, err := makeJSONRequest(http.MethodPost, baseURL, account)
	if err != nil {
		return Account{}, err
	}

	req.Header.Add(contentTypeHeader, jsonContentType)

	resp, err := c.doer.Do(c.client, req)
	if err != nil {
		return Account{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		return Account{}, normalizeResponseError(resp)
	}

	var rData AccountResponse
	if err := unmarshalBody(resp.Body, &rData); err != nil {
		return Account{}, err
	}

	return rData.Account, nil
}

func containsOnlyBlanks(value string) bool {
	return len(strings.TrimSpace(value)) == 0
}

func normalizeResponseError(resp *http.Response) error {
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

func makeJSONRequest(method, path string, body interface{}) (*http.Request, error) {
	content, err := json.Marshal(body)
	if err != nil {
		return nil, ErrSerializeRequest
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add(contentTypeHeader, jsonContentType)
	}

	return req, nil
}

func unmarshalBody(body io.Reader, v any) error {
	content, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		return nil
	}

	if err := json.Unmarshal(content, &v); err != nil {
		return ErrUnmarshalInvalidValue
	}

	return nil
}
