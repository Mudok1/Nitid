package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nitid/internal/core"
)

func runInit(args []string) error {
	target := "."
	if len(args) > 1 {
		return errors.New("init accepts at most one path argument")
	}
	if len(args) == 1 {
		target = args[0]
	}

	root, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	if err := core.New(root).Init(); err != nil {
		return err
	}

	fmt.Printf("initialized nitid vault at %s\n", root)
	return nil
}

func runCapture(args []string) error {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	title := fs.String("title", "", "note title")
	domain := fs.String("domain", "", "domain id")
	tags := fs.String("tags", "", "comma-separated tags")
	kind := fs.String("kind", "note", "note kind")

	if err := fs.Parse(args); err != nil {
		return err
	}

	body, err := resolveBody(fs.Args())
	if err != nil {
		return err
	}

	if strings.TrimSpace(*title) == "" {
		*title = inferTitle(body)
	}

	noteKind := strings.ToLower(strings.TrimSpace(*kind))
	if !isAllowedKind(noteKind) {
		return fmt.Errorf("invalid kind %q", *kind)
	}

	now := time.Now().UTC()
	note := Note{
		ID:        newULID(now),
		Title:     strings.TrimSpace(*title),
		CreatedAt: now,
		UpdatedAt: now,
		Domain:    strings.TrimSpace(*domain),
		Tags:      parseCSV(*tags),
		Kind:      noteKind,
		Links:     []string{},
		Body:      strings.TrimSpace(body),
	}

	if err := validateNoteForWrite(note); err != nil {
		return err
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	relPath, err := svc.Create(note)
	if err != nil {
		return err
	}

	fmt.Printf("saved %s\n", relPath)
	return nil
}

func resolveBody(args []string) (string, error) {
	if len(args) > 0 {
		return strings.TrimSpace(strings.Join(args, " ")), nil
	}

	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (stdinInfo.Mode() & os.ModeCharDevice) == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		text := strings.TrimSpace(string(b))
		if text != "" {
			return text, nil
		}
	}

	return "", errors.New("capture requires text argument or piped stdin")
}

func inferTitle(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return "Untitled"
	}

	line := strings.Split(body, "\n")[0]
	line = strings.TrimSpace(line)
	if len(line) <= 72 {
		return line
	}
	return strings.TrimSpace(line[:72])
}
