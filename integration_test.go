package form3client_test

import (
	"context"
	"errors"
	f3Client "form3-client-library"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_BasicSettingFuncs runs the basic test helper create,delete and fetch funcs used
// in the integration tests. If any non contemplated error is encountered
// in the execution or data is not clear,the flow will panic preventing any further tests.
func TestIntegration_BasicSettingFuncs(t *testing.T) {
	client := f3Client.NewClient()

	accountID := createTestAccount(t, &client).ID
	cleanTestAccounts(t, &client, accountID)

	account := fetchTestAccountByID(t, &client, accountID)
	if account.ID != "" {
		t.Fatal("test data was not clear")
	}
}

func TestIntegrFetchTimeout_WhenDeadlineReached_ThenReturnTimeoutErr(t *testing.T) {
	var (
		id     = uuid.NewString()
		client = f3Client.NewClient(f3Client.Timeout(1 * time.Microsecond))
	)

	account, err := client.Fetch(context.Background(), id)

	assert.Empty(t, account)
	assert.ErrorContains(t, err, "context deadline exceeded (Client.Timeout exceeded while awaiting headers)")
}

func TestIntegrFetch_WhenInvalidUUID_ThenReturnBadRequest(t *testing.T) {
	var (
		expErr = f3Client.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        errors.New("id is not a valid uuid;"),
		}

		client = f3Client.NewClient()

		account, err = client.Fetch(context.Background(), "invalid_uuid")
	)

	assert.Empty(t, account)
	assert.Error(t, err)
	assert.EqualValues(t, err, expErr)
}

func TestIntegrFetch_WhenResourceNotFound_ThenReturnBadRequest(t *testing.T) {
	var (
		expErr = f3Client.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        f3Client.ErrRecordNotFound,
		}

		client = f3Client.NewClient()
	)

	account, err := client.Fetch(context.Background(), uuid.NewString())

	assert.Empty(t, account)
	assert.Error(t, err)
	assert.EqualValues(t, err, expErr)
}

func TestIntegrFetch_WhenClientErr_ThenReturnGenericErr(t *testing.T) {
	expErr := errors.New("client_err")
	doerMockFunc := func(client http.Client, req *http.Request) (resp *http.Response, err error) {
		return &http.Response{}, expErr
	}
	client := f3Client.NewClient(f3Client.MockDoer(doerMockFunc))

	account, err := client.Fetch(context.Background(), uuid.NewString())

	assert.Empty(t, account)
	assert.EqualError(t, err, expErr.Error())
}

func TestIntegrFetch_WhenResoruceFound_ThenReturnAccount(t *testing.T) {
	client := f3Client.NewClient()

	expAccount := createTestAccount(t, &client)
	defer cleanTestAccounts(t, &client, expAccount.ID)

	account, err := client.Fetch(context.Background(), expAccount.ID)

	assert.Equal(t, expAccount, account)
	assert.NoError(t, err)
}

func TestIntegrDeleteTimeout_WhenDeadlineReached_ThenReturnTimeoutErr(t *testing.T) {
	var (
		id     = uuid.NewString()
		client = f3Client.NewClient(f3Client.Timeout(1 * time.Microsecond))
	)

	account, err := client.Fetch(context.Background(), id)

	assert.Empty(t, account)
	assert.EqualError(t, err, f3Client.ErrTimeout.Error())
}

func TestIntegrDelete_WhenInvalidUUID_ThenReturnBadRequest(t *testing.T) {
	var (
		expErr = f3Client.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        errors.New("id is not a valid uuid;"),
		}

		client = f3Client.NewClient()
	)

	err := client.Delete(context.Background(), "invalid_uuid")

	assert.EqualValues(t, err, expErr)
}

func TestIntegrDelete_WhenResourceNotFound_ThenReturnBadRequest(t *testing.T) {
	var (
		expErr = f3Client.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        f3Client.ErrRecordNotFound,
		}

		client = f3Client.NewClient()
	)

	err := client.Delete(context.Background(), uuid.NewString())

	assert.EqualValues(t, err, expErr)
}

func TestIntegrDelete_WhenClientErr_ThenReturnGenericErr(t *testing.T) {
	expErr := errors.New("client_err")
	doerMockFunc := func(client http.Client, req *http.Request) (resp *http.Response, err error) {
		return &http.Response{}, expErr
	}
	client := f3Client.NewClient(f3Client.MockDoer(doerMockFunc))

	err := client.Delete(context.Background(), uuid.NewString())

	assert.EqualError(t, err, expErr.Error())
}

func TestIntegrDelete_WhenResourceRemoved_ThenSuccessWithNilErr(t *testing.T) {
	client := f3Client.NewClient()
	account := createTestAccount(t, &client)

	err := client.Delete(context.Background(), account.ID)
	deletedAccount := fetchTestAccountByID(t, &client, account.ID)

	assert.Empty(t, deletedAccount)
	assert.NoError(t, err)
}

