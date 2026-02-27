package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nitid/internal/core"
)

func runNew(args []string) error {
	if len(args) == 0 {
		return errors.New("new requires a template name")
	}

	templateName := strings.ToLower(strings.TrimSpace(args[0]))
	if _, ok := templateDefinitions()[templateName]; !ok {
		return fmt.Errorf("unknown template %q", templateName)
	}

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	title := fs.String("title", "", "note title")
	domain := fs.String("domain", "", "domain id")
	tags := fs.String("tags", "", "comma-separated tags")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	extraText := strings.TrimSpace(strings.Join(fs.Args(), " "))
	def := templateDefinitions()[templateName]
	body := templateBody(def, extraText)

	if strings.TrimSpace(*title) == "" {
		*title = def.DefaultTitle
		if templateName == "note" {
			*title = inferTitle(body)
		}
	}

	now := time.Now().UTC()
	note := Note{
		ID:        newULID(now),
		Title:     strings.TrimSpace(*title),
		CreatedAt: now,
		UpdatedAt: now,
		Domain:    strings.TrimSpace(*domain),
		Tags:      parseCSV(*tags),
		Kind:      def.Kind,
		Links:     []string{},
		Body:      body,
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

func runDaily(args []string) error {
	fs := flag.NewFlagSet("daily", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dateArg := fs.String("date", "", "date in YYYY-MM-DD")
	openEditor := fs.Bool("edit", false, "open daily note after create/find")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) > 0 {
		return errors.New("daily does not accept positional arguments")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}
	if err := svc.Init(); err != nil {
		return err
	}

	targetDate, err := resolveDailyDate(*dateArg)
	if err != nil {
		return err
	}

	existing, found, err := svc.FindDailyByDate(targetDate)
	if err != nil {
		return err
	}

	if found {
		fmt.Printf("daily note already exists: %s\n", existing.RelPath)
		if *openEditor {
			return core.OpenInEditor(existing.Path)
		}
		return nil
	}

	note := Note{
		ID:        newULID(targetDate),
		Title:     fmt.Sprintf("Daily %s", targetDate.Format("2006-01-02")),
		CreatedAt: targetDate,
		UpdatedAt: targetDate,
		Domain:    "",
		Tags:      []string{"daily"},
		Kind:      "daily",
		Links:     []string{},
		Body:      defaultDailyBody(targetDate),
	}

	rel, err := svc.Create(note)
	if err != nil {
		return err
	}
	fmt.Printf("saved %s\n", rel)

	if *openEditor {
		return core.OpenInEditor(filepath.Join(svc.Root(), filepath.FromSlash(rel)))
	}

	return nil
}

func runTemplates(args []string) error {
	defs := templateDefinitions()
	if len(args) == 0 {
		fmt.Println("available templates:")
		names := make([]string, 0, len(defs))
		for name := range defs {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Printf("- %s\n", name)
		}
		return nil
	}

	if len(args) == 2 && args[0] == "show" {
		name := strings.ToLower(strings.TrimSpace(args[1]))
		def, ok := defs[name]
		if !ok {
			return fmt.Errorf("unknown template %q", name)
		}
		fmt.Printf("template: %s\n", name)
		fmt.Printf("kind: %s\n\n", def.Kind)
		fmt.Println(strings.TrimSpace(def.BodyTemplate))
		return nil
	}

	return errors.New("templates usage: ntd templates OR ntd templates show <name>")
}
