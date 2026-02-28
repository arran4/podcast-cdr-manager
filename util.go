package podcast_cdr_manager

import (
	"fmt"
	"net/http"
	"strconv"
)

func SkipFirstN[T any](args []T, i int) []T {
	m := i
	if m >= len(args) {
		m = len(args)
	}
	return args[m:]
}

func GetContentLength(url string) (int, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, fmt.Errorf("http head request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentLengthStr := resp.Header.Get("Content-Length")
	if contentLengthStr == "" {
		return 0, fmt.Errorf("no content-length header")
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return 0, fmt.Errorf("invalid content-length: %w", err)
	}

	return contentLength, nil
}
