package form3client

import (
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	empty, noRetry = 0, 0
)

type retryDoer struct {
	retryAttempts  int
	backoffIntvl   int
	maxJitterIntvl int
}

// NewRetryDoer has the default implementation of the client Do strategy, implementing
// an exponential backoff algorithm.
//
// RetryAttempts: type uint specifys the amount of retries attempts
//
// BackoffInterval: type uint specifys the behind backoff interval in miliseconds
// and is to use progressively exponential longer waits between retries for consecutive error responses
//
// MaximumJitterInterval: type uint specifys the maximum jitter interval (randomized delay) in miliseconds to prevent successive collisions
// use in the exponential backoff interval algorithm
//
// Retries is consider default if any of the params is set to its zero/empty value, so it will not retry
func NewRetryDoer(retryAttempts, backoffIntvl, maxJitterIntvl uint) Doer {
	if retryAttempts == noRetry || maxJitterIntvl == empty || backoffIntvl == empty {
		retryAttempts = noRetry
	}

	return retryDoer{
		int(retryAttempts),
		int(backoffIntvl),
		int(maxJitterIntvl),
	}
}

// Do will execute the request with an retry strategy.
//
// client (http.Client) the go standar net/http that sends the HTTP request.
//
// req (*http.Request) contains the request data
func (r retryDoer) Do(client http.Client, req *http.Request) (resp *http.Response, err error) {
	retries := empty

	if r.retryAttempts == noRetry {
		return client.Do(req)
	}

	for ; retries < r.retryAttempts; retries++ {
		resp, err = client.Do(req)
		if err == nil || os.IsTimeout(err) {
			break
		}

		backoffIntvl := int(float64(r.backoffIntvl)*math.Exp2(float64(retries))) + rand.Intn(r.maxJitterIntvl)
		time.Sleep(time.Duration(float64(backoffIntvl)))
	}

	if retries == r.retryAttempts {
		return resp, ErrRetryLimit
	}

	return resp, err
}
