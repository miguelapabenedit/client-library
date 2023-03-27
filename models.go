package form3client

import "time"

type CreateAccountRequest struct {
	Data AccountRequest `json:"data"`
}

type AccountRequest struct {
	Attributes     *AccountAttributesRequest `json:"attributes,omitempty"`
	ID             string                    `json:"id"`
	OrganisationID string                    `json:"organisation_id"`
	Type           string                    `json:"type"`
	Version        *int64                    `json:"version,omitempty"`
}

type AccountAttributesRequest struct {
	AccountClassification   *string  `json:"account_classification,omitempty"`
	AccountMatchingOptOut   *bool    `json:"account_matching_opt_out,omitempty"`
	AccountNumber           string   `json:"account_number,omitempty"`
	AlternativeNames        []string `json:"alternative_names,omitempty"`
	BankID                  string   `json:"bank_id,omitempty"`
	BankIDCode              string   `json:"bank_id_code,omitempty"`
	BaseCurrency            string   `json:"base_currency,omitempty"`
	Bic                     string   `json:"bic,omitempty"`
	Country                 string   `json:"country"`
	Iban                    string   `json:"iban,omitempty"`
	JointAccount            *bool    `json:"joint_account,omitempty"`
	Name                    []string `json:"name"`
	SecondaryIdentification string   `json:"secondary_identification,omitempty"`
	Status                  *string  `json:"status,omitempty"`
	Switched                *bool    `json:"switched,omitempty"`
}

type ResponseError struct {
	ErrorMessage string `json:"error_message"`
}

type AccountResponse struct {
	Account Account `json:"data"`
	Links   Links   `json:"links"`
}

type Account struct {
	ID                string            `json:"id"`
	OrganisationID    string            `json:"organisation_id"`
	Type              string            `json:"type"`
	Version           int               `json:"version"`
	AccountAttributes AccountAttributes `json:"attributes"`
	CreatedOn         time.Time         `json:"created_on"`
	ModifiedOn        time.Time         `json:"modified_on"`
}

type AccountAttributes struct {
	Country          string   `json:"country"`
	Name             []string `json:"name"`
	AlternativeNames []string `json:"alternative_names"`
}

type Links struct {
	Self string `json:"self"`
}
