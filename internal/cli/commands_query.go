package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"nitid/internal/core"
)

func runList(args []string) error {
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	domain := fs.String("domain", "", "filter by domain")
	tag := fs.String("tag", "", "filter by tag")
	status := fs.String("status", "", "filter by status")
	kind := fs.String("kind", "", "filter by kind")
	long := fs.Bool("long", false, "print detailed rows")
	sortBy := fs.String("sort", "updated", "sort by updated|created|title|id")
	asc := fs.Bool("asc", false, "sort ascending")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) > 0 {
		return errors.New("ls does not accept positional arguments")
	}

	statusFilter := strings.ToLower(strings.TrimSpace(*status))
	if statusFilter != "" {
		if _, ok := allowedStatuses[statusFilter]; !ok {
			return fmt.Errorf("invalid status %q", statusFilter)
		}
	}

	domainFilter := strings.ToLower(strings.TrimSpace(*domain))
	if domainFilter != "" && !domainIDPattern.MatchString(domainFilter) {
		return fmt.Errorf("invalid domain %q: use lowercase kebab-case", domainFilter)
	}

	tagFilter := strings.ToLower(strings.TrimSpace(*tag))
	if tagFilter != "" && !tagPattern.MatchString(tagFilter) {
		return fmt.Errorf("invalid tag %q: use lowercase kebab-case", tagFilter)
	}

	kindFilter := strings.ToLower(strings.TrimSpace(*kind))
	if kindFilter != "" && !isAllowedKind(kindFilter) {
		return fmt.Errorf("invalid kind %q", kindFilter)
	}

	sortMode := strings.ToLower(strings.TrimSpace(*sortBy))
	if sortMode != "updated" && sortMode != "created" && sortMode != "title" && sortMode != "id" {
		return fmt.Errorf("invalid sort %q", sortMode)
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	notes, err := svc.List(core.NoteFilter{
		Domain: domainFilter,
		Tag:    tagFilter,
		Status: statusFilter,
		Kind:   kindFilter,
	}, sortMode, *asc)
	if err != nil {
		return err
	}

	if len(notes) == 0 {
		fmt.Println("no notes found")
		return nil
	}

	if *long {
		for _, item := range notes {
			domainLabel := displayDomain(item.Note.Domain)
			tags := displayTags(item.Note.Tags)
			fmt.Printf("%s %s [%s/%s] domain=%s tags=%s %s\n", item.Note.ID, item.RelPath, item.Note.Status, item.Note.Kind, domainLabel, tags, item.Note.Title)
		}
		return nil
	}

	prefixes := uniqueIDPrefixes(notes, 8)
	fmt.Printf("%-5s  %-12s  %-8s  %-8s  %-16s  %-18s  %s\n", "REF", "ID", "STATUS", "KIND", "DOMAIN", "TAGS", "TITLE")
	fmt.Printf("%-5s  %-12s  %-8s  %-8s  %-16s  %-18s  %s\n", strings.Repeat("-", 5), strings.Repeat("-", 12), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 16), strings.Repeat("-", 18), strings.Repeat("-", 30))
	for idx, item := range notes {
		short := prefixes[item.Note.ID]
		if short == "" {
			short = shortID(item.Note.ID)
		}
		fmt.Printf("%-5s  %-12s  %-8s  %-8s  %-16s  %-18s  %s\n",
			fmt.Sprintf("@%d", idx+1),
			short,
			item.Note.Status,
			item.Note.Kind,
			displayDomain(item.Note.Domain),
			displayTagsCompact(item.Note.Tags),
			truncate(item.Note.Title, 72),
		)
	}

	return nil
}

