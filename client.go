package form3client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	accountsPath = "/v1/organisation/accounts"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
)

// Doer provides the main interface to send an HTTP request and returns an HTTP response
// client. A mock library can be found under the file ./mocks with an implementation of
// a ClientOption configuration.
type Doer interface {
	Do(client http.Client, req *http.Request) (resp *http.Response, err error)
}

// Client implements a simple wrapper around the Go standard package
// http.Client.
//
// To use it, create an instance with NewClient, the zero value of this Client
// is not safe to use.
type Client struct {
	client  http.Client
	doer    Doer
	baseURL string
}

// NewClient is the only way to properly instantiate a Form3 Client.
//
// The NewClient feature accepts ClientOptions that will allow
// the proper setting of the client configurations. For more
// information please visit ./client_options file.
//
// Default settings will be applied if no options are injected, and
// only selected will be applied leaving the rest as default.
func NewClient(options ...ClientOption) Client {
	client := Client{
		client: http.Client{},
	}

	var defaultOptions = []ClientOption{
		BaseURL("http://accountapi:8080"),
		Timeout(2 * time.Second),
		Retries(2, 2250, 150),
	}

	options = append(defaultOptions, options...)

	for _, option := range options {
		client = option(client)
	}

	return client
}

// Fetch allows to retrieve an account resource by its identifier providing:
//
// ctx (context.Context) context carries a deadline, a cancellation signal, and other values across API boundaries.
//
// id  (string) identifier of the Form3 account to fetch.
//
// If no id is provided RequestErr with the ErrRequiredID is returned and it can't
// contain only blanks.
//
// Errors related to the request or resource trying to be obtained will be of type
// RequestError, while server side errors will be of type error.
func (c *Client) Fetch(ctx context.Context, id string) (Account, error) {
	if containsOnlyBlanks(id) {
		return Account{}, ErrRequiredID
	}

	req, err := c.makeJSONRequest(ctx, http.MethodGet, fmt.Sprintf("%s/%s", accountsPath, id), nil)
	if err != nil {
		return Account{}, err
	}

	resp, err := c.doer.Do(c.client, req)
	if err != nil {
		return Account{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Account{}, handleResponseError(resp)
	}

	var rData AccountResponse
	if err := unmarshalBody(resp.Body, &rData); err != nil {
		return Account{}, err
	}

	return rData.Account, nil
}

// Delete allows to remove an account by its identifier providing:
//
// ctx (context.Context) context carries a deadline, a cancellation signal, and other values across API boundaries.
//
// id  (string) identifier of the Form3 account to delete.
//
// If no id is provided RequestErr with the ErrRequiredID is returned and it can't
// contain only blanks.
//
// Errors related to the request or resource trying to be deleted will be of type
// RequestError, while server side errors will be of type error.
func (c *Client) Delete(ctx context.Context, id string) error {
	if containsOnlyBlanks(id) {
		return ErrRequiredID
	}

	req, err := c.makeJSONRequest(ctx, http.MethodDelete, fmt.Sprintf("%s/%s?version=0", accountsPath, id), nil)
	if err != nil {
		return err
	}

	resp, err := c.doer.Do(c.client, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return handleResponseError(resp)
	}

	return nil
}

// Create allows to register a new account resource:
//
// ctx (context.Context) context carries a deadline, a cancellation signal, and other values across API boundaries.
//
// account  (AccountRequest) account contains the basic and optional data for the register of a new account.
//
// If the resource is successfully created the func will return an (Account) object with the base information.
//
// Errors related to the request  will be of type
// RequestError, while server side errors will be of type error.
func (c *Client) Create(ctx context.Context, account AccountRequest) (Account, error) {
	req, err := c.makeJSONRequest(ctx, http.MethodPost, accountsPath, CreateAccountRequest{account})
	if err != nil {
		return Account{}, err
	}

	resp, err := c.doer.Do(c.client, req)
	if err != nil {
		return Account{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Account{}, handleResponseError(resp)
	}

	var rData AccountResponse
	if err := unmarshalBody(resp.Body, &rData); err != nil {
		return Account{}, err
	}

	return rData.Account, nil
}

func (c *Client) resolveURL(path string) (*url.URL, error) {
	return url.Parse(c.baseURL + path)
}

func containsOnlyBlanks(value string) bool {
	return len(strings.TrimSpace(value)) == 0
}

func (c *Client) makeJSONRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, ErrSerializeRequest
	}

	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.URL, err = c.resolveURL(path)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add(contentTypeHeader, jsonContentType)
	}

	return req, nil
}

func unmarshalBody(body io.Reader, value any) error {
	content, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		return nil
	}

	if err := json.Unmarshal(content, &value); err != nil {
		return ErrUnmarshalInvalidValue
	}

	return nil
}
