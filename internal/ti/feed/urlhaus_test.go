package feed

import (
	"os"
	"testing"
)

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
