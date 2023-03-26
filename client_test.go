package form3client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	f3Client "form3-client-library"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	baseURL = "http://localhost:8080/v1/organisation/accounts"
)

func TestFetch_WhenEmptyIDs_ThenFailsWithBadRequest(t *testing.T) {
	c := f3Client.NewClient()

	tests := []string{
		"",
		" ",
		"  ",
		"   ",
	}

	for _, test := range tests {
		account, err := c.Fetch(context.Background(), test)

		assert.EqualError(t, err, f3Client.ErrRequiredID.Error())
		assert.Equal(t, account, f3Client.Account{})
	}
}

func TestFetch_WhenDoError_ThenFailsWithErr(t *testing.T) {
	var (
		expErr = f3Client.ErrClientInternal

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			return nil, errors.New("client internal error")
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := client.Fetch(context.Background(), "test-id")

	assert.Equal(t, f3Client.Account{}, account)
	assert.EqualError(t, err, expErr.Error())
}

func TestFetch_WhenRequestAndUnmarshalError_ThenFailsWithNormalizedErr(t *testing.T) {
	var (
		sendReqURL    string
		testAccountID = uuid.NewString()

		expReqURL    = fmt.Sprintf("%s/%s", baseURL, testAccountID)
		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       getReaderFromInterface(-1),
			}, nil
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := client.Fetch(context.Background(), testAccountID)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, f3Client.Account{}, account)
	assert.EqualError(t, err, f3Client.ErrUnmarshalInvalidValue.Error())
}

func TestFetch_WhenRecordNotFound_ThenFailsWithNormalizedErr(t *testing.T) {
	var (
		sendReqURL string
		testID     = uuid.NewString()

		expReqURL = fmt.Sprintf("%s/%s", baseURL, testID)
		expErr    = f3Client.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        f3Client.ErrRecordNotFound,
		}

		apiErr       = f3Client.ResponseError{ErrorMessage: fmt.Sprintf("record %s does not exist", testID)}
		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       getReaderFromInterface(apiErr),
			}, nil
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := client.Fetch(context.Background(), testID)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, f3Client.Account{}, account)
	assert.EqualError(t, err, expErr.Error())
}

func TestFetch_WhenUnmarshalError_ThenFailsWithNormalizedErr(t *testing.T) {
	var (
		sendReqURL    string
		testAccountID = uuid.NewString()

		expReqURL    = fmt.Sprintf("%s/%s", baseURL, testAccountID)
		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       getReaderFromInterface(-1),
			}, nil
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := client.Fetch(context.Background(), testAccountID)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, f3Client.Account{}, account)
	assert.EqualError(t, err, f3Client.ErrUnmarshalInvalidValue.Error())
}

func TestFetch_WhenAccountFound_ThenSucceessWithAccountDetail(t *testing.T) {
	var (
		sendReqURL    string
		testAccountID = uuid.NewString()

		expReqURL  = fmt.Sprintf("%s/%s", baseURL, testAccountID)
		expAccount = f3Client.Account{
			ID:             testAccountID,
			OrganizationID: uuid.NewString(),
			Type:           "accounts",
			Version:        0,
			AccountAttributes: f3Client.AccountAttributes{
				Country:          "AR",
				Name:             []string{"name_test"},
				AlternativeNames: []string{"alternative_name_test"},
			},
		}

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: getReaderFromInterface(f3Client.AccountResponse{
					Account: expAccount,
					Links:   f3Client.Links{Self: sendReqURL},
				}),
			}, nil
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := client.Fetch(context.Background(), testAccountID)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, expAccount, account)
	assert.NoError(t, err)
}

func TestDelete_WhenEmptyIDs_ThenFailsWithBadRequest(t *testing.T) {
	c := f3Client.NewClient()

	tests := []string{
		"",
		" ",
		"  ",
		"   ",
	}

	for _, test := range tests {
		err := c.Delete(context.Background(), test)

		assert.EqualError(t, err, f3Client.ErrRequiredID.Error())
	}
}

func TestDelete_WhenDoError_ThenFailsWithErr(t *testing.T) {
	var (
		expErr = f3Client.ErrClientInternal

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			return nil, errors.New("client internal error")
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	err := client.Delete(context.Background(), "test-id")

	assert.EqualError(t, err, expErr.Error())
}

func TestDelete_WhenRecordNotFound_ThenFailsWithNormalizedErr(t *testing.T) {
	var (
		sendReqURL string
		testID     = uuid.NewString()

		expReqURL = fmt.Sprintf("%s/%s?version=0", baseURL, testID)
		expErr    = f3Client.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        f3Client.ErrRecordNotFound,
		}

		apiErr       = f3Client.ResponseError{ErrorMessage: fmt.Sprintf("record %s does not exist", testID)}
		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       getReaderFromInterface(apiErr),
			}, nil
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	err := client.Delete(context.Background(), testID)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.EqualError(t, err, expErr.Error())
}

