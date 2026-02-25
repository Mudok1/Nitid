package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	exitOK    = 0
	exitError = 1
)

func run() int {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		return exitOK
	}

	var err error
	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return exitOK
	case "init":
		err = runInit(args[1:])
	case "capture":
		err = runCapture(args[1:])
	case "ls":
		err = runList(args[1:])
	case "move":
		err = runMove(args[1:])
	case "tag":
		err = runTag(args[1:])
	case "archive":
		err = runArchive(args[1:])
	case "show":
		err = runShow(args[1:])
	case "completion":
		err = runCompletion(args[1:])
	case "__complete_ids":
		err = runCompleteIDs(args[1:])
	default:
		err = fmt.Errorf("unknown command %q", args[0])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ntd: %v\n", err)
		return exitError
	}

	return exitOK
}

func printUsage() {
	fmt.Println("Nitid (ntd) - developer second brain CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ntd init [path]")
	fmt.Println("  ntd capture [text] [--title \"...\"] [--domain <id>] [--tags t1,t2] [--kind note|adr|snippet|daily]")
	fmt.Println("  ntd ls [--domain <id>] [--tag <tag>] [--status inbox|active|archived] [--kind note|adr|snippet|daily]")
	fmt.Println("  ntd move <id|#ref> --domain <id>")
	fmt.Println("  ntd tag <id|#ref> add|rm <tag>")
	fmt.Println("  ntd archive <id|#ref>")
	fmt.Println("  ntd show <id|#ref>")
	fmt.Println("  ntd completion bash")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ntd init .")
	fmt.Println("  ntd capture \"Investigate goroutine leak in worker pool\"")
	fmt.Println("  ntd ls --status inbox")
	fmt.Println("  ntd ls --long")
	fmt.Println("  ntd move #1 --domain engineering")
	fmt.Println("  ntd tag #1 add concurrency")
	fmt.Println("  ntd archive #1")
	fmt.Println("  ntd show #1")
	fmt.Println("  source <(ntd completion bash)")
	fmt.Println("  ntd capture --domain engineering --tags go,debug --title \"Worker leak\" \"Found issue in retry loop\"")
}

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

	if err := createVaultStructure(root); err != nil {
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

	root, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	if err := createVaultStructure(root); err != nil {
		return err
	}

	relPath, err := writeNote(root, note)
	if err != nil {
		return err
	}

	fmt.Printf("saved %s\n", relPath)
	return nil
}

func runList(args []string) error {
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	domain := fs.String("domain", "", "filter by domain")
	tag := fs.String("tag", "", "filter by tag")
	status := fs.String("status", "", "filter by status")
	kind := fs.String("kind", "", "filter by kind")
	long := fs.Bool("long", false, "print detailed rows")

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

	root, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	notes, err := listNotes(root, NoteFilter{
		Domain: domainFilter,
		Tag:    tagFilter,
		Status: statusFilter,
		Kind:   kindFilter,
	})
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
			fmt.Sprintf("#%d", idx+1),
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

func runMove(args []string) error {
	if len(args) == 0 {
		return errors.New("move requires exactly one <id|#ref> argument")
	}
	selector := strings.TrimSpace(args[0])
	if selector == "" {
		return errors.New("move requires exactly one <id|#ref> argument")
	}

	fs := flag.NewFlagSet("move", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	domain := fs.String("domain", "", "target domain")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("unexpected arguments for move")
	}

	domainID := strings.ToLower(strings.TrimSpace(*domain))
	if !domainIDPattern.MatchString(domainID) {
		return fmt.Errorf("invalid domain %q: use lowercase kebab-case", domainID)
	}

	root, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	noteFile, err := findNoteBySelector(root, selector)
	if err != nil {
		return err
	}

	note := noteFile.Note
	note.Domain = domainID
	note.Status = statusActive
	note.UpdatedAt = time.Now().UTC()

	rel, err := saveNote(root, noteFile.Path, note)
	if err != nil {
		return err
	}

	fmt.Printf("moved %s -> %s\n", note.ID, rel)
	return nil
}

func runTag(args []string) error {
	if len(args) != 3 {
		return errors.New("tag usage: ntd tag <id|#ref> add|rm <tag>")
	}

	selector := strings.TrimSpace(args[0])
	action := strings.ToLower(strings.TrimSpace(args[1]))
	tag := strings.ToLower(strings.TrimSpace(args[2]))

	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("invalid tag %q: use lowercase kebab-case", tag)
	}
	if action != "add" && action != "rm" {
		return fmt.Errorf("invalid tag action %q", action)
	}

	root, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	noteFile, err := findNoteBySelector(root, selector)
	if err != nil {
		return err
	}

	note := noteFile.Note
	note.Tags = updateTags(note.Tags, action, tag)
	note.UpdatedAt = time.Now().UTC()

	rel, err := saveNote(root, noteFile.Path, note)
	if err != nil {
		return err
	}

	fmt.Printf("updated tags for %s -> %s\n", note.ID, rel)
	return nil
}

