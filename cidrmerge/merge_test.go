// go test -v -run="TestMergeCIDRs"

package cidrmerge

import (
	"reflect"
	"testing"
)

func TestMergeCIDRs(t *testing.T) {
	type TestCase struct {
		Input  []string
		Output []string
		Error  bool
	}

	testCases := []TestCase{
		{
			Input:  nil,
			Output: nil,
			Error:  false,
		},
		{
			Input:  []string{},
			Output: []string{},
			Error:  false,
		},
		{
			Input:  []string{"10.0.0.0/8"},
			Output: []string{"10.0.0.0/8"},
			Error:  false,
		},
		{
			Input:  []string{"10.0.0.0/8", "0.0.0.0/0"},
			Output: []string{"0.0.0.0/0"},
			Error:  false,
		},
	}

	for _, testCase := range testCases {
		output, err := MergeCIDRs(testCase.Input)
		if err != nil {
			if !testCase.Error {
				t.Errorf("MergeCIDRS(%#v) failed: %s", testCase.Input, err.Error())
			}
		}
		if !reflect.DeepEqual(testCase.Output, output) {
			t.Errorf("MergeCIDRS(%#v) expected: %#v, got: %#v", testCase.Input, testCase.Output, output)
		}
	}
}
