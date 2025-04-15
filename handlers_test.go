package main

import (
	"fmt"
	"testing"
)

func TestFilterProfaneWords(t *testing.T) {

	cases := []struct {
		input  string
		output string
	}{
		{
			input:  "This is a kerfuffle opinion I need to share with the world",
			output: "This is a **** opinion I need to share with the world",
		},
	}

	for _, c := range cases {
		actual := filterProfaneWords(c.input)
		if actual != c.output {
			t.Errorf("test failed\n Expected: %s\n Actual: %s\n", c.output, actual)
		}
	}
	fmt.Println("All tests passed!!!")
	for _, c := range cases {
		t.Logf("input: %s\n", filterProfaneWords(c.input))
		t.Logf("output: %s\n", filterProfaneWords(c.output))
	}
}
