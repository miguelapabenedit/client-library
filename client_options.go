package form3client

import (
	"time"
)

const (
	//Zero specify an empty time.Duration
	empty = 0
)

// ClientOption is any function that can work as an option to set Client
// features as defined in Functional Options, please see:
//
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type ClientOption func(Client) Client

// Timeout specifies a time limit for requests made by the Client.
//
// The timeout includes connection time, any redirects, and reading
// the response body. The timer remains running after Get, Head, Post,
// or Do return and will interrupt reading of the Response.Body.
//
// A Timeout of zero means no timeout.
//
// The Client cancels requests to the underlying Transport
// as if the Request's Context ended.
func Timeout(timeout time.Duration) ClientOption {
	return func(c Client) Client {
		c.client.Timeout = timeout
		return c
	}
}

// Retries specifies the number of request attempts made by the Client
// in the presence of an error. This is optional and its default is to
// not retry.
//
// RetryAttempts: type int specifys the amount of retries attempts
//
// BackoffInterval: type int specifys the behind backoff interval in miliseconds
// and is to use progressively exponential longer waits between retries for consecutive error responses
//
// MaximumJitterInterval: type int specifys the maximum jitter interval (randomized delay) in miliseconds to prevent successive collisions
// use in the exponential backoff interval algorithm
//
// Retries is consider default if any of the params is set to its zero/empty value, so it will not retry
func Retries(retryAttempts, backoffIntvl, maxJitterIntvl int) ClientOption {
	return func(c Client) Client {
		if retryAttempts < 1 || maxJitterIntvl <= empty || backoffIntvl <= empty {
			retryAttempts = 0
		}

		c.retryAttempts = retryAttempts
		c.backoffIntvl = backoffIntvl
		c.maxJitterIntvl = maxJitterIntvl

		return c
	}
}