func runArchive(args []string) error {
	if len(args) != 1 {
		return errors.New("archive requires exactly one <id|#ref> argument")
	}

	root, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	noteFile, err := findNoteBySelector(root, strings.TrimSpace(args[0]))
	if err != nil {
		return err
	}

	note := noteFile.Note
	note.Status = statusArchived
	note.UpdatedAt = time.Now().UTC()

	rel, err := saveNote(root, noteFile.Path, note)
	if err != nil {
		return err
	}

	fmt.Printf("archived %s -> %s\n", note.ID, rel)
	return nil
}

func runShow(args []string) error {
	if len(args) != 1 {
		return errors.New("show requires exactly one <id|#ref> argument")
	}

	root, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	noteFile, err := findNoteBySelector(root, strings.TrimSpace(args[0]))
	if err != nil {
		return err
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

func runCompletion(args []string) error {
	if len(args) != 1 || strings.TrimSpace(args[0]) != "bash" {
		return errors.New("completion usage: ntd completion bash")
	}
	fmt.Print(bashCompletionScript())
	return nil
}

func runCompleteIDs(args []string) error {
	if len(args) > 0 {
		return errors.New("__complete_ids does not accept arguments")
	}
	root, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	notes, err := listNotes(root, NoteFilter{})
	if err != nil {
		return err
	}
	for i, item := range notes {
		fmt.Printf("#%d\n", i+1)
		fmt.Println(item.Note.ID)
	}
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

func updateTags(tags []string, action, value string) []string {
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

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max <= 1 {
		return value[:max]
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func displayDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return "-"
	}
	return domain
}

func displayTags(tags []string) string {
	if len(tags) == 0 {
		return "-"
	}
	return strings.Join(tags, ",")
}

func displayTagsCompact(tags []string) string {
	if len(tags) == 0 {
		return "-"
	}
	if len(tags) <= 2 {
		return strings.Join(tags, ",")
	}
	return fmt.Sprintf("%s,%s,+%d", tags[0], tags[1], len(tags)-2)
}

func bashCompletionScript() string {
	return strings.TrimSpace(`
_ntd_complete() {
  local cur prev cmd
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  cmd="${COMP_WORDS[1]}"

  if [[ ${COMP_CWORD} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "help init capture ls move tag archive show completion" -- "${cur}") )
    return 0
  fi

  case "${cmd}" in
    move|tag|archive|show)
      if [[ ${COMP_CWORD} -eq 2 ]]; then
        COMPREPLY=( $(compgen -W "$(ntd __complete_ids 2>/dev/null)" -- "${cur}") )
        return 0
      fi
      ;;
    tag)
      if [[ ${COMP_CWORD} -eq 3 ]]; then
        COMPREPLY=( $(compgen -W "add rm" -- "${cur}") )
        return 0
      fi
      ;;
    completion)
      if [[ ${COMP_CWORD} -eq 2 ]]; then
        COMPREPLY=( $(compgen -W "bash" -- "${cur}") )
        return 0
      fi
      ;;
  esac
}

complete -F _ntd_complete ntd
`) + "\n"
}
