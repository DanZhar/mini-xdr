package domain

import (
	"errors"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "uppercase host", input: "https://GitHub.com/a", want: "https://github.com/a"},
		{name: "path case preserved", input: "https://GitHub.com/Path/File", want: "https://github.com/Path/File"},
		{name: "uppercase scheme", input: "HTTPS://github.com/a", want: "https://github.com/a"},
		{name: "empty string", input: "", wantErr: true},
		{name: "no scheme", input: "github.com/a", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidURL) {
					t.Fatalf("NormalizeURL() error = %v, want ErrInvalidURL", err)
				}
				return
			}

			if got != tt.want {
				t.Fatalf("NormalizeURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