func runFind(args []string) error {
	domainFilter := ""
	tagFilter := ""
	statusFilter := ""
	kindFilter := ""
	limit := 20
	queryParts := make([]string, 0)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--domain":
			if i+1 >= len(args) {
				return errors.New("--domain requires a value")
			}
			domainFilter = strings.ToLower(strings.TrimSpace(args[i+1]))
			i++
		case "--tag":
			if i+1 >= len(args) {
				return errors.New("--tag requires a value")
			}
			tagFilter = strings.ToLower(strings.TrimSpace(args[i+1]))
			i++
		case "--status":
			if i+1 >= len(args) {
				return errors.New("--status requires a value")
			}
			statusFilter = strings.ToLower(strings.TrimSpace(args[i+1]))
			i++
		case "--kind":
			if i+1 >= len(args) {
				return errors.New("--kind requires a value")
			}
			kindFilter = strings.ToLower(strings.TrimSpace(args[i+1]))
			i++
		case "--limit":
			if i+1 >= len(args) {
				return errors.New("--limit requires a value")
			}
			parsed, err := strconv.Atoi(strings.TrimSpace(args[i+1]))
			if err != nil {
				return errors.New("--limit must be a number")
			}
			limit = parsed
			i++
		default:
			queryParts = append(queryParts, arg)
		}
	}

	if len(queryParts) == 0 {
		return errors.New("find requires a query string")
	}

	query := strings.ToLower(strings.TrimSpace(strings.Join(queryParts, " ")))
	if query == "" {
		return errors.New("find query cannot be empty")
	}
	if limit < 1 {
		return errors.New("limit must be at least 1")
	}

	if statusFilter != "" {
		if _, ok := allowedStatuses[statusFilter]; !ok {
			return fmt.Errorf("invalid status %q", statusFilter)
		}
	}

	if domainFilter != "" && !domainIDPattern.MatchString(domainFilter) {
		return fmt.Errorf("invalid domain %q: use lowercase kebab-case", domainFilter)
	}

	if tagFilter != "" && !tagPattern.MatchString(tagFilter) {
		return fmt.Errorf("invalid tag %q: use lowercase kebab-case", tagFilter)
	}

	if kindFilter != "" && !isAllowedKind(kindFilter) {
		return fmt.Errorf("invalid kind %q", kindFilter)
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	matches, err := svc.Find(query, core.NoteFilter{Domain: domainFilter, Tag: tagFilter, Status: statusFilter, Kind: kindFilter}, limit)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		fmt.Println("no matching notes found")
		return nil
	}

	prefixes := uniqueIDPrefixes(matches, 8)
	fmt.Printf("%-5s  %-12s  %-8s  %-16s  %s\n", "REF", "ID", "STATUS", "DOMAIN", "TITLE")
	fmt.Printf("%-5s  %-12s  %-8s  %-16s  %s\n", strings.Repeat("-", 5), strings.Repeat("-", 12), strings.Repeat("-", 8), strings.Repeat("-", 16), strings.Repeat("-", 30))
	for idx, item := range matches {
		idPrefix := prefixes[item.Note.ID]
		if idPrefix == "" {
			idPrefix = shortID(item.Note.ID)
		}
		fmt.Printf("%-5s  %-12s  %-8s  %-16s  %s\n",
			fmt.Sprintf("@%d", idx+1),
			idPrefix,
			item.Note.Status,
			displayDomain(item.Note.Domain),
			truncate(item.Note.Title, 72),
		)
	}

	return nil
}

func runShow(args []string) error {
	raw := false
	selector := ""
	for _, arg := range args {
		if arg == "--raw" {
			raw = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unknown flag %q", arg)
		}
		if selector == "" {
			selector = arg
			continue
		}
		return errors.New("show requires exactly one <id|@ref> argument")
	}
	if strings.TrimSpace(selector) == "" {
		return errors.New("show requires exactly one <id|@ref> argument")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	noteFile, err := svc.FindBySelector(strings.TrimSpace(selector))
	if err != nil {
		return err
	}

	if raw {
		rawContent, err := os.ReadFile(noteFile.Path)
		if err != nil {
			return err
		}
		fmt.Print(string(rawContent))
		return nil
	}

	note := noteFile.Note
	fmt.Printf("ID:      %s\n", note.ID)
	fmt.Printf("Title:   %s\n", note.Title)
	fmt.Printf("Status:  %s\n", note.Status)
	fmt.Printf("Kind:    %s\n", note.Kind)
	fmt.Printf("Domain:  %s\n", displayDomain(note.Domain))
	fmt.Printf("Tags:    %s\n", displayTags(note.Tags))
	fmt.Printf("Path:    %s\n", noteFile.RelPath)
	fmt.Printf("Created: %s\n", note.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", note.UpdatedAt.Format(time.RFC3339))
	fmt.Println()
	if strings.TrimSpace(note.Body) == "" {
		fmt.Println("(empty body)")
		return nil
	}
	fmt.Println(note.Body)
	return nil
}

func noteMatchesQuery(note Note, query string) bool {
	return core.NoteMatchesQuery(note, query)
}

func sortNotes(notes []NoteFile, mode string, asc bool) {
	core.SortNotes(notes, mode, asc)
}
