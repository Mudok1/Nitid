package cli

import (
	"errors"
	"fmt"
	"strings"

	"nitid/internal/core"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type notesLoadedMsg struct {
	notes []core.NoteFile
	err   error
}

type opDoneMsg struct {
	status string
	err    error
}

type editorDoneMsg struct {
	err error
}

type tuiModel struct {
	svc             *core.Service
	notes           []core.NoteFile
	selected        int
	pendingSelectID string
	width           int
	height          int
	status          string
	commandInput    textinput.Model
	mode            string
	confirmArchive  bool
	activeQuery     string
	loading         bool
}

func runTUI(args []string) error {
	if len(args) > 0 {
		return errors.New("tui does not accept arguments")
	}

	svc, err := newCoreService()
	if err != nil {
		return err
	}

	program := tea.NewProgram(newTUIModel(svc), tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func newTUIModel(svc *core.Service) tuiModel {
	input := textinput.New()
	input.Prompt = ":"
	input.CharLimit = 256

	return tuiModel{
		svc:          svc,
		notes:        []core.NoteFile{},
		selected:     0,
		status:       "loading notes...",
		commandInput: input,
		mode:         "normal",
		loading:      true,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return loadListCmd(m.svc)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		return m, nil
	case notesLoadedMsg:
		m.loading = false
		if typed.err != nil {
			m.status = fmt.Sprintf("error: %v", typed.err)
			return m, nil
		}

		m.notes = typed.notes
		m.reselectPending()
		m.clampSelection()
		if len(m.notes) == 0 {
			if m.activeQuery != "" {
				m.status = fmt.Sprintf("no notes found for query: %q", m.activeQuery)
			} else {
				m.status = "no notes found"
			}
		}
		return m, nil
	case opDoneMsg:
		if typed.err != nil {
			m.status = fmt.Sprintf("error: %v", typed.err)
			m.confirmArchive = false
			return m, nil
		}
		m.status = typed.status
		m.confirmArchive = false
		return m, loadListCmd(m.svc)
	case editorDoneMsg:
		if typed.err != nil {
			m.status = fmt.Sprintf("editor error: %v", typed.err)
		} else {
			m.status = "editor closed"
		}
		return m, loadListCmd(m.svc)
	case tea.KeyMsg:
		if m.mode == "command" {
			return m.updateCommandMode(typed)
		}
		return m.updateNormalMode(typed)
	}

	if m.mode == "command" {
		var cmd tea.Cmd
		m.commandInput, cmd = m.commandInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m tuiModel) updateCommandMode(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		command := strings.TrimSpace(m.commandInput.Value())
		m.commandInput.SetValue("")
		m.commandInput.Blur()
		m.mode = "normal"
		return m.execCommand(command)
	case "esc":
		m.commandInput.SetValue("")
		m.commandInput.Blur()
		m.mode = "normal"
		m.status = ""
		return m, nil
	default:
		var cmd tea.Cmd
		m.commandInput, cmd = m.commandInput.Update(key)
		return m, cmd
	}
}

func (m tuiModel) updateNormalMode(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmArchive {
		switch key.String() {
		case "y":
			note, ok := m.selectedNote()
			if !ok {
				m.confirmArchive = false
				m.status = "no note selected"
				return m, nil
			}
			m.pendingSelectID = note.Note.ID
			return m, archiveNoteCmd(m.svc, note.Note.ID)
		case "n", "esc":
			m.confirmArchive = false
			m.status = "archive cancelled"
			return m, nil
		}
	}

	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "j", "down":
		m.selected++
		m.clampSelection()
		return m, nil
	case "k", "up":
		m.selected--
		m.clampSelection()
		return m, nil
	case "g":
		m.selected = 0
		m.clampSelection()
		return m, nil
	case "G":
		m.selected = len(m.notes) - 1
		m.clampSelection()
		return m, nil
	case "r":
		m.activeQuery = ""
		m.status = "refreshed"
		return m, loadListCmd(m.svc)
	case ":":
		m.mode = "command"
		m.commandInput.SetValue("")
		m.commandInput.Focus()
		return m, nil
	case "/":
		m.mode = "command"
		m.commandInput.SetValue("find ")
		m.commandInput.CursorEnd()
		m.commandInput.Focus()
		return m, nil
	case "a":
		note, ok := m.selectedNote()
		if !ok {
			m.status = "no note selected"
			return m, nil
		}
		m.confirmArchive = true
		m.status = fmt.Sprintf("archive %s? press y/n", shortID(note.Note.ID))
		return m, nil
	case "e":
		note, ok := m.selectedNote()
		if !ok {
			m.status = "no note selected"
			return m, nil
		}
		cmd, err := core.EditorCommand(note.Path)
		if err != nil {
			m.status = fmt.Sprintf("error: %v", err)
			return m, nil
		}
		return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
			return editorDoneMsg{err: err}
		})
	}

	return m, nil
}

