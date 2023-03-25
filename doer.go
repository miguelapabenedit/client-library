package form3client

import (
	"math"
	"math/rand"
	"net/http"
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

func (r retryDoer) Do(client http.Client, req *http.Request) (resp *http.Response, err error) {
	retries := empty

	for ; retries <= r.retryAttempts; retries++ {
		if retries > empty {
			backoffIntvl := int(float64(r.backoffIntvl)*math.Exp2(float64(retries))) + rand.Intn(r.maxJitterIntvl)
			time.Sleep(time.Duration(float64(backoffIntvl)))
		}

		resp, err = client.Do(req)
		if err == nil {
			break
		}
	}

	if retries >= r.retryAttempts {
		return resp, ErrRetryRequest
	}

	return resp, err
}
