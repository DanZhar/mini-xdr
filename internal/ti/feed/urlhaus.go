package feed

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DanZhar/mini-xdr/internal/ti/domain"
)

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

		indicator := domain.Indicator{
			Type:      domain.TypeURL,
			Value:     rec[2],
			Threat:    rec[5],
			Tags:      tags,
			FirstSeen: firstSeen,
			LastSeen:  firstSeen,
		}

		result = append(result, indicator)
	}

	return result, nil
}