func (m tuiModel) execCommand(input string) (tea.Model, tea.Cmd) {
	input = strings.TrimSpace(input)
	if input == "" {
		m.status = ""
		return m, nil
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return m, nil
	}

	switch parts[0] {
	case "q", "quit":
		return m, tea.Quit
	case "help":
		m.status = "commands: ls, find <query>, move <domain>, tag add|rm <tag>, archive, quit"
		return m, nil
	case "ls":
		m.activeQuery = ""
		m.status = "listing all notes"
		return m, loadListCmd(m.svc)
	case "find":
		if len(parts) < 2 {
			m.status = "usage: find <query>"
			return m, nil
		}
		query := strings.TrimSpace(strings.Join(parts[1:], " "))
		m.activeQuery = query
		m.status = fmt.Sprintf("searching for %q", query)
		return m, findNotesCmd(m.svc, query)
	case "archive":
		note, ok := m.selectedNote()
		if !ok {
			m.status = "no note selected"
			return m, nil
		}
		m.confirmArchive = true
		m.status = fmt.Sprintf("archive %s? press y/n", shortID(note.Note.ID))
		return m, nil
	case "move":
		if len(parts) != 2 {
			m.status = "usage: move <domain>"
			return m, nil
		}
		note, ok := m.selectedNote()
		if !ok {
			m.status = "no note selected"
			return m, nil
		}
		m.pendingSelectID = note.Note.ID
		return m, moveNoteCmd(m.svc, note.Note.ID, parts[1])
	case "tag":
		if len(parts) != 3 {
			m.status = "usage: tag add|rm <tag>"
			return m, nil
		}
		note, ok := m.selectedNote()
		if !ok {
			m.status = "no note selected"
			return m, nil
		}
		m.pendingSelectID = note.Note.ID
		return m, tagNoteCmd(m.svc, note.Note.ID, parts[1], parts[2])
	default:
		m.status = fmt.Sprintf("unknown command: %s", input)
		return m, nil
	}
}

func (m tuiModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "starting ntd tui..."
	}

	contentHeight := m.height - 2
	if contentHeight < 8 {
		contentHeight = 8
	}

	leftW := m.width * 30 / 100
	rightW := m.width * 25 / 100
	centerW := m.width - leftW - rightW - 4
	if leftW < 28 {
		leftW = 28
	}
	if rightW < 28 {
		rightW = 28
	}
	if centerW < 32 {
		centerW = 32
	}

	panelStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	listPanel := panelStyle.Width(leftW).Height(contentHeight).Render(m.renderList(contentHeight - 2))
	previewPanel := panelStyle.Width(centerW).Height(contentHeight).Render(m.renderPreview(contentHeight - 2))
	metaPanel := panelStyle.Width(rightW).Height(contentHeight).Render(m.renderMeta(contentHeight - 2))

	body := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, previewPanel, metaPanel)
	status := m.status
	if m.mode == "command" {
		status = m.commandInput.View()
	}
	if strings.TrimSpace(status) == "" {
		status = "j/k move  / find  : commands  e edit  a archive  q quit"
	}

	return body + "\n" + status
}

func (m *tuiModel) clampSelection() {
	if len(m.notes) == 0 {
		m.selected = 0
		return
	}
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.notes) {
		m.selected = len(m.notes) - 1
	}
}

func (m *tuiModel) reselectPending() {
	if m.pendingSelectID == "" {
		return
	}
	for idx, item := range m.notes {
		if item.Note.ID == m.pendingSelectID {
			m.selected = idx
			break
		}
	}
	m.pendingSelectID = ""
}