func TestIntegrCreate_WhenInvalidRequest_ThenErrWithBadRequets(t *testing.T) {
	var (
		expErr = f3Client.RequestError{
			StatusCode: http.StatusBadRequest,
			Err: errors.New("attributes in body is required;id in body is " +
				"required;organisation_id in body is required;type in body is required;"),
		}
		client = f3Client.NewClient()
		req    = f3Client.AccountRequest{}
	)

	account, err := client.Create(context.Background(), req)

	assert.Empty(t, account)
	assert.EqualValues(t, err, expErr)
}

func TestIntegrCreateTimeout_WhenDeadlineReached_ThenReturnTimeoutErr(t *testing.T) {
	client := f3Client.NewClient(f3Client.Timeout(1 * time.Microsecond))

	account, err := client.Create(context.Background(), f3Client.AccountRequest{})

	assert.Empty(t, account)
	assert.EqualError(t, err, f3Client.ErrTimeout.Error())
}

func TestIntegrCreate_WhenDuplicatedUUID_ThenReturnErr(t *testing.T) {
	expErr := f3Client.RequestError{
		StatusCode: http.StatusConflict,
		Err:        errors.New("Account cannot be created as it violates a duplicate constraint"),
	}

	client := f3Client.NewClient()
	accountTest := createTestAccount(t, &client)

	defer cleanTestAccounts(t, &client, accountTest.ID)

	req := f3Client.AccountRequest{
		ID:             accountTest.ID,
		OrganisationID: accountTest.OrganisationID,
		Type:           accountTest.Type,
		Attributes: &f3Client.AccountAttributesRequest{
			Name:    accountTest.AccountAttributes.Name,
			Country: accountTest.AccountAttributes.Country,
		},
	}

	account, err := client.Create(context.Background(), req)

	assert.Empty(t, account)
	assert.EqualValues(t, err, expErr)
}

func TestIntegrCreate_WhenRequestSuccess_ThenCreatedAccount(t *testing.T) {
	expAccount := f3Client.Account{
		ID:             uuid.NewString(),
		OrganisationID: uuid.NewString(),
		Type:           "accounts",
		AccountAttributes: f3Client.AccountAttributes{
			Country: "AR",
			Name:    []string{"INT_TEST_DATA_account_name"},
		},
	}
	client := f3Client.NewClient()
	req := f3Client.AccountRequest{
		ID:             expAccount.ID,
		OrganisationID: expAccount.OrganisationID,
		Type:           expAccount.Type,
		Attributes: &f3Client.AccountAttributesRequest{
			Name:    expAccount.AccountAttributes.Name,
			Country: expAccount.AccountAttributes.Country,
		},
	}

	account, err := client.Create(context.Background(), req)
	defer cleanTestAccounts(t, &client, account.ID)

	expAccount.CreatedOn = account.CreatedOn
	expAccount.ModifiedOn = account.ModifiedOn

	assert.Equal(t, account, expAccount)
	assert.NoError(t, err)
}

func createTestAccount(t *testing.T, c *f3Client.Client) f3Client.Account {
	account, err := c.Create(context.Background(), f3Client.AccountRequest{
		ID:             uuid.NewString(),
		OrganisationID: uuid.NewString(),
		Type:           "accounts",
		Version:        nil,
		Attributes: &f3Client.AccountAttributesRequest{
			Name:    []string{"INT_TEST_DATA_account_name"},
			Country: "AR",
		},
	})

	if err != nil {
		t.Fatalf("Integration createTestAccount error [%s] while creating account test data", err.Error())
	}

	return account
}

const (
	deleteNotFoundErr = "status:404, error:'record does not exist'."
)

func cleanTestAccounts(t *testing.T, c *f3Client.Client, ids ...string) {
	var hasError bool

	for _, id := range ids {
		if err := c.Delete(context.Background(), id); err != nil && !strings.EqualFold(err.Error(), deleteNotFoundErr) {

			t.Logf("Intgration cleanTestAccounts error [%s] encountered while trying to delete id[%s]", err.Error(), ids)
			if !hasError {
				hasError = true
			}
		}
	}

	if hasError {
		t.Fatal("clean func end with errors, review for pending data to be clear")
	}
}

func fetchTestAccountByID(t *testing.T, c *f3Client.Client, accountID string) f3Client.Account {
	account, err := c.Fetch(context.Background(), accountID)
	switch err.(type) {
	case *f3Client.RequestError:
		return account
	default:
		t.Fatalf("Integration fetchTestAccountByID error [%s] while trying to fetch id[%s]", err.Error(), accountID)
	}

	return account
}
