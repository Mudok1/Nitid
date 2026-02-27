package cli

import (
	"errors"
	"fmt"
	"strings"

	"nitid/internal/core"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tuiModeNormal  = "normal"
	tuiModeCommand = "command"
	tuiModeEdit    = "edit"
)

var (
	tuiBgColor     = lipgloss.Color("0")
	tuiFgColor     = lipgloss.Color("252")
	tuiMutedColor  = lipgloss.Color("244")
	tuiAccentColor = lipgloss.Color("81")
	tuiBorderColor = lipgloss.Color("240")
	tuiStatusBg    = lipgloss.Color("236")
)

type notesLoadedMsg struct {
	notes []core.NoteFile
	err   error
}

type opDoneMsg struct {
	status string
	err    error
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
	textarea        textarea.Model
	mode            string
	confirmArchive  bool
	activeQuery     string
	loading         bool
	editingNoteID   string
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

	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = true

	return tuiModel{
		svc:          svc,
		notes:        []core.NoteFile{},
		selected:     0,
		status:       "loading notes...",
		commandInput: input,
		textarea:     ta,
		mode:         tuiModeNormal,
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
		m.resizeEditor()
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
	case tea.KeyMsg:
		switch m.mode {
		case tuiModeCommand:
			return m.updateCommandMode(typed)
		case tuiModeEdit:
			return m.updateEditMode(typed)
		default:
			return m.updateNormalMode(typed)
		}
	}

	switch m.mode {
	case tuiModeCommand:
		var cmd tea.Cmd
		m.commandInput, cmd = m.commandInput.Update(msg)
		return m, cmd
	case tuiModeEdit:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m tuiModel) updateCommandMode(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		command := strings.TrimSpace(m.commandInput.Value())
		m.commandInput.SetValue("")
		m.commandInput.Blur()
		m.mode = tuiModeNormal
		return m.execCommand(command)
	case "esc":
		m.commandInput.SetValue("")
		m.commandInput.Blur()
		m.mode = tuiModeNormal
		m.status = ""
		return m, nil
	default:
		var cmd tea.Cmd
		m.commandInput, cmd = m.commandInput.Update(key)
		return m, cmd
	}
}

func (m tuiModel) updateEditMode(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+s":
		if strings.TrimSpace(m.editingNoteID) == "" {
			m.status = "no note selected"
			m.mode = tuiModeNormal
			return m, nil
		}

		noteID := m.editingNoteID
		body := m.textarea.Value()
		m.pendingSelectID = noteID
		m.mode = tuiModeNormal
		m.editingNoteID = ""
		m.status = "saving note..."
		return m, updateBodyCmd(m.svc, noteID, body)
	case "esc":
		m.mode = tuiModeNormal
		m.editingNoteID = ""
		m.status = "edit cancelled"
		return m, nil
	default:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(key)
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
		m.mode = tuiModeCommand
		m.commandInput.SetValue("")
		m.commandInput.Focus()
		return m, nil
	case "/":
		m.mode = tuiModeCommand
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
		return m.beginEdit()
	}

	return m, nil
}

func (m tuiModel) beginEdit() (tea.Model, tea.Cmd) {
	note, ok := m.selectedNote()
	if !ok {
		m.status = "no note selected"
		return m, nil
	}

	m.mode = tuiModeEdit
	m.editingNoteID = note.Note.ID
	m.confirmArchive = false
	m.status = fmt.Sprintf("editing %s (Ctrl+S save, Esc cancel)", shortID(note.Note.ID))

	ta := textarea.New()
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.ShowLineNumbers = true
	ta.SetValue(note.Note.Body)
	ta.Focus()
	m.textarea = ta
	m.resizeEditor()

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
		m.status = "commands: ls, find <query>, edit, move <domain>, tag add|rm <tag>, archive, quit"
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
	case "edit":
		return m.beginEdit()
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

	if m.mode == tuiModeEdit {
		return m.renderEditView()
	}

	contentHeight := maxInt(8, m.height-2)
	leftW, centerW, rightW := m.panelWidths()

	listPanel := m.renderPanel(leftW, contentHeight, m.renderList(contentHeight-2))
	previewPanel := m.renderPanel(centerW, contentHeight, m.renderPreview(contentHeight-2))
	metaPanel := m.renderPanel(rightW, contentHeight, m.renderMeta(contentHeight-2))

	spacer := lipgloss.NewStyle().Width(1).Background(tuiBgColor).Render(" ")
	body := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, spacer, previewPanel, spacer, metaPanel)

	statusText := m.status
	if m.mode == tuiModeCommand {
		statusText = m.commandInput.View()
	}
	if strings.TrimSpace(statusText) == "" {
		statusText = "j/k move  / find  : commands  e edit  a archive  q quit"
	}

	statusLine := lipgloss.NewStyle().
		Width(m.width).
		Foreground(tuiFgColor).
		Background(tuiStatusBg).
		Padding(0, 1).
		Render(statusText)

	return lipgloss.NewStyle().Background(tuiBgColor).Foreground(tuiFgColor).Render(body + "\n" + statusLine)
}

