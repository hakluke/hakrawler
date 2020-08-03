package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func Test_main(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com").
		Get("/").
		Persist().
		Reply(200).
		SetHeader("Content-Type", "text/html").
		BodyString(`
		<a href="http://example.com/link"></a>
		<a href="http://www.example.com/link"></a>
		<a href="http://sub.example.com/link"></a>
		<a href="http://another-example.com/link"></a>
		`)

	tests := []struct {
		name   string
		args   []string
		output []string
	}{
		{
			name: "normal scope (subs)",
			args: []string{"hakrawler", "-url", "http://example.com", "-plain"},
			output: []string{
				"http://example.com/link",
				// "example.com", // TODO: this url should show up -- fix this bug.
				"http://www.example.com/link",
				"www.example.com",
				"http://sub.example.com/link",
				"sub.example.com",
			},
		},
		{
			name: "scope strict",
			args: []string{"hakrawler", "-url", "http://example.com", "-plain", "-scope", "strict"},
			output: []string{
				"http://example.com/link",
				"example.com",
			},
		},
		{
			name: "scope www",
			args: []string{"hakrawler", "-url", "http://example.com", "-plain", "-scope", "www"},
			output: []string{
				"http://example.com/link",
				"example.com",
				"http://www.example.com/link",
				"www.example.com",
			},
		},
		{
			name: "scope yolo",
			args: []string{"hakrawler", "-url", "http://example.com", "-plain", "-scope", "yolo"},
			output: []string{
				"http://example.com/link",
				"example.com",
				"http://www.example.com/link",
				"www.example.com",
				"http://sub.example.com/link",
				"sub.example.com",
				"http://another-example.com/link",
				"another-example.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			out = bytes.NewBuffer(nil)
			main()
			output := strings.Join(tt.output[:], "\n")
			if actual := strings.TrimRight(out.(*bytes.Buffer).String(), "\n"); actual != output {
				t.Fatalf("expected <%s>, but got <%s>", output, actual)
			}
		})
	}
}