func (m tuiModel) selectedNote() (core.NoteFile, bool) {
	if len(m.notes) == 0 {
		return core.NoteFile{}, false
	}
	if m.selected < 0 || m.selected >= len(m.notes) {
		return core.NoteFile{}, false
	}
	return m.notes[m.selected], true
}

func (m tuiModel) renderList(maxLines int) string {
	lines := []string{"Notes"}
	if m.activeQuery != "" {
		lines = append(lines, fmt.Sprintf("filter: %q", m.activeQuery))
	}

	if m.loading {
		lines = append(lines, "", "loading...")
		return strings.Join(lines, "\n")
	}

	if len(m.notes) == 0 {
		lines = append(lines, "", "(empty)")
		return strings.Join(lines, "\n")
	}

	for idx, item := range m.notes {
		marker := " "
		if idx == m.selected {
			marker = ">"
		}
		line := fmt.Sprintf("%s @%d %-8s %s", marker, idx+1, item.Note.Status, truncate(item.Note.Title, 42))
		if idx == m.selected {
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}
		lines = append(lines, line)
		if len(lines) >= maxLines {
			break
		}
	}

	return strings.Join(lines, "\n")
}

func (m tuiModel) renderPreview(maxLines int) string {
	noteFile, ok := m.selectedNote()
	if !ok {
		return "Preview\n\nNo note selected"
	}

	lines := []string{"Preview", "", truncate(noteFile.Note.Title, 100), ""}
	body := strings.TrimSpace(noteFile.Note.Body)
	if body == "" {
		lines = append(lines, "(empty body)")
		return strings.Join(lines, "\n")
	}

	bodyLines := strings.Split(body, "\n")
	space := maxLines - len(lines)
	if space < 1 {
		space = 1
	}
	if len(bodyLines) > space {
		bodyLines = bodyLines[:space]
	}
	lines = append(lines, bodyLines...)
	return strings.Join(lines, "\n")
}

func (m tuiModel) renderMeta(maxLines int) string {
	noteFile, ok := m.selectedNote()
	if !ok {
		return "Meta\n\nNo note selected"
	}

	lines := []string{
		"Meta",
		"",
		fmt.Sprintf("id: %s", noteFile.Note.ID),
		fmt.Sprintf("status: %s", noteFile.Note.Status),
		fmt.Sprintf("kind: %s", noteFile.Note.Kind),
		fmt.Sprintf("domain: %s", displayDomain(noteFile.Note.Domain)),
		fmt.Sprintf("tags: %s", displayTags(noteFile.Note.Tags)),
		fmt.Sprintf("updated: %s", noteFile.Note.UpdatedAt.Format("2006-01-02 15:04")),
		"",
		"Actions",
		"- e edit",
		"- a archive",
		"- :move <domain>",
		"- :tag add|rm <tag>",
		"- :find <query>",
	}

	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	return strings.Join(lines, "\n")
}

func loadListCmd(svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		notes, err := svc.List(core.NoteFilter{}, "updated", false)
		return notesLoadedMsg{notes: notes, err: err}
	}
}

func findNotesCmd(svc *core.Service, query string) tea.Cmd {
	return func() tea.Msg {
		notes, err := svc.Find(query, core.NoteFilter{}, 200)
		return notesLoadedMsg{notes: notes, err: err}
	}
}

func moveNoteCmd(svc *core.Service, selector, domain string) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.Move(selector, domain)
		if err != nil {
			return opDoneMsg{err: err}
		}
		return opDoneMsg{status: fmt.Sprintf("moved %s -> %s", result.NoteID, result.RelPath)}
	}
}

func tagNoteCmd(svc *core.Service, selector, action, tag string) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.Tag(selector, action, tag)
		if err != nil {
			return opDoneMsg{err: err}
		}
		return opDoneMsg{status: fmt.Sprintf("updated tags for %s -> %s", result.NoteID, result.RelPath)}
	}
}

func archiveNoteCmd(svc *core.Service, selector string) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.Archive(selector)
		if err != nil {
			return opDoneMsg{err: err}
		}
		return opDoneMsg{status: fmt.Sprintf("archived %s -> %s", result.NoteID, result.RelPath)}
	}
}
