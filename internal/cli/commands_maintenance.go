package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runClean(args []string) error {
	dryRun := false
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			continue
		}
		return errors.New("clean usage: ntd clean [--dry-run]")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	targets, err := svc.FindEditorTempFiles()
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		fmt.Println("no editor temp files found")
		return nil
	}

	for _, path := range targets {
		rel, relErr := filepath.Rel(svc.Root(), path)
		if relErr != nil {
			rel = path
		}
		if dryRun {
			fmt.Printf("would remove %s\n", filepath.ToSlash(rel))
			continue
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		fmt.Printf("removed %s\n", filepath.ToSlash(rel))
	}

	return nil
}

func runValidate(args []string) error {
	if len(args) > 0 {
		return errors.New("validate does not accept arguments")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	report, err := svc.Validate()
	if err != nil {
		return err
	}
	if report.Total == 0 {
		fmt.Println("no notes found")
		return nil
	}

	fmt.Printf("validated %d notes\n", report.Total)
	if len(report.Warnings) > 0 {
		fmt.Printf("warnings: %d\n", len(report.Warnings))
		for _, w := range report.Warnings {
			fmt.Printf("- %s\n", w)
		}
	}

	if len(report.Errors) > 0 {
		fmt.Printf("errors: %d\n", len(report.Errors))
		for _, e := range report.Errors {
			fmt.Printf("- %s\n", e)
		}
		return errors.New("validation failed")
	}

	fmt.Println("validation passed")
	return nil
}

func runDoctor(args []string) error {
	if len(args) > 0 {
		return errors.New("doctor does not accept arguments")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	status := "ok"

	notesRoot := filepath.Join(svc.Root(), "notes")
	if info, statErr := os.Stat(notesRoot); statErr != nil || !info.IsDir() {
		fmt.Printf("[fail] notes directory missing: %s\n", notesRoot)
		status = "fail"
	} else {
		fmt.Printf("[ok] notes directory: %s\n", notesRoot)
	}

	editor := resolveEditor()
	editorBin := strings.Fields(editor)
	if len(editorBin) == 0 {
		fmt.Println("[warn] no editor configured")
		if status == "ok" {
			status = "warn"
		}
	} else {
		if _, lookErr := exec.LookPath(editorBin[0]); lookErr != nil {
			fmt.Printf("[warn] editor not found in PATH: %s\n", editorBin[0])
			if status == "ok" {
				status = "warn"
			}
		} else {
			fmt.Printf("[ok] editor: %s\n", editor)
		}
	}

	fmt.Println("[ok] completion command available: ntd completion bash")

	report, valErr := svc.Validate()
	if valErr != nil {
		fmt.Printf("[fail] validate failed: %v\n", valErr)
		status = "fail"
	} else {
		fmt.Printf("[ok] parsed notes: %d\n", report.Total)
		if len(report.Errors) > 0 {
			fmt.Printf("[fail] validation errors: %d\n", len(report.Errors))
			status = "fail"
		}
		if len(report.Warnings) > 0 {
			fmt.Printf("[warn] validation warnings: %d\n", len(report.Warnings))
			if status == "ok" {
				status = "warn"
			}
		}
	}

	switch status {
	case "ok":
		fmt.Println("doctor status: ok")
		return nil
	case "warn":
		fmt.Println("doctor status: warn")
		return nil
	default:
		fmt.Println("doctor status: fail")
		return errors.New("doctor found critical issues")
	}
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
	svc, err := newCoreService()
	if err != nil {
		return err
	}
	selectors, err := svc.CompleteSelectors()
	if err != nil {
		return err
	}
	for _, selector := range selectors {
		fmt.Println(selector)
	}
	return nil
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
	    COMPREPLY=( $(compgen -W "help version init capture new daily templates ls find move tag archive delete show edit clean validate doctor tui completion" -- "${cur}") )
	    return 0
	  fi

  case "${cmd}" in
    move|tag|archive|delete|show|edit)
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
