package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty", input: "", want: []string{}},
		{name: "normal", input: "go,cli,debug", want: []string{"cli", "debug", "go"}},
		{name: "spaces and duplicates", input: "go, cli,GO, ,debug", want: []string{"cli", "debug", "go"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCSV(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("len mismatch: got=%v want=%v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("value mismatch: got=%v want=%v", got, tt.want)
				}
			}
		})
	}
}

func TestResolveNotePathByStatus(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 2, 25, 10, 0, 0, 0, time.UTC)
	note := Note{
		ID:        "01JN8PX5WP8J67JAY2P2CVJH6D",
		Title:     "Test note",
		CreatedAt: now,
		UpdatedAt: now,
		Kind:      "note",
	}

	note.Status = statusInbox
	inboxPath, err := resolveNotePath(root, note)
	if err != nil {
		t.Fatalf("resolve inbox path: %v", err)
	}
	if filepath.Dir(inboxPath) != filepath.Join(root, "notes", "inbox") {
		t.Fatalf("unexpected inbox path: %s", inboxPath)
	}

	note.Status = statusActive
	note.Domain = "engineering"
	activePath, err := resolveNotePath(root, note)
	if err != nil {
		t.Fatalf("resolve active path: %v", err)
	}
	if filepath.Dir(activePath) != filepath.Join(root, "notes", "domains", "engineering") {
		t.Fatalf("unexpected active path: %s", activePath)
	}

	note.Status = statusArchived
	archivePath, err := resolveNotePath(root, note)
	if err != nil {
		t.Fatalf("resolve archive path: %v", err)
	}
	if filepath.Dir(archivePath) != filepath.Join(root, "notes", "archive") {
		t.Fatalf("unexpected archive path: %s", archivePath)
	}
}

func TestSaveAndReadNoteRoundtrip(t *testing.T) {
	root := t.TempDir()
	if err := createVaultStructure(root); err != nil {
		t.Fatalf("create vault: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	note := Note{
		ID:        newULID(now),
		Title:     "Roundtrip test",
		CreatedAt: now,
		UpdatedAt: now,
		Domain:    "engineering",
		Tags:      []string{"go", "cli"},
		Status:    statusActive,
		Kind:      "note",
		Links:     []string{},
		Body:      "hello world",
	}

	rel, err := saveNote(root, "", note)
	if err != nil {
		t.Fatalf("save note: %v", err)
	}

	loaded, err := readNote(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read note: %v", err)
	}

	if loaded.ID != note.ID || loaded.Title != note.Title || loaded.Domain != note.Domain {
		t.Fatalf("loaded note mismatch: got=%+v want=%+v", loaded, note)
	}
	if loaded.Body != note.Body {
		t.Fatalf("body mismatch: got=%q want=%q", loaded.Body, note.Body)
	}
}

func TestFindNoteBySelectorWithAtRef(t *testing.T) {
	root := t.TempDir()
	if err := createVaultStructure(root); err != nil {
		t.Fatalf("create vault: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	noteA := Note{
		ID:        newULID(now),
		Title:     "First note",
		CreatedAt: now,
		UpdatedAt: now,
		Domain:    "engineering",
		Status:    statusActive,
		Kind:      "note",
		Body:      "a",
	}
	noteB := Note{
		ID:        newULID(now.Add(time.Second)),
		Title:     "Second note",
		CreatedAt: now.Add(time.Second),
		UpdatedAt: now.Add(time.Second),
		Domain:    "engineering",
		Status:    statusActive,
		Kind:      "note",
		Body:      "b",
	}

	if _, err := saveNote(root, "", noteA); err != nil {
		t.Fatalf("save noteA: %v", err)
	}
	if _, err := saveNote(root, "", noteB); err != nil {
		t.Fatalf("save noteB: %v", err)
	}

	selected, err := findNoteBySelector(root, "@1")
	if err != nil {
		t.Fatalf("select @1: %v", err)
	}
	if selected.Note.ID != noteB.ID {
		t.Fatalf("@1 should be most recently updated note: got=%s want=%s", selected.Note.ID, noteB.ID)
	}
}