func TestDelete_WhenRecordDeleted_ThenSuccessWithNoContentResponse(t *testing.T) {
	var (
		sendReqURL string
		testID     = uuid.NewString()

		expReqURL = fmt.Sprintf("%s/%s?version=0", baseURL, testID)

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusNoContent,
			}, nil
		}

		client = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	err := client.Delete(context.Background(), testID)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.NoError(t, err)
}

func TestCreate_WhenDoError_ThenFailsWithErr(t *testing.T) {
	var (
		sendReqURL string
		req        f3Client.AccountRequest

		expReqURL  = baseURL
		expAccount f3Client.Account
		expErr     = f3Client.ErrClientInternal

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return nil, errors.New("client internal error")
		}

		c = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := c.Create(context.TODO(), req)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, expAccount, account)
	assert.EqualError(t, err, expErr.Error())
}

func TestCreate_WhenBadRequest_ThenFailsWithNormalizedErr(t *testing.T) {
	var (
		sendReqURL string
		req        f3Client.AccountRequest

		expReqURL  = baseURL
		expAccount f3Client.Account
		expErr     = f3Client.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        errors.New("country in body is required;"),
		}

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body: getReaderFromInterface(f3Client.ResponseError{
					ErrorMessage: "validation failure list:\nvalidation failure list:" +
						"\nvalidation failure list:\ncountry in body is required",
				}),
			}, nil
		}

		c = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := c.Create(context.TODO(), req)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, expAccount, account)
	assert.EqualError(t, err, expErr.Error())
}

func TestCreate_WhenUnmarshalErr_ThenFailsWithNormalizedErr(t *testing.T) {
	var (
		sendReqURL string
		req        f3Client.AccountRequest

		expReqURL  = baseURL
		expAccount f3Client.Account

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       getReaderFromInterface(-1),
			}, nil
		}

		c = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := c.Create(context.TODO(), req)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, expAccount, account)
	assert.EqualError(t, err, f3Client.ErrUnmarshalInvalidValue.Error())
}

func TestCreate_WhenAccountCreated_ThenSuccess(t *testing.T) {
	var (
		sendReqURL string

		req = f3Client.AccountRequest{
			ID:             uuid.NewString(),
			OrganisationID: uuid.NewString(),
			Type:           "accounts",
			Version:        nil,
			Attributes: &f3Client.AccountAttributesRequest{
				Country:          "AR",
				Name:             []string{"name_test"},
				AlternativeNames: []string{"alternative_name_test"},
			},
		}

		accountID   = uuid.NewString()
		accountResp = f3Client.AccountResponse{
			Links: f3Client.Links{Self: fmt.Sprintf("%s/%s", baseURL, accountID)},
			Account: f3Client.Account{
				ID:             accountID,
				OrganizationID: req.OrganisationID,
				Type:           req.Type,
				Version:        0,
				AccountAttributes: f3Client.AccountAttributes{
					Country:          req.Attributes.Country,
					Name:             req.Attributes.Name,
					AlternativeNames: req.Attributes.AlternativeNames,
				},
			},
		}

		expReqURL  = baseURL
		expAccount = f3Client.Account{
			ID:             accountID,
			OrganizationID: accountResp.Account.OrganizationID,
			Type:           accountResp.Account.Type,
			Version:        accountResp.Account.Version,
			AccountAttributes: f3Client.AccountAttributes{
				Country:          accountResp.Account.AccountAttributes.Country,
				Name:             accountResp.Account.AccountAttributes.Name,
				AlternativeNames: accountResp.Account.AccountAttributes.AlternativeNames,
			},
		}

		doerMockFunc = func(client http.Client, req *http.Request) (resp *http.Response, err error) {
			sendReqURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       getReaderFromInterface(accountResp),
			}, nil
		}

		c = f3Client.NewClient(f3Client.MockDoer(doerMockFunc))
	)

	account, err := c.Create(context.TODO(), req)

	assert.Equal(t, expReqURL, sendReqURL)
	assert.Equal(t, expAccount, account)
	assert.NoError(t, err)
}

func getReaderFromInterface(i interface{}) io.ReadCloser {
	b, _ := json.Marshal(&i)
	return io.NopCloser(bytes.NewBufferString(string(b)))
}
