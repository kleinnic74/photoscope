package formats

import (
	"testing"
	"time"
)

func TestLocationDecoding(t *testing.T) {
	raw := "+48.0880+016.2884+224.225/"
	expected := "+48.088000+016.288400/"
	var qt Quicktime
	if err := setLocation(&qt, "key", raw); err != nil {
		t.Fatal(err)
	}
	iso6709 := qt.coords.ISO6709()
	if iso6709 != expected {
		t.Errorf("Bad value for ISO6709, expected %s, got %s", expected, iso6709)
	}
}

func TestCreationDateDecoding(t *testing.T) {
	raw := "2016-12-04T08:25:39+0100"
	expected := "2016-12-04T08:25:39+01:00"
	var qt Quicktime
	if err := setCreationDate(&qt, "key", raw); err != nil {
		t.Fatal(err)
	}
	formatted := qt.creationDate.Format(time.RFC3339)
	if formatted != expected {
		t.Errorf("Bad value for time, expected %s, got %s", expected, formatted)
	}

}
