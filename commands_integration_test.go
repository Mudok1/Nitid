package main

import (
	"io"
	"nitid/internal/cli"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const exitOKCode = 0

type cliResult struct {
	code   int
	stdout string
	stderr string
}

func runCLI(t *testing.T, dir string, args []string, stdin string) cliResult {
	t.Helper()

	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldStdin := os.Stdin
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW
	os.Stdin = stdinR
	os.Args = append([]string{"ntd"}, args...)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if stdin != "" {
		if _, err := stdinW.Write([]byte(stdin)); err != nil {
			t.Fatalf("write stdin: %v", err)
		}
	}
	_ = stdinW.Close()

	code := cli.Run(args)

	_ = stdoutW.Close()
	_ = stderrW.Close()

	stdoutBytes, _ := io.ReadAll(stdoutR)
	stderrBytes, _ := io.ReadAll(stderrR)

	_ = stdoutR.Close()
	_ = stderrR.Close()
	_ = stdinR.Close()

	os.Args = oldArgs
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	os.Stdin = oldStdin
	_ = os.Chdir(oldCwd)

	return cliResult{code: code, stdout: string(stdoutBytes), stderr: string(stderrBytes)}
}

func mustOK(t *testing.T, r cliResult) {
	t.Helper()
	if r.code != exitOKCode {
		t.Fatalf("expected exitOK, got %d, stderr=%s", r.code, r.stderr)
	}
}

func mustFail(t *testing.T, r cliResult) {
	t.Helper()
	if r.code == exitOKCode {
		t.Fatalf("expected failure, stdout=%s", r.stdout)
	}
}

func TestCLI_HelpAndVersion(t *testing.T) {
	dir := t.TempDir()

	r := runCLI(t, dir, []string{"help"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "Nitid (ntd)") {
		t.Fatalf("help output missing header: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"version"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "ntd ") {
		t.Fatalf("version output missing: %s", r.stdout)
	}
}

func TestCLI_InitCaptureListShow(t *testing.T) {
	dir := t.TempDir()

	mustOK(t, runCLI(t, dir, []string{"init", "."}, ""))
	mustOK(t, runCLI(t, dir, []string{"capture", "--title", "Hello note", "--domain", "engineering", "hello body"}, ""))

	r := runCLI(t, dir, []string{"ls"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "Hello note") || !strings.Contains(r.stdout, "@1") {
		t.Fatalf("ls output unexpected: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"show", "@1"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "Title:   Hello note") || !strings.Contains(r.stdout, "hello body") {
		t.Fatalf("show output unexpected: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"show", "@1", "--raw"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "---") || !strings.Contains(r.stdout, "title: \"Hello note\"") {
		t.Fatalf("raw show output unexpected: %s", r.stdout)
	}
}

func TestCLI_MoveTagArchiveFlow(t *testing.T) {
	dir := t.TempDir()
	mustOK(t, runCLI(t, dir, []string{"init", "."}, ""))
	mustOK(t, runCLI(t, dir, []string{"capture", "--title", "Inbox test", "work"}, ""))

	mustOK(t, runCLI(t, dir, []string{"move", "@1", "--domain", "engineering"}, ""))
	mustOK(t, runCLI(t, dir, []string{"tag", "@1", "add", "go"}, ""))
	mustOK(t, runCLI(t, dir, []string{"archive", "@1"}, ""))

	r := runCLI(t, dir, []string{"ls", "--status", "archived"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "Inbox test") {
		t.Fatalf("expected archived note in ls: %s", r.stdout)
	}
}

func TestCLI_FindAndSort(t *testing.T) {
	dir := t.TempDir()
	mustOK(t, runCLI(t, dir, []string{"init", "."}, ""))
	mustOK(t, runCLI(t, dir, []string{"capture", "--title", "beta note", "hello world"}, ""))
	mustOK(t, runCLI(t, dir, []string{"capture", "--title", "alpha note", "gopher"}, ""))

	r := runCLI(t, dir, []string{"ls", "--sort", "title", "--asc"}, "")
	mustOK(t, r)
	alpha := strings.Index(r.stdout, "alpha note")
	beta := strings.Index(r.stdout, "beta note")
	if alpha < 0 || beta < 0 || alpha > beta {
		t.Fatalf("expected alpha before beta in sorted list: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"find", "hello", "--limit", "1"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "beta note") {
		t.Fatalf("find output unexpected: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"find", "does-not-exist"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "no matching notes found") {
		t.Fatalf("expected no matches message: %s", r.stdout)
	}
}

func TestCLI_TemplatesNewDaily(t *testing.T) {
	dir := t.TempDir()
	mustOK(t, runCLI(t, dir, []string{"init", "."}, ""))

	r := runCLI(t, dir, []string{"templates"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "adr") || !strings.Contains(r.stdout, "bug") {
		t.Fatalf("templates output unexpected: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"templates", "show", "adr"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "## Context") {
		t.Fatalf("template show output unexpected: %s", r.stdout)
	}

	mustOK(t, runCLI(t, dir, []string{"new", "adr", "--title", "Decision A"}, ""))

	r = runCLI(t, dir, []string{"find", "decision"}, "")
	mustOK(t, r)
	if !strings.Contains(strings.ToLower(r.stdout), "decision a") {
		t.Fatalf("new template note not found: %s", r.stdout)
	}

	mustOK(t, runCLI(t, dir, []string{"daily", "--date", "2026-02-25"}, ""))
	r = runCLI(t, dir, []string{"daily", "--date", "2026-02-25"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "already exists") {
		t.Fatalf("expected existing daily note message: %s", r.stdout)
	}
}

func TestCLI_CleanValidateDoctorCompletion(t *testing.T) {
	dir := t.TempDir()
	mustOK(t, runCLI(t, dir, []string{"init", "."}, ""))
	mustOK(t, runCLI(t, dir, []string{"capture", "--title", "Health check", "body"}, ""))

	swapPath := filepath.Join(dir, "notes", "inbox", ".temp-note.swp")
	if err := os.WriteFile(swapPath, []byte("swap"), 0o644); err != nil {
		t.Fatalf("write swap file: %v", err)
	}

	r := runCLI(t, dir, []string{"clean", "--dry-run"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "would remove") {
		t.Fatalf("expected clean dry-run output: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"clean"}, "")
	mustOK(t, r)
	if _, err := os.Stat(swapPath); !os.IsNotExist(err) {
		t.Fatalf("swap file should be removed")
	}

	r = runCLI(t, dir, []string{"validate"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "validation passed") {
		t.Fatalf("validate output unexpected: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"doctor"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "doctor status: ok") {
		t.Fatalf("doctor output unexpected: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"completion", "bash"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "complete -F _ntd_complete ntd") {
		t.Fatalf("completion output unexpected: %s", r.stdout)
	}

	r = runCLI(t, dir, []string{"__complete_ids"}, "")
	mustOK(t, r)
	if !strings.Contains(r.stdout, "@1") {
		t.Fatalf("complete ids output unexpected: %s", r.stdout)
	}
}

func TestCLI_InvalidUsageCases(t *testing.T) {
	dir := t.TempDir()
	mustOK(t, runCLI(t, dir, []string{"init", "."}, ""))

	cases := []struct {
		args    []string
		message string
	}{
		{args: []string{"show"}, message: "show requires exactly one"},
		{args: []string{"move", "@1"}, message: "invalid domain"},
		{args: []string{"tag", "@1", "nope", "go"}, message: "invalid tag action"},
		{args: []string{"find"}, message: "find requires a query string"},
		{args: []string{"daily", "--date", "25-02-2026"}, message: "YYYY-MM-DD"},
		{args: []string{"templates", "show", "unknown"}, message: "unknown template"},
		{args: []string{"ls", "--sort", "invalid"}, message: "invalid sort"},
		{args: []string{"completion", "zsh"}, message: "completion usage"},
	}

	for _, tc := range cases {
		r := runCLI(t, dir, tc.args, "")
		mustFail(t, r)
		if !strings.Contains(strings.ToLower(r.stderr), strings.ToLower(tc.message)) {
			t.Fatalf("expected error message %q, stderr=%s", tc.message, r.stderr)
		}
	}
}
