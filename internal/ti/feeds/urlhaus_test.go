package feeds

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchURLhaus(t *testing.T) {
	const authKey = "test-key/with-slash"
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		const wantURL = "https://urlhaus-api.abuse.ch/v2/files/exports/test-key%2Fwith-slash/recent.csv"
		if req.Method != http.MethodGet || req.URL.String() != wantURL {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL)
		}

		const csv = "1,2026-01-02 15:04:05,https://Example.invalid/Path,online,,malware_download,,,tester\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(csv)),
		}, nil
	})}

	indicators, err := FetchURLhaus(t.Context(), client, authKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(indicators) != 1 || indicators[0].Value != "https://example.invalid/Path" {
		t.Fatalf("unexpected indicators: %+v", indicators)
	}
}

func TestFetchURLhausRejectsRedirect(t *testing.T) {
	requests := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if requests > 1 {
			t.Fatal("followed URLhaus redirect")
		}
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": {"https://redirect.invalid/"}},
			Body:       http.NoBody,
		}, nil
	})}

	_, err := FetchURLhaus(t.Context(), client, "test-key")
	if err == nil || !strings.Contains(err.Error(), "HTTP 302") {
		t.Fatalf("expected HTTP 302 error, got %v", err)
	}
}

func TestFetchURLhausTransportErrorOmitsKey(t *testing.T) {
	const authKey = "private-test-key"
	transportErr := errors.New("connection failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportErr
	})}

	_, err := FetchURLhaus(t.Context(), client, authKey)
	if !errors.Is(err, transportErr) || strings.Contains(err.Error(), authKey) {
		t.Fatalf("unexpected download error: %v", err)
	}
}

func TestURLhausParser_Fixture(t *testing.T) {
	f, err := os.Open("testdata/urlhaus_recent.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	got, err := URLhausParser{}.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("parsed %d indicators, first: %+v", len(got), got[0])
}
