package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	args := os.Args
	if len(args) <= 1 {
		fmt.Printf("Usage: %s <url>\n", args[0])
		return
	}

	url := args[1]

	maxRetries := 3
	retryCount := 1
	delay := 500 * time.Millisecond

	for retryCount <= maxRetries {
		resp, err := http.Get(url)

		if shouldRetry(resp, err) {
			if resp != nil {
				fmt.Printf("retry on response code: %d, round #%d\n", resp.StatusCode, retryCount)
			} else {
				fmt.Printf("retry on client error: %v, round #%d\n", err, retryCount)
			}

			retryCount++
			time.Sleep(delay)
			delay *= 2 // exponential backoff
			continue
		}

		fmt.Println("Got final result")
		if err == nil {
			fmt.Printf("Got response: %v\n", resp)
		} else {
			fmt.Printf("Got Err: %v\n", err)
		}

		break
	}
}

func shouldRetry(resp *http.Response, err error) bool {
	if err == nil {
		statusCode := resp.StatusCode
		return statusCode >= 500 || statusCode == http.StatusTooManyRequests
	}

	if err, ok := err.(*url.Error); ok && err.Timeout() {
		return true
	}

	if strings.Contains(err.Error(), "connection refused") {
		return true
	}

	return IsProbableEOF(err)
}

// IsProbableEOF returns true if the given error resembles a connection termination
// scenario that would justify assuming that the watch is empty.
// These errors are what the Go http stack returns back to us which are general
// connection closure errors (strongly correlated) and callers that need to
// differentiate probable errors in connection behavior between normal "this is
// disconnected" should use the method.
func IsProbableEOF(err error) bool {
	var uerr *url.Error
	if errors.As(err, &uerr) {
		err = uerr.Err
	}
	msg := err.Error()
	switch {
	case err == io.EOF:
		return true
	case err == io.ErrUnexpectedEOF:
		return true
	case msg == "http: can't write HTTP request on broken connection":
		return true
	case strings.Contains(msg, "http2: server sent GOAWAY and closed the connection"):
		return true
	case strings.Contains(msg, "connection reset by peer"):
		return true
	case strings.Contains(strings.ToLower(msg), "use of closed network connection"):
		return true
	}
	return false
}
