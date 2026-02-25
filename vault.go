package main

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"gopkg.in/yaml.v3"
)

var (
	domainIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	tagPattern      = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

const (
	statusInbox    = "inbox"
	statusActive   = "active"
	statusArchived = "archived"
)

var allowedStatuses = map[string]struct{}{
	statusInbox:    {},
	statusActive:   {},
	statusArchived: {},
}

type Note struct {
	ID        string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Domain    string
	Tags      []string
	Status    string
	Kind      string
	Links     []string
	Body      string
}

type NoteFile struct {
	Path    string
	RelPath string
	Note    Note
}

type NoteFilter struct {
	Domain string
	Tag    string
	Status string
	Kind   string
}

type noteFrontmatter struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	CreatedAt string   `yaml:"created_at"`
	UpdatedAt string   `yaml:"updated_at"`
	Domain    string   `yaml:"domain"`
	Tags      []string `yaml:"tags"`
	Status    string   `yaml:"status"`
	Kind      string   `yaml:"kind"`
	Links     []string `yaml:"links"`
}

func createVaultStructure(root string) error {
	dirs := []string{
		filepath.Join(root, "notes", "inbox"),
		filepath.Join(root, "notes", "domains"),
		filepath.Join(root, "notes", "daily"),
		filepath.Join(root, "notes", "archive"),
		filepath.Join(root, "assets"),
		filepath.Join(root, ".nitid", "cache"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	configPath := filepath.Join(root, ".nitid", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte(defaultConfig()), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func writeNote(root string, note Note) (string, error) {
	note.Status = normalizeStatus(note)
	path, err := resolveNotePath(root, note)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}

	content := renderMarkdown(note)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(rel), nil
}

func normalizeStatus(note Note) string {
	status := strings.ToLower(strings.TrimSpace(note.Status))
	if _, ok := allowedStatuses[status]; ok {
		return status
	}
	return deriveStatus(note)
}

func deriveStatus(note Note) string {
	if note.Kind == "daily" {
		return statusActive
	}
	if strings.TrimSpace(note.Domain) == "" {
		return statusInbox
	}
	return statusActive
}

func resolveNotePath(root string, note Note) (string, error) {
	fileName := fmt.Sprintf("%s--%s.md", note.ID, slugify(note.Title))

	if note.Kind == "daily" {
		y := note.CreatedAt.Format("2006")
		m := note.CreatedAt.Format("01")
		return filepath.Join(root, "notes", "daily", y, m, fileName), nil
	}

	if note.Status == statusArchived {
		return filepath.Join(root, "notes", "archive", fileName), nil
	}

	if note.Status == statusInbox {
		return filepath.Join(root, "notes", "inbox", fileName), nil
	}

	return filepath.Join(root, "notes", "domains", note.Domain, fileName), nil
}

func validateNoteForWrite(note Note) error {
	note.ID = strings.TrimSpace(note.ID)
	if note.ID == "" {
		return fmt.Errorf("note id cannot be empty")
	}
	if _, err := ulid.ParseStrict(note.ID); err != nil {
		return fmt.Errorf("note id must be a valid ULID")
	}
	if strings.TrimSpace(note.Title) == "" {
		return fmt.Errorf("note title cannot be empty")
	}
	if !isAllowedKind(note.Kind) {
		return fmt.Errorf("invalid kind %q", note.Kind)
	}
	status := normalizeStatus(note)
	if _, ok := allowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status %q", note.Status)
	}
	if note.Domain != "" && !domainIDPattern.MatchString(note.Domain) {
		return fmt.Errorf("invalid domain %q: use lowercase kebab-case", note.Domain)
	}

	for _, tag := range note.Tags {
		if !tagPattern.MatchString(tag) {
			return fmt.Errorf("invalid tag %q: use lowercase kebab-case", tag)
		}
	}

	return nil
}

func saveNote(root string, currentPath string, note Note) (string, error) {
	note.Status = normalizeStatus(note)
	if err := validateNoteForWrite(note); err != nil {
		return "", err
	}

	newPath, err := resolveNotePath(root, note)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return "", err
	}

	if err := os.WriteFile(newPath, []byte(renderMarkdown(note)), 0o644); err != nil {
		return "", err
	}

	if currentPath != "" {
		same, err := sameFilePath(currentPath, newPath)
		if err != nil {
			return "", err
		}
		if !same {
			if err := os.Remove(currentPath); err != nil {
				return "", err
			}
		}
	}

	rel, err := filepath.Rel(root, newPath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func sameFilePath(a, b string) (bool, error) {
	aAbs, err := filepath.Abs(a)
	if err != nil {
		return false, err
	}
	bAbs, err := filepath.Abs(b)
	if err != nil {
		return false, err
	}
	return aAbs == bAbs, nil
}

func listNotes(root string, filter NoteFilter) ([]NoteFile, error) {
	notesRoot := filepath.Join(root, "notes")
	result := make([]NoteFile, 0)

	err := filepath.WalkDir(notesRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		note, err := readNote(path)
		if err != nil {
			return err
		}

		if !matchesFilter(note, filter) {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result = append(result, NoteFile{
			Path:    path,
			RelPath: filepath.ToSlash(rel),
			Note:    note,
		})
		return nil
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []NoteFile{}, nil
		}
		return nil, err
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Note.UpdatedAt.Equal(result[j].Note.UpdatedAt) {
			return result[i].Note.ID > result[j].Note.ID
		}
		return result[i].Note.UpdatedAt.After(result[j].Note.UpdatedAt)
	})

	return result, nil
}

func matchesFilter(note Note, filter NoteFilter) bool {
	if filter.Domain != "" && note.Domain != filter.Domain {
		return false
	}
	if filter.Tag != "" {
		found := false
		for _, tag := range note.Tags {
			if tag == filter.Tag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Status != "" && note.Status != filter.Status {
		return false
	}
	if filter.Kind != "" && note.Kind != filter.Kind {
		return false
	}
	return true
}

func findNoteByID(root, id string) (NoteFile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return NoteFile{}, fmt.Errorf("note id is required")
	}

	notes, err := listNotes(root, NoteFilter{})
	if err != nil {
		return NoteFile{}, err
	}

	exact := make([]NoteFile, 0)
	prefix := make([]NoteFile, 0)
	for _, item := range notes {
		if item.Note.ID == id {
			exact = append(exact, item)
			continue
		}
		if strings.HasPrefix(item.Note.ID, id) {
			prefix = append(prefix, item)
		}
	}

	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) > 1 {
		return NoteFile{}, fmt.Errorf("multiple notes found for id %q", id)
	}
	if len(prefix) == 1 {
		return prefix[0], nil
	}
	if len(prefix) > 1 {
		return NoteFile{}, fmt.Errorf("multiple notes match prefix %q", id)
	}

	return NoteFile{}, fmt.Errorf("note %q not found", id)
}

func findNoteBySelector(root, selector string) (NoteFile, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return NoteFile{}, fmt.Errorf("note selector is required")
	}

	if strings.HasPrefix(selector, "#") || strings.HasPrefix(selector, "@") {
		idxRaw := strings.TrimPrefix(selector, "#")
		idxRaw = strings.TrimPrefix(idxRaw, "@")
		idx, err := strconv.Atoi(idxRaw)
		if err != nil || idx < 1 {
			return NoteFile{}, fmt.Errorf("invalid note ref %q", selector)
		}
		notes, err := listNotes(root, NoteFilter{})
		if err != nil {
			return NoteFile{}, err
		}
		if idx > len(notes) {
			return NoteFile{}, fmt.Errorf("note ref %q out of range", selector)
		}
		return notes[idx-1], nil
	}

	return findNoteByID(root, selector)
}

func uniqueIDPrefixes(notes []NoteFile, minLen int) map[string]string {
	result := make(map[string]string, len(notes))
	if minLen < 1 {
		minLen = 1
	}

	ids := make([]string, 0, len(notes))
	for _, item := range notes {
		ids = append(ids, item.Note.ID)
	}

	for _, id := range ids {
		chosen := id
		for size := minLen; size <= len(id); size++ {
			prefix := id[:size]
			unique := true
			for _, other := range ids {
				if other == id {
					continue
				}
				if strings.HasPrefix(other, prefix) {
					unique = false
					break
				}
			}
			if unique {
				chosen = prefix
				break
			}
		}
		result[id] = chosen
	}

	return result
}

func readNote(path string) (Note, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Note{}, err
	}

	fmRaw, body, err := splitFrontmatter(b)
	if err != nil {
		return Note{}, fmt.Errorf("read %s: %w", path, err)
	}

	var fm noteFrontmatter
	if err := yaml.Unmarshal(fmRaw, &fm); err != nil {
		return Note{}, fmt.Errorf("parse frontmatter in %s: %w", path, err)
	}

	createdAt, err := parseRFC3339Field("created_at", fm.CreatedAt)
	if err != nil {
		return Note{}, fmt.Errorf("%s: %w", path, err)
	}
	updatedAt, err := parseRFC3339Field("updated_at", fm.UpdatedAt)
	if err != nil {
		return Note{}, fmt.Errorf("%s: %w", path, err)
	}

	note := Note{
		ID:        strings.TrimSpace(fm.ID),
		Title:     strings.TrimSpace(fm.Title),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Domain:    strings.TrimSpace(fm.Domain),
		Tags:      sanitizeTags(fm.Tags),
		Status:    strings.TrimSpace(fm.Status),
		Kind:      strings.TrimSpace(fm.Kind),
		Links:     fm.Links,
		Body:      strings.TrimSpace(body),
	}

	if err := validateNoteForWrite(note); err != nil {
		return Note{}, fmt.Errorf("invalid note in %s: %w", path, err)
	}
	note.Status = normalizeStatus(note)

	return note, nil
}

