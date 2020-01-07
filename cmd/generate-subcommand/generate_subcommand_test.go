package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	update = flag.Bool("update", false, "update .golden files")
)

func Test(t *testing.T) {
	tests := []struct {
		desc       string
		goldenFile string
		params     tmplParams
	}{{
		"Bar",
		"testdata/bar.golden",
		tmplParams{
			Cmd:      "Bar",
			Pkg:      "bar",
			Synopsis: "sit amet, consectetur",
			Usage: `usage: bar [flags]

wopr bar -x`,
			User: "bob",
		},
	}, {
		"Baz",
		"testdata/baz.golden",
		tmplParams{
			Cmd:      "baz",
			Pkg:      "main",
			Synopsis: "lorem ipsum dolor",
			Usage: `usage: baz [flags]

wopr baz -q=42`,
			User: "alice",
		},
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var got bytes.Buffer

			if err := tmpl.Execute(&got, test.params); err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if *update {
				if err := ioutil.WriteFile(test.goldenFile, got.Bytes(), 0644); err != nil {
					t.Fatalf("Failed to update golden file (%s): %v", test.goldenFile, err)
				}
			}

			want, err := ioutil.ReadFile(test.goldenFile)
			if err != nil {
				t.Fatalf("Failed to open golden file (%s): %v", test.goldenFile, err)
			}

			if diff := cmp.Diff(got.String(), string(want)); diff != "" {
				t.Errorf("Golden files differ (-got +want)\n%s", diff)
			}
		})
	}
}