func (m tuiModel) renderEditView() string {
	title := m.editingTitle()
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(tuiAccentColor).
		Render(fmt.Sprintf("Editing: %s", title))

	footer := lipgloss.NewStyle().
		Foreground(tuiMutedColor).
		Render("Ctrl+S save    Esc cancel")

	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tuiAccentColor).
		Background(tuiBgColor).
		Foreground(tuiFgColor).
		Padding(0, 1).
		Width(maxInt(20, m.width-4)).
		Height(maxInt(6, m.height-3)).
		Render(header + "\n\n" + m.textarea.View() + "\n" + footer)

	statusLine := lipgloss.NewStyle().
		Width(m.width).
		Foreground(tuiFgColor).
		Background(tuiStatusBg).
		Padding(0, 1).
		Render(m.status)

	return lipgloss.NewStyle().Background(tuiBgColor).Foreground(tuiFgColor).Render(frame + "\n" + statusLine)
}

func (m tuiModel) panelWidths() (int, int, int) {
	width := maxInt(90, m.width)
	left := maxInt(28, width*28/100)
	right := maxInt(28, width*24/100)
	center := width - left - right - 2

	if center < 40 {
		deficit := 40 - center
		leftRoom := left - 28
		rightRoom := right - 28

		cutLeft := minInt(leftRoom, deficit/2+deficit%2)
		left -= cutLeft
		deficit -= cutLeft

		cutRight := minInt(rightRoom, deficit)
		right -= cutRight
		deficit -= cutRight

		if deficit > 0 && left > 24 {
			extra := minInt(left-24, deficit)
			left -= extra
			deficit -= extra
		}
		if deficit > 0 && right > 24 {
			extra := minInt(right-24, deficit)
			right -= extra
		}

		center = width - left - right - 2
	}

	return left, maxInt(24, center), right
}

func (m tuiModel) renderPanel(totalWidth, totalHeight int, content string) string {
	contentWidth := maxInt(1, totalWidth-4)
	contentHeight := maxInt(1, totalHeight-2)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tuiBorderColor).
		Background(tuiBgColor).
		Foreground(tuiFgColor).
		Padding(0, 1).
		Width(contentWidth).
		Height(contentHeight).
		Render(content)
}

func (m *tuiModel) resizeEditor() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	textWidth := maxInt(20, m.width-12)
	textHeight := maxInt(6, m.height-10)
	m.textarea.SetWidth(textWidth)
	m.textarea.SetHeight(textHeight)
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

func (m tuiModel) editingTitle() string {
	for _, item := range m.notes {
		if item.Note.ID == m.editingNoteID {
			return item.Note.Title
		}
	}
	return m.editingNoteID
}

func (m tuiModel) renderList(maxLines int) string {
	lines := []string{"Notes"}
	if m.activeQuery != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(tuiMutedColor).Render(fmt.Sprintf("filter: %q", m.activeQuery)))
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
		line := fmt.Sprintf("%s @%d %-8s %s", marker, idx+1, item.Note.Status, truncate(item.Note.Title, 38))
		if idx == m.selected {
			line = lipgloss.NewStyle().Bold(true).Foreground(tuiAccentColor).Render(line)
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
		"- e edit in TUI",
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

func updateBodyCmd(svc *core.Service, selector, body string) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.UpdateBody(selector, body)
		if err != nil {
			return opDoneMsg{err: err}
		}
		return opDoneMsg{status: fmt.Sprintf("saved %s -> %s", result.NoteID, result.RelPath)}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