func splitFrontmatter(content []byte) ([]byte, string, error) {
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(normalized, []byte("---\n")) {
		return nil, "", fmt.Errorf("missing YAML frontmatter header")
	}

	remaining := normalized[len("---\n"):]
	endMarker := []byte("\n---\n")
	idx := bytes.Index(remaining, endMarker)
	if idx < 0 {
		return nil, "", fmt.Errorf("missing YAML frontmatter footer")
	}

	fm := remaining[:idx]
	body := remaining[idx+len(endMarker):]
	return fm, string(body), nil
}

func parseRFC3339Field(field, value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("%s is required", field)
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be RFC3339", field)
	}
	return t.UTC(), nil
}

func sanitizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	clean := make([]string, 0, len(tags))
	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "" {
			continue
		}
		if _, exists := seen[t]; exists {
			continue
		}
		seen[t] = struct{}{}
		clean = append(clean, t)
	}
	sort.Strings(clean)
	return clean
}

func isAllowedKind(kind string) bool {
	return kind == "note" || kind == "adr" || kind == "snippet" || kind == "daily"
}

func parseCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}

	set := map[string]struct{}{}
	for _, piece := range strings.Split(value, ",") {
		t := strings.ToLower(strings.TrimSpace(piece))
		if t == "" {
			continue
		}
		set[t] = struct{}{}
	}

	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func newULID(now time.Time) string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(now), entropy).String()
}

