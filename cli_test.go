package main

import "testing"

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
