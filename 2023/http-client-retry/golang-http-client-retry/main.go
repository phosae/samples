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

	var uerr *url.Error
	if !errors.As(err, &uerr) {
		return false
	}

	if uerr.Timeout() {
		return true
	}

	err = uerr.Err
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
	case strings.Contains(msg, "connection refused"):
		return true
	case strings.Contains(strings.ToLower(msg), "use of closed network connection"):
		return true
	}
	return false
}
