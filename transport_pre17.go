// +build !go1.7

package httpmock

import (
	"net/http"
	"errors"
	"fmt"
)

func runCancelable(responder Responder, req *http.Request) (*http.Response, error) {
	// Set up a goroutine that translates a close(req.Cancel) into a
	// "request canceled" error, and another one that runs the
	// responder. Then race them: first to the result channel wins.
	if req.Cancel == nil {
		return responder(req)
	}
	type result struct {
		response *http.Response
		err      error
	}
	resultch := make(chan result, 1)
	done := make(chan struct{}, 1)

	go func() {
		select {
		case <-req.Cancel:
			resultch <- result{
				response: nil,
				err:      errors.New("request canceled"),
			}
		case <-done:
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				resultch <- result{
					response: nil,
					err:      fmt.Errorf("panic in responder: got %q", err),
				}
			}
		}()

		response, err := responder(req)
		resultch <- result{
			response: response,
			err:      err,
		}
	}()

	r := <-resultch

	// if a close(req.Cancel) is never coming,
	// we'll need to unblock the first goroutine.
	done <- struct{}{}

	return r.response, r.err
}