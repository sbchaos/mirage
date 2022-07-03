package tui

import (
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/sbchaos/mirage/job"
)

var (
	listWidth = 50
)

func NewWindow() (*DataWindow, error) {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))
	return NewDataWindow(width, height, time.Now())
}

func NewDataWindow(width, height int, ref time.Time) (*DataWindow, error) {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(1)

	l := list.New([]list.Item{
		listItem{
			name:        "Hourly",
			description: "Job runs on hourly data",
		},
		listItem{
			name:        "Daily",
			description: "Job runs on daily data",
		},
		listItem{
			name:        "Weekly",
			description: "Job runs on weekly data",
		},
		listItem{
			name:        "Monthly",
			description: "Job runs on monthly data",
		},
	}, delegate, listWidth, height-25)
	l.Title = "Select data size"
	l.SetFilteringEnabled(false)
	//l.DisableQuitKeybindings()
	l.KeyMap = listKeyMap() // Remove the J/K keyboard navigation.

	return &DataWindow{
		width:         width,
		height:        height,
		referenceTime: ref,
		listTruncate:  l,
	}, nil
}

type DataWindow struct {
	width  int
	height int

	referenceTime time.Time

	size       time.Duration
	offset     time.Duration
	truncateTo string

	listTruncate list.Model
}

var _ tea.Model = (*DataWindow)(nil)

func (e *DataWindow) Init() tea.Cmd {
	return nil
}

// UpdateSize updates the size of the event browser's rendering area.
func (e *DataWindow) UpdateSize(width, height int) {
	if width < 100 {
		listWidth = 24
	} else {
		listWidth = 50
	}

	e.width = width
	e.height = height
	e.listTruncate.SetHeight(height)
	e.listTruncate.SetWidth(listWidth)
}

// Update handles incoming keypresses, mouse moves, resize events etc.
func (e *DataWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlBackslash:
			return e, tea.Quit
		}

		if msg.String() == "q" {
			return e, tea.Quit
		}
	}

	e.listTruncate, cmd = e.listTruncate.Update(msg)
	cmds = append(cmds, cmd)

	return e, tea.Batch(cmds...)
}

func (e DataWindow) Selected() *job.DataWindow {
	selectedItem := e.listTruncate.SelectedItem()
	if selectedItem == nil {
		return nil
	}

	var truncate string
	var size time.Duration
	switch selectedItem.FilterValue() {
	case "Hourly":
		truncate = "h"
		size = time.Hour * 1
	case "Daily":
		truncate = "d"
		size = time.Hour * 24
	case "Weekly":
		truncate = "w"
		size = time.Hour * 7 * 24
	case "Monthly":
		truncate = "M"
		size = time.Hour * 7 * 30
	}

	return &job.DataWindow{
		Size:       size,
		Offset:     0,
		TruncateTo: truncate,
	}
}

// View renders the list.
func (e *DataWindow) View() string {
	b := &strings.Builder{}
	b.WriteString(e.renderHeader())

	selected := e.Selected()
	start, end := selected.GetNextInterval(e.referenceTime)

	list := e.renderList()
	detail := e.renderDetail(start, end)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, list, detail))

	return b.String()
}

func (e *DataWindow) renderList() string {
	left := lipgloss.NewStyle().
		Width(listWidth+2). // plus padding
		Padding(2, 2, 2, 0).
		Render(e.listTruncate.View())
	return left
}

func (e *DataWindow) renderDetail(start, end time.Time) string {
	duration := humanizeDuration(end.Sub(start))

	desc := lipgloss.NewStyle().PaddingLeft(1).
		Foreground(Feint).Render("Duration: " + duration)

	content := lipgloss.JoinVertical(lipgloss.Center,
		"\n",
		"\n",
		desc,
		"\n",
		renderWithStartEnd(start, end),
		renderDataFrame("text"))

	return content
}

func (e *DataWindow) renderHeader() string {
	// Render two columns of text.
	headerMsg := lipgloss.JoinVertical(lipgloss.Center,
		BoldStyle.Copy().Foreground(Feint).Render("Data Window"),
		TextStyle.Copy().Foreground(Feint).Render("Showing representation of data for the window ")+
			BoldStyle.Copy().Foreground(Feint).Render("Tab: select. Ctrl-j/k: navigate code"),
	)

	return lipgloss.Place(e.width, 3, lipgloss.Center, lipgloss.Center, headerMsg)
}

func renderDataFrame(name string) string {
	highlight := lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	tabBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab := lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 3).
		Width(20)

	tabGap := tab.Copy().
		Width(15).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)
	gap := tabGap.Render(" ")

	return lipgloss.JoinHorizontal(lipgloss.Bottom, gap, tab.Render(name), gap)
}

func renderWithStartEnd(start, end time.Time) string {
	empty := lipgloss.NewStyle().Width(14)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		renderDate(start),
		empty.Render(""),
		renderDate(end),
	)
}

func renderDate(date time.Time) string {
	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#A550DF")).
		Width(12).
		Padding(0, 1)

	return lipgloss.JoinVertical(lipgloss.Center,
		status.Render(date.Format(dateFormat)),
		status.Render(date.Format(time.Kitchen)),
	)
}

func listKeyMap() list.KeyMap {
	return list.KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "next page"),
		),
	}
}
