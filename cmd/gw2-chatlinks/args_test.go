package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/Ev3nt1ne/gw2-chatlinks-go/api"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCode    string
		wantResolve bool
		wantJSON    bool
		wantVersion bool
		wantErr     bool
		wantHelp    bool
	}{
		{name: "bare code", args: []string{"[&abc]"}, wantCode: "[&abc]"},
		{name: "resolve before code", args: []string{"--resolve", "[&abc]"}, wantCode: "[&abc]", wantResolve: true},
		{name: "resolve after code", args: []string{"[&abc]", "--resolve"}, wantCode: "[&abc]", wantResolve: true},
		{name: "single dash resolve", args: []string{"-resolve", "[&abc]"}, wantCode: "[&abc]", wantResolve: true},
		{name: "json flag", args: []string{"--json", "[&abc]"}, wantCode: "[&abc]", wantJSON: true},
		{name: "version flag", args: []string{"--version"}, wantVersion: true},
		{name: "no args is valid (stdin path)", args: nil, wantCode: ""},
		{name: "help", args: []string{"-h"}, wantHelp: true, wantErr: true},
		{name: "long help", args: []string{"--help"}, wantHelp: true, wantErr: true},
		{name: "unknown flag", args: []string{"--nope"}, wantErr: true},
		{name: "too many positionals", args: []string{"[&a]", "[&b]"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			opts, code, showVersion, err := parseArgs(tt.args, &stderr)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (stderr=%q)", stderr.String())
				}
				if tt.wantHelp && !errors.Is(err, flag.ErrHelp) {
					t.Errorf("expected flag.ErrHelp, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if opts.resolve != tt.wantResolve {
				t.Errorf("resolve = %v, want %v", opts.resolve, tt.wantResolve)
			}
			if opts.asJSON != tt.wantJSON {
				t.Errorf("asJSON = %v, want %v", opts.asJSON, tt.wantJSON)
			}
			if showVersion != tt.wantVersion {
				t.Errorf("showVersion = %v, want %v", showVersion, tt.wantVersion)
			}
		})
	}
}

func TestParseArgs_UsageMentionsFlags(t *testing.T) {
	var stderr bytes.Buffer
	_, _, _, err := parseArgs([]string{"-h"}, &stderr)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected flag.ErrHelp, got %v", err)
	}
	for _, want := range []string{"resolve", "json", "version", "stdin"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("usage output missing %q:\n%s", want, stderr.String())
		}
	}
}

func TestResolveCode(t *testing.T) {
	t.Run("positional wins", func(t *testing.T) {
		got, err := resolveCode("[&abc]", strings.NewReader("ignored"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "[&abc]" {
			t.Errorf("got %q, want [&abc]", got)
		}
	})
	t.Run("empty reads stdin", func(t *testing.T) {
		got, err := resolveCode("", strings.NewReader("  [&fromstdin]\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "[&fromstdin]" {
			t.Errorf("got %q, want [&fromstdin]", got)
		}
	})
	t.Run("dash reads stdin", func(t *testing.T) {
		got, err := resolveCode("-", strings.NewReader("[&dashstdin]"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "[&dashstdin]" {
			t.Errorf("got %q, want [&dashstdin]", got)
		}
	})
	t.Run("empty stdin errors", func(t *testing.T) {
		if _, err := resolveCode("", strings.NewReader("   \n")); err == nil {
			t.Error("expected error for empty stdin, got nil")
		}
	})
}

func TestRun_JSONOutput_BuildTemplate(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, thiefSample, options{asJSON: true}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out struct {
		Type       string `json:"type"`
		Profession string `json:"profession"`
		Skills     []struct {
			Slot      string `json:"slot"`
			PaletteID int    `json:"palette_id"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if out.Type != "build_template" {
		t.Errorf("type = %q, want build_template", out.Type)
	}
	if out.Profession != "Thief" {
		t.Errorf("profession = %q, want Thief", out.Profession)
	}
	if len(out.Skills) == 0 || out.Skills[0].PaletteID != 3876 {
		t.Errorf("skills = %+v, want first palette 3876", out.Skills)
	}
}

func TestRun_JSONOutput_SimpleLink(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, "[&AgEAAAA=]", options{asJSON: true}, &api.Client{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out struct {
		Type string `json:"type"`
		ID   int    `json:"id"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if out.Type != "item" {
		t.Errorf("type = %q, want item", out.Type)
	}
}
