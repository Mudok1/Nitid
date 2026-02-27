package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type templateDef struct {
	Kind         string
	DefaultTitle string
	BodyTemplate string
}

type validationReport struct {
	Total    int
	Warnings []string
	Errors   []string
}

func templateDefinitions() map[string]templateDef {
	return map[string]templateDef{
		"note": {
			Kind:         "note",
			DefaultTitle: "Untitled",
			BodyTemplate: "",
		},
		"adr": {
			Kind:         "adr",
			DefaultTitle: "Architecture decision",
			BodyTemplate: "## Context\n\n## Decision\n\n## Consequences\n",
		},
		"meeting": {
			Kind:         "note",
			DefaultTitle: "Meeting notes",
			BodyTemplate: "## Date\n\n## Attendees\n\n## Notes\n\n## Action items\n- [ ] ",
		},
		"bug": {
			Kind:         "note",
			DefaultTitle: "Bug report",
			BodyTemplate: "## Symptoms\n\n## Steps to reproduce\n\n## Root cause\n\n## Fix\n\n## Validation\n",
		},
	}
}

func templateBody(def templateDef, extraText string) string {
	base := strings.TrimSpace(def.BodyTemplate)
	extra := strings.TrimSpace(extraText)

	if base == "" {
		return extra
	}
	if extra == "" {
		return base
	}
	return base + "\n\n" + extra
}

func resolveDailyDate(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC), nil
	}

	t, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, errors.New("daily --date must use YYYY-MM-DD")
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, time.UTC), nil
}

func findDailyByDate(root string, date time.Time) (NoteFile, bool, error) {
	notes, err := listNotes(root, NoteFilter{Kind: "daily"})
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

func defaultDailyBody(date time.Time) string {
	return strings.TrimSpace(fmt.Sprintf(`
## Plan (%s)

- [ ] Top priority 1
- [ ] Top priority 2

## Notes

## Wins

## Follow-ups
`, date.Format("2006-01-02")))
}

func openNoteInEditor(path string) error {
	editor := resolveEditor()
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return errors.New("invalid editor command")
	}

	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open editor %q: %w", editor, err)
	}
	return nil
}

func validateVault(root string) (validationReport, error) {
	paths, err := collectNotePaths(filepath.Join(root, "notes"))
	if err != nil {
		return validationReport{}, err
	}

	report := validationReport{
		Total:    len(paths),
		Warnings: []string{},
		Errors:   []string{},
	}

	seenIDs := make(map[string]string)
	for _, path := range paths {
		note, readErr := readNote(path)
		relPath := toRelOrAbs(root, path)
		if readErr != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", relPath, readErr))
			continue
		}

		if first, exists := seenIDs[note.ID]; exists {
			report.Errors = append(report.Errors, fmt.Sprintf("duplicate id %s: %s and %s", note.ID, first, relPath))
		} else {
			seenIDs[note.ID] = relPath
		}

		expected, expectedErr := resolveNotePath(root, note)
		if expectedErr != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", relPath, expectedErr))
			continue
		}
		same, sameErr := sameFilePath(path, expected)
		if sameErr != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", relPath, sameErr))
			continue
		}
		if !same {
			report.Warnings = append(report.Warnings, fmt.Sprintf("%s expected at %s", relPath, toRelOrAbs(root, expected)))
		}
	}

	sort.Strings(report.Errors)
	sort.Strings(report.Warnings)

	return report, nil
}
