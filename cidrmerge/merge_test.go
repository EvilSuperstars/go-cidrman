package cidrmerge

import (
	"testing"
)

func TestMergeCIDRs(t *testing.T) {
	merged, err := MergeCIDRs([]string{})
	if err != nil {
		t.Errorf("Failed: %s", err.Error())
	}
	if len(merged) != 0 {
		t.Errorf("Got %d", len(merged))
	}

	merged, err = MergeCIDRs([]string{"10.0.0.0/8"})
	if err != nil {
		t.Errorf("Failed: %s", err.Error())
	}
	if len(merged) != 1 {
		t.Errorf("Got %d", len(merged))
	}
}
