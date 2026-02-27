package cli

import (
	"testing"
	"time"

	"nitid/internal/core"
)

func TestUpdateTags(t *testing.T) {
	tags := []string{"go", "cli"}

	got := updateTags(tags, "add", "debug")
	want := []string{"cli", "debug", "go"}
	if len(got) != len(want) {
		t.Fatalf("add len mismatch: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("add mismatch: got=%v want=%v", got, want)
		}
	}

	got = updateTags(got, "rm", "cli")
	want = []string{"debug", "go"}
	if len(got) != len(want) {
		t.Fatalf("rm len mismatch: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rm mismatch: got=%v want=%v", got, want)
		}
	}
}

func TestNoteMatchesQuery(t *testing.T) {
	note := Note{
		Title:  "Investigate worker leak",
		Body:   "Found goroutine stuck in retry loop",
		Domain: "engineering",
		Tags:   []string{"go", "debug"},
	}

	if !noteMatchesQuery(note, "goroutine") {
		t.Fatalf("expected body match")
	}
	if !noteMatchesQuery(note, "engineering") {
		t.Fatalf("expected domain match")
	}
	if !noteMatchesQuery(note, "debug") {
		t.Fatalf("expected tag match")
	}
	if noteMatchesQuery(note, "nonexistent") {
		t.Fatalf("did not expect match")
	}
}

func TestSortNotesByTitleAsc(t *testing.T) {
	now := time.Now().UTC()
	notes := []NoteFile{
		{Note: Note{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Title: "zebra", UpdatedAt: now}},
		{Note: Note{ID: "01ARZ3NDEKTSV4RRFFQ69G5FB0", Title: "alpha", UpdatedAt: now}},
	}

	sortNotes(notes, "title", true)
	if notes[0].Note.Title != "alpha" {
		t.Fatalf("unexpected first title: %s", notes[0].Note.Title)
	}
}

func TestCollectNotePaths(t *testing.T) {
	root := t.TempDir()
	if err := createVaultStructure(root); err != nil {
		t.Fatalf("create vault: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	n := Note{
		ID:        newULID(now),
		Title:     "collect path",
		CreatedAt: now,
		UpdatedAt: now,
		Kind:      "note",
		Status:    statusInbox,
		Body:      "x",
	}
	if _, err := saveNote(root, "", n); err != nil {
		t.Fatalf("save note: %v", err)
	}

	report, err := core.New(root).Validate()
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if report.Total != 1 {
		t.Fatalf("expected 1 path, got %d", report.Total)
	}
}
