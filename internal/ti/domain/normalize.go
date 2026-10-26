package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidURL = errors.New("invalid url")

func NormalizeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("%w: missing scheme or host in %q", ErrInvalidURL, raw)
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	return u.String(), nil
}
