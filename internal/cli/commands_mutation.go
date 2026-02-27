package cli

import (
	"errors"
	"fmt"
	"strings"

	"nitid/internal/core"
)

func runMove(args []string) error {
	if len(args) == 0 {
		return errors.New("move requires exactly one <id|@ref> argument")
	}
	selector := strings.TrimSpace(args[0])
	if selector == "" {
		return errors.New("move requires exactly one <id|@ref> argument")
	}

	domainID := ""
	for i := 1; i < len(args); i++ {
		if args[i] == "--domain" && i+1 < len(args) {
			domainID = strings.ToLower(strings.TrimSpace(args[i+1]))
			i++
			continue
		}
		return errors.New("move usage: ntd move <id|@ref> --domain <domain_id>")
	}

	if !domainIDPattern.MatchString(domainID) {
		return fmt.Errorf("invalid domain %q: use lowercase kebab-case", domainID)
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	result, err := svc.Move(selector, domainID)
	if err != nil {
		return err
	}

	fmt.Printf("moved %s -> %s\n", result.NoteID, result.RelPath)
	return nil
}

func runTag(args []string) error {
	if len(args) != 3 {
		return errors.New("tag usage: ntd tag <id|@ref> add|rm <tag>")
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

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	result, err := svc.Tag(selector, action, tag)
	if err != nil {
		return err
	}

	fmt.Printf("updated tags for %s -> %s\n", result.NoteID, result.RelPath)
	return nil
}

func runArchive(args []string) error {
	if len(args) != 1 {
		return errors.New("archive requires exactly one <id|@ref> argument")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	result, err := svc.Archive(strings.TrimSpace(args[0]))
	if err != nil {
		return err
	}

	fmt.Printf("archived %s -> %s\n", result.NoteID, result.RelPath)
	return nil
}

func runDelete(args []string) error {
	if len(args) < 1 {
		return errors.New("delete usage: ntd delete <id|@ref> --yes")
	}

	selector := strings.TrimSpace(args[0])
	if selector == "" {
		return errors.New("delete usage: ntd delete <id|@ref> --yes")
	}

	confirmed := false
	for _, arg := range args[1:] {
		if arg == "--yes" || arg == "-y" {
			confirmed = true
			continue
		}
		return errors.New("delete usage: ntd delete <id|@ref> --yes")
	}

	if !confirmed {
		return errors.New("delete requires confirmation flag --yes")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	result, err := svc.Delete(selector)
	if err != nil {
		return err
	}

	fmt.Printf("deleted %s -> %s\n", result.NoteID, result.RelPath)
	return nil
}

func runEdit(args []string) error {
	if len(args) != 1 {
		return errors.New("edit requires exactly one <id|@ref> argument")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	return svc.Edit(strings.TrimSpace(args[0]))
}

func updateTags(tags []string, action, value string) []string {
	return core.UpdateTags(tags, action, value)
}

func resolveEditor() string {
	return core.ResolveEditor()
}
