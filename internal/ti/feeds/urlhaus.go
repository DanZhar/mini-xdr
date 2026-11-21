package feeds

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DanZhar/mini-xdr/internal/ti/domain"
)

func FetchURLhaus(ctx context.Context, client *http.Client, authKey string) ([]domain.Indicator, error) {
	endpoint := "https://urlhaus-api.abuse.ch/v2/files/exports/" + url.PathEscape(authKey) + "/recent.csv"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build URLhaus request: %w", err)
	}

	// The request URL contains the key and must not become a Referer on redirects.
	downloadClient := *client
	downloadClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := downloadClient.Do(req)
	if err != nil {
		if requestErr, ok := err.(*url.Error); ok {
			err = requestErr.Err
		}
		return nil, fmt.Errorf("download URLhaus CSV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download URLhaus CSV: HTTP %d", resp.StatusCode)
	}

	indicators, err := (URLhausParser{}).Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse URLhaus CSV: %w", err)
	}
	if len(indicators) == 0 {
		return nil, errors.New("URLhaus CSV contained no valid indicators")
	}

	return indicators, nil
}

type URLhausParser struct{}

func (p URLhausParser) Parse(r io.Reader) ([]domain.Indicator, error) {
	cr := csv.NewReader(r)
	cr.Comment = '#'
	cr.FieldsPerRecord = 9
	var result []domain.Indicator

	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if errors.Is(err, csv.ErrFieldCount) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read csv: %w", err)
		}

		firstSeen, err := time.Parse("2006-01-02 15:04:05", rec[1])
		if err != nil {
			continue
		}

		var tags []string
		if rec[6] != "" {
			tags = strings.Split(rec[6], ",")
		}

		value, err := domain.NormalizeURL(rec[2])
		if err != nil {
			continue
		}

		indicator := domain.Indicator{
			Type:      domain.TypeURL,
			Value:     value,
			Threat:    rec[5],
			Tags:      tags,
			FirstSeen: firstSeen,
			LastSeen:  firstSeen,
		}

		result = append(result, indicator)
	}

	return result, nil
}
