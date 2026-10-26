package domain

import "time"

type IndicatorType string

const (
	TypeURL    IndicatorType = "url"
	TypeIP     IndicatorType = "ip"
	TypeDomain IndicatorType = "domain"
	TypeMD5    IndicatorType = "md5"
	TypeSHA256 IndicatorType = "sha256"
)

type Indicator struct {
	Type      IndicatorType
	Value     string
	Threat    string
	Tags      []string
	FirstSeen time.Time
	LastSeen  time.Time
}
