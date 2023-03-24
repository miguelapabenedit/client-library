package form3client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
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

type Doer interface

type Client struct {
	client         http.Client

	retryAttempts  int
	backoffIntvl   int
	maxJitterIntvl int
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
	resp, err := c.do(ctx, http.MethodGet, path, nil)
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
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return normalizeResponseError(resp)
	}

	return nil
}

func containsOnlyBlanks(value string) bool {
	return len(strings.TrimSpace(value)) == 0
}

func (c *Client) Create(ctx context.Context, account AccountRequest) (Account, error) {
	content, err := json.Marshal(CreateAccountRequest{account})
	if err != nil {
		return Account{}, ErrSerializeRequest
	}

	resp, err := c.do(ctx, http.MethodPost, baseURL, bytes.NewReader(content))
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

func (c *Client) do(ctx context.Context, method string, path string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, body)
	if err != nil {
		return resp, err
	}

	if body != nil {
		req.Header.Add(contentTypeHeader, jsonContentType)
	}

	retries := 0
	for ; retries <= c.retryAttempts; retries++ {
		if retries > 0 {
			backoffIntvl := int(float64(c.backoffIntvl)*math.Exp2(float64(retries))) + rand.Intn(c.maxJitterIntvl)
			time.Sleep(time.Duration(float64(backoffIntvl)))
		}

		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
	}

	if retries >= c.retryAttempts {
		return resp, ErrRetryRequest
	}

	return resp, err
}

func normalizeResponseError(resp *http.Response) error {
	var respErr ResponseError
	if err := unmarshalBody(resp.Body, &respErr); err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		var (
			required = strings.Split(respErr.ErrorMessage, "\n")
			errMsg   bytes.Buffer
		)

		for _, msg := range required {
			if !strings.EqualFold(msg, "validation failure list:") {
				errMsg.WriteString(msg)
				errMsg.WriteString(";")
			}
		}
		respErr.ErrorMessage = errMsg.String()

	case http.StatusNotFound:
		respErr.ErrorMessage = ErrRecordNotFound
	}

	return RequestError{ErrorMessage: respErr.ErrorMessage, StatusCode: resp.StatusCode}
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
