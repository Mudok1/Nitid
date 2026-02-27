package cli

import (
	"fmt"
	"os"
)

const (
	exitOK     = 0
	exitError  = 1
	appVersion = "0.1.0-dev"
)

func run() int {
	return Run(os.Args[1:])
}

func Run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return exitOK
	}

	var err error
	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return exitOK
	case "version":
		fmt.Printf("ntd %s\n", appVersion)
		return exitOK
	case "init":
		err = runInit(args[1:])
	case "capture":
		err = runCapture(args[1:])
	case "new":
		err = runNew(args[1:])
	case "daily":
		err = runDaily(args[1:])
	case "templates":
		err = runTemplates(args[1:])
	case "ls":
		err = runList(args[1:])
	case "find":
		err = runFind(args[1:])
	case "move":
		err = runMove(args[1:])
	case "tag":
		err = runTag(args[1:])
	case "archive":
		err = runArchive(args[1:])
	case "delete":
		err = runDelete(args[1:])
	case "show":
		err = runShow(args[1:])
	case "edit":
		err = runEdit(args[1:])
	case "clean":
		err = runClean(args[1:])
	case "validate":
		err = runValidate(args[1:])
	case "doctor":
		err = runDoctor(args[1:])
	case "tui":
		err = runTUI(args[1:])
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
	fmt.Println("  ntd version")
	fmt.Println("  ntd init [path]")
	fmt.Println("  ntd capture [text] [--title \"...\"] [--domain <id>] [--tags t1,t2] [--kind note|adr|snippet|daily]")
	fmt.Println("  ntd new <template> [text] [--title \"...\"] [--domain <id>] [--tags t1,t2]")
	fmt.Println("  ntd daily [--date YYYY-MM-DD] [--edit]")
	fmt.Println("  ntd templates")
	fmt.Println("  ntd templates show <name>")
	fmt.Println("  ntd ls [--domain <id>] [--tag <tag>] [--status inbox|active|archived] [--kind note|adr|snippet|daily] [--sort updated|created|title|id] [--asc]")
	fmt.Println("  ntd find <query> [--domain <id>] [--tag <tag>] [--status inbox|active|archived] [--kind note|adr|snippet|daily] [--limit N]")
	fmt.Println("  ntd move <id|@ref> --domain <id>")
	fmt.Println("  ntd tag <id|@ref> add|rm <tag>")
	fmt.Println("  ntd archive <id|@ref>")
	fmt.Println("  ntd delete <id|@ref> --yes")
	fmt.Println("  ntd show <id|@ref> [--raw]")
	fmt.Println("  ntd edit <id|@ref>")
	fmt.Println("  ntd clean [--dry-run]")
	fmt.Println("  ntd validate")
	fmt.Println("  ntd doctor")
	fmt.Println("  ntd tui")
	fmt.Println("  ntd completion bash")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ntd version")
	fmt.Println("  ntd init .")
	fmt.Println("  ntd capture \"Investigate goroutine leak in worker pool\"")
	fmt.Println("  ntd new adr --title \"Use ULID for note IDs\"")
	fmt.Println("  ntd daily --edit")
	fmt.Println("  ntd templates")
	fmt.Println("  ntd ls --status inbox --sort updated")
	fmt.Println("  ntd find worker --limit 10")
	fmt.Println("  ntd ls --long")
	fmt.Println("  ntd move @1 --domain engineering")
	fmt.Println("  ntd tag @1 add concurrency")
	fmt.Println("  ntd archive @1")
	fmt.Println("  ntd delete @1 --yes")
	fmt.Println("  ntd show @1")
	fmt.Println("  ntd show @1 --raw")
	fmt.Println("  ntd edit @1")
	fmt.Println("  ntd clean")
	fmt.Println("  ntd validate")
	fmt.Println("  ntd doctor")
	fmt.Println("  ntd tui")
	fmt.Println("  source <(ntd completion bash)")
	fmt.Println("  ntd capture --domain engineering --tags go,debug --title \"Worker leak\" \"Found issue in retry loop\"")
}
