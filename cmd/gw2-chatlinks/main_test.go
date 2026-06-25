package main

import (
	"bytes"
	"strings"
	"testing"
)

const thiefSample = "[&DQUAAAAAAAAkDyQPAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACLwBVAAA=]"

func TestRun_BuildTemplate_NoResolve(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, thiefSample, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "type: build_template") {
		t.Errorf("output missing type line: %s", out)
	}
	if !strings.Contains(out, "profession: Thief") {
		t.Errorf("output missing profession line: %s", out)
	}
	if !strings.Contains(out, "palette=3876") {
		t.Errorf("output missing heal palette: %s", out)
	}
}

func TestRun_ItemLink_NoResolve(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, "[&AgEAAAA=]", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "type: item") {
		t.Errorf("output missing type line: %s", out)
	}
	if !strings.Contains(out, "id: 0") {
		t.Errorf("output missing id line: %s", out)
	}
}

func TestRun_UnimplementedType(t *testing.T) {
	var buf bytes.Buffer
	// header 0x01 = coin, no decoder implemented.
	if err := run(&buf, "[&AQAAAAA=]", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "decoder for this type not implemented yet") {
		t.Errorf("expected fallback message, got: %s", buf.String())
	}
}

func TestRun_InvalidCode(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, "[&not-valid!!]", false); err == nil {
		t.Error("expected error for invalid chat link, got nil")
	}
}

func TestRun_ResolveUnsupportedForRecipe(t *testing.T) {
	// Recipe header (0x09) with a 3-byte id; --resolve has no recipe-name
	// endpoint wired up, so this should surface an error rather than hang
	// or silently no-op.
	var buf bytes.Buffer
	err := run(&buf, "[&CQEAAAA=]", true)
	if err == nil {
		t.Error("expected error for --resolve on a recipe link, got nil")
	}
}