func slugify(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return "untitled"
	}

	var b strings.Builder
	prevDash := false
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}

		if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "untitled"
	}
	return slug
}

func renderMarkdown(note Note) string {
	note.Status = normalizeStatus(note)
	note.Tags = sanitizeTags(note.Tags)
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: \"%s\"\n", note.ID))
	b.WriteString(fmt.Sprintf("title: \"%s\"\n", yamlEscape(note.Title)))
	b.WriteString(fmt.Sprintf("created_at: \"%s\"\n", note.CreatedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("updated_at: \"%s\"\n", note.UpdatedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("domain: \"%s\"\n", note.Domain))
	b.WriteString(fmt.Sprintf("tags: %s\n", renderInlineList(note.Tags)))
	b.WriteString(fmt.Sprintf("status: \"%s\"\n", note.Status))
	b.WriteString(fmt.Sprintf("kind: \"%s\"\n", note.Kind))
	b.WriteString(fmt.Sprintf("links: %s\n", renderInlineList(note.Links)))
	b.WriteString("---\n\n")
	b.WriteString(note.Body)
	b.WriteString("\n")
	return b.String()
}

func renderInlineList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}

	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf("\"%s\"", yamlEscape(item)))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func yamlEscape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
}

func defaultConfig() string {
	return strings.TrimSpace(`
[vault]
version = 1
default_domain = ""
default_kind = "note"
`) + "\n"
}
