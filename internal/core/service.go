package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"nitid/internal/vault"
)

type Note = vault.Note
type NoteFile = vault.NoteFile
type NoteFilter = vault.NoteFilter

type Service struct {
	root string
}

type MutationResult struct {
	NoteID  string
	RelPath string
}

func New(root string) *Service {
	return &Service{root: root}
}

func (s *Service) Root() string {
	return s.root
}

func (s *Service) Init() error {
	return vault.CreateVaultStructure(s.root)
}

func (s *Service) Create(note Note) (string, error) {
	if err := vault.ValidateNoteForWrite(note); err != nil {
		return "", err
	}
	if err := s.Init(); err != nil {
		return "", err
	}
	return vault.WriteNote(s.root, note)
}

func (s *Service) List(filter NoteFilter, sortBy string, asc bool) ([]NoteFile, error) {
	notes, err := vault.ListNotes(s.root, filter)
	if err != nil {
		return nil, err
	}
	SortNotes(notes, sortBy, asc)
	return notes, nil
}

func (s *Service) Find(query string, filter NoteFilter, limit int) ([]NoteFile, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil, errors.New("find query cannot be empty")
	}
	if limit < 1 {
		return nil, errors.New("limit must be at least 1")
	}

	notes, err := vault.ListNotes(s.root, filter)
	if err != nil {
		return nil, err
	}

	matches := make([]NoteFile, 0)
	for _, item := range notes {
		if NoteMatchesQuery(item.Note, query) {
			matches = append(matches, item)
		}
	}
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

func (s *Service) FindBySelector(selector string) (NoteFile, error) {
	return vault.FindNoteBySelector(s.root, selector)
}

func (s *Service) FindDailyByDate(date time.Time) (NoteFile, bool, error) {
	notes, err := vault.ListNotes(s.root, NoteFilter{Kind: "daily"})
	if err != nil {
		return NoteFile{}, false, err
	}

	target := date.Format("2006-01-02")
	for _, item := range notes {
		if item.Note.CreatedAt.Format("2006-01-02") == target {
			return item, true, nil
		}
	}

	return NoteFile{}, false, nil
}

func (s *Service) Move(selector, domain string) (MutationResult, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if !vault.IsValidDomainID(domain) {
		return MutationResult{}, fmt.Errorf("invalid domain %q: use lowercase kebab-case", domain)
	}

	noteFile, err := s.FindBySelector(selector)
	if err != nil {
		return MutationResult{}, err
	}

	note := noteFile.Note
	note.Domain = domain
	note.Status = vault.StatusActive
	note.UpdatedAt = time.Now().UTC()

	rel, err := vault.SaveNote(s.root, noteFile.Path, note)
	if err != nil {
		return MutationResult{}, err
	}

	return MutationResult{NoteID: note.ID, RelPath: rel}, nil
}

func (s *Service) Tag(selector, action, tag string) (MutationResult, error) {
	action = strings.ToLower(strings.TrimSpace(action))
	tag = strings.ToLower(strings.TrimSpace(tag))

	if !vault.IsValidTag(tag) {
		return MutationResult{}, fmt.Errorf("invalid tag %q: use lowercase kebab-case", tag)
	}
	if action != "add" && action != "rm" {
		return MutationResult{}, fmt.Errorf("invalid tag action %q", action)
	}

	noteFile, err := s.FindBySelector(selector)
	if err != nil {
		return MutationResult{}, err
	}

	note := noteFile.Note
	note.Tags = UpdateTags(note.Tags, action, tag)
	note.UpdatedAt = time.Now().UTC()

	rel, err := vault.SaveNote(s.root, noteFile.Path, note)
	if err != nil {
		return MutationResult{}, err
	}

	return MutationResult{NoteID: note.ID, RelPath: rel}, nil
}

func (s *Service) Archive(selector string) (MutationResult, error) {
	noteFile, err := s.FindBySelector(selector)
	if err != nil {
		return MutationResult{}, err
	}

	note := noteFile.Note
	note.Status = vault.StatusArchived
	note.UpdatedAt = time.Now().UTC()

	rel, err := vault.SaveNote(s.root, noteFile.Path, note)
	if err != nil {
		return MutationResult{}, err
	}

	return MutationResult{NoteID: note.ID, RelPath: rel}, nil
}

func (s *Service) Edit(selector string) error {
	noteFile, err := s.FindBySelector(selector)
	if err != nil {
		return err
	}
	return OpenInEditor(noteFile.Path)
}

func NoteMatchesQuery(note Note, query string) bool {
	if strings.Contains(strings.ToLower(note.Title), query) {
		return true
	}
	if strings.Contains(strings.ToLower(note.Body), query) {
		return true
	}
	if strings.Contains(strings.ToLower(note.Domain), query) {
		return true
	}
	for _, tag := range note.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func SortNotes(notes []NoteFile, mode string, asc bool) {
	less := func(i, j int) bool {
		a := notes[i].Note
		b := notes[j].Note

		switch mode {
		case "created":
			if a.CreatedAt.Equal(b.CreatedAt) {
				if asc {
					return a.ID < b.ID
				}
				return a.ID > b.ID
			}
			if asc {
				return a.CreatedAt.Before(b.CreatedAt)
			}
			return a.CreatedAt.After(b.CreatedAt)
		case "title":
			aTitle := strings.ToLower(a.Title)
			bTitle := strings.ToLower(b.Title)
			if aTitle == bTitle {
				if asc {
					return a.ID < b.ID
				}
				return a.ID > b.ID
			}
			if asc {
				return aTitle < bTitle
			}
			return aTitle > bTitle
		case "id":
			if asc {
				return a.ID < b.ID
			}
			return a.ID > b.ID
		default:
			if a.UpdatedAt.Equal(b.UpdatedAt) {
				if asc {
					return a.ID < b.ID
				}
				return a.ID > b.ID
			}
			if asc {
				return a.UpdatedAt.Before(b.UpdatedAt)
			}
			return a.UpdatedAt.After(b.UpdatedAt)
		}
	}

	sort.SliceStable(notes, less)
}

func UpdateTags(tags []string, action, value string) []string {
	set := map[string]struct{}{}
	for _, tag := range tags {
		set[tag] = struct{}{}
	}

	if action == "add" {
		set[value] = struct{}{}
	} else {
		delete(set, value)
	}

	out := make([]string, 0, len(set))
	for tag := range set {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func ResolveEditor() string {
	editor := strings.TrimSpace(os.Getenv("VISUAL"))
	if editor != "" {
		return editor
	}

	editor = strings.TrimSpace(os.Getenv("EDITOR"))
	if editor != "" {
		return editor
	}

	if _, err := exec.LookPath("nano"); err == nil {
		return "nano"
	}
	return "vi"
}

func EditorCommand(path string) (*exec.Cmd, error) {
	editor := ResolveEditor()
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return nil, errors.New("invalid editor command")
	}

	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func OpenInEditor(path string) error {
	cmd, err := EditorCommand(path)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open editor %q: %w", ResolveEditor(), err)
	}
	return nil
}
