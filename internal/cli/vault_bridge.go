package cli

import (
	"regexp"
	"strings"
	"time"

	"nitid/internal/vault"
)

var (
	domainIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	tagPattern      = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

const (
	statusInbox    = vault.StatusInbox
	statusActive   = vault.StatusActive
	statusArchived = vault.StatusArchived
)

var allowedStatuses = map[string]struct{}{
	statusInbox:    {},
	statusActive:   {},
	statusArchived: {},
}

type Note = vault.Note
type NoteFile = vault.NoteFile
type NoteFilter = vault.NoteFilter

func createVaultStructure(root string) error { return vault.CreateVaultStructure(root) }
func writeNote(root string, note Note) (string, error) {
	return vault.WriteNote(root, note)
}
func saveNote(root, currentPath string, note Note) (string, error) {
	return vault.SaveNote(root, currentPath, note)
}
func validateNoteForWrite(note Note) error { return vault.ValidateNoteForWrite(note) }
func newULID(now time.Time) string         { return vault.NewULID(now) }
func parseCSV(value string) []string       { return vault.ParseCSV(value) }
func isAllowedKind(kind string) bool {
	return vault.IsAllowedKind(strings.ToLower(strings.TrimSpace(kind)))
}
func listNotes(root string, filter NoteFilter) ([]NoteFile, error) {
	return vault.ListNotes(root, filter)
}
func uniqueIDPrefixes(notes []NoteFile, minLen int) map[string]string {
	return vault.UniqueIDPrefixes(notes, minLen)
}
func findNoteBySelector(root, selector string) (NoteFile, error) {
	return vault.FindNoteBySelector(root, selector)
}
func readNote(path string) (Note, error) { return vault.ReadNote(path) }
func resolveNotePath(root string, note Note) (string, error) {
	return vault.ResolveNotePath(root, note)
}
func sameFilePath(a, b string) (bool, error) { return vault.SameFilePath(a, b) }
