package hook

import (
	"strings"
	"testing"
)

func TestFormatIdentityBox(t *testing.T) {
	output := FormatIdentityBox(
		"~/code/project",
		"git@github.com:user/repo.git",
		"John Doe",
		"john@example.com",
		true,
		"work (github)",
	)

	if !strings.Contains(output, "John Doe") {
		t.Error("expected output to contain name")
	}
	if !strings.Contains(output, "john@example.com") {
		t.Error("expected output to contain email")
	}
	if !strings.Contains(output, "(local)") {
		t.Error("expected output to contain '(local)'")
	}
	if !strings.Contains(output, "work (github)") {
		t.Error("expected output to contain profile info")
	}
	if !strings.Contains(output, "Identity Check") {
		t.Error("expected output to contain 'Identity Check'")
	}
	if !strings.Contains(output, "~/code/project") {
		t.Error("expected output to contain repo path")
	}
	if !strings.Contains(output, "git@github.com:user/repo.git") {
		t.Error("expected output to contain remote URL")
	}
}

func TestFormatIdentityBoxGlobalWarning(t *testing.T) {
	output := FormatIdentityBox(
		"~/code/project",
		"git@github.com:user/repo.git",
		"John Doe",
		"john@example.com",
		false,
		"",
	)

	if !strings.Contains(output, "(global ⚠)") {
		t.Error("expected output to contain '(global ⚠)'")
	}
	if !strings.Contains(output, "No local config — using global fallback") {
		t.Error("expected output to contain global fallback warning message")
	}
	if !strings.Contains(output, "⚠ Identity Check") {
		t.Error("expected header to contain warning symbol")
	}
}

func TestFormatIdentityBoxLocalUnrecognized(t *testing.T) {
	output := FormatIdentityBox(
		"~/code/project",
		"git@github.com:user/repo.git",
		"John Doe",
		"john@example.com",
		true,
		"",
	)

	if !strings.Contains(output, "unrecognized") {
		t.Error("expected output to contain 'unrecognized' when local but no profile")
	}
	if !strings.Contains(output, "(local)") {
		t.Error("expected output to contain '(local)'")
	}
}

func TestPromptYes(t *testing.T) {
	reader := strings.NewReader("Y\n")
	if !Prompt(reader) {
		t.Error("expected Prompt to return true for 'Y'")
	}

	reader = strings.NewReader("y\n")
	if !Prompt(reader) {
		t.Error("expected Prompt to return true for 'y'")
	}

	reader = strings.NewReader("yes\n")
	if !Prompt(reader) {
		t.Error("expected Prompt to return true for 'yes'")
	}

	reader = strings.NewReader("YES\n")
	if !Prompt(reader) {
		t.Error("expected Prompt to return true for 'YES'")
	}
}

func TestPromptEnterDefault(t *testing.T) {
	reader := strings.NewReader("\n")
	if !Prompt(reader) {
		t.Error("expected Prompt to return true for empty input (default yes)")
	}
}

func TestPromptNo(t *testing.T) {
	reader := strings.NewReader("n\n")
	if Prompt(reader) {
		t.Error("expected Prompt to return false for 'n'")
	}

	reader = strings.NewReader("N\n")
	if Prompt(reader) {
		t.Error("expected Prompt to return false for 'N'")
	}

	reader = strings.NewReader("no\n")
	if Prompt(reader) {
		t.Error("expected Prompt to return false for 'no'")
	}

	reader = strings.NewReader("anything\n")
	if Prompt(reader) {
		t.Error("expected Prompt to return false for arbitrary input")
	}
}
