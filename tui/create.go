package tui

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	humancron "github.com/lnquy/cron"
	"github.com/robfig/cron/v3"
	"golang.org/x/term"
)

type state int

const (
	stateAskName state = iota
	stateAskOwner
	stateAskTrigger
	stateAskStartDate
	stateAskCron
	stateAskWindow
	stateAskTask
	stateTaskConfig
	stateDone
	stateQuit
)

const (
	startDatePlace  = "Specify the start date of schedule?"
	cronPlaceholder = "Specify the cron schedule, eg. '0 2 * * *' for 2AM every day."

	dateFormat = "2006-01-02"

	triggerManual    = "Manual"
	triggerScheduled = "Scheduled"
)

// NewCreateModel renders the UI for creating a new job
func NewCreateModel() (*createModel, error) {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	f := &createModel{
		width:     width,
		height:    height,
		state:     stateAskName,
		textinput: textinput.New(),
		questions: 1,
	}

	lineHeight := height - 25

	f.triggerList = list.New([]list.Item{
		listItem{
			name:        "Scheduled",
			description: "Job runs automatically on a schedule",
		},
		listItem{
			name:        "Manual",
			description: "Job should be triggered manually",
		},
	}, list.NewDefaultDelegate(), width, lineHeight)

	f.textinput.Focus()
	f.textinput.CharLimit = 256
	f.textinput.Width = width
	f.textinput.Prompt = "→  "

	hideListChrome(&f.triggerList)

	return f, nil
}

// createModel represents the survey state when creating a new function.
type createModel struct {
	width  int
	height int

	state     state
	questions int

	name  string
	owner string

	// triggerType is the type of trigger. cron or manual.
	triggerType string

	// cron expression and the next invocation
	startDate    time.Time
	startDateErr error

	cron      string
	humanCron string
	cronError error
	nextCron  time.Time

	window string

	textinput   textinput.Model
	triggerList list.Model
}

// Ensure that createModel fulfils the tea.Model interface.
var _ tea.Model = (*createModel)(nil)

func (c createModel) Init() tea.Cmd {
	return tea.Batch()
}

func hideListChrome(lists ...*list.Model) {
	for _, l := range lists {
		l.SetShowFilter(false)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetShowTitle(false)
	}
}

func (c *createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if c.state == stateDone {
		return c, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlBackslash:
			c.state = stateQuit
			return c, tea.Quit
		}

		if msg.String() == "q" {
			c.state = stateDone
			return c, tea.Quit
		}
	}

	originalState := c.state
	_, cmd = func() (tea.Model, tea.Cmd) {
		switch c.state {
		case stateAskName:
			return c.updateName(msg)
		case stateAskOwner:
			return c.updateOwner(msg)
		case stateAskTrigger:
			return c.updateTrigger(msg)
		case stateAskStartDate:
			return c.updateStartDate(msg)
		case stateAskCron:
			return c.updateCron(msg)
		}
		return c, nil
	}()
	if c.state != originalState {
		c.questions++
	}

	// Merge the async commands from each state into the top-level commands to run.
	cmds = append(cmds, cmd)

	// Return our updated state and all commands to run.
	return c, tea.Batch(cmds...)
}

func (c *createModel) updateName(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.textinput.Placeholder = "Name of the job?"
	c.name = c.textinput.Value()
	c.textinput, cmd = c.textinput.Update(msg)

	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter && c.name != "" {
		c.textinput.Placeholder = ""
		c.textinput.SetValue("")
		c.state = stateAskOwner
	}

	return c, cmd
}

func (c *createModel) updateOwner(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.textinput.Placeholder = "Owner of the job?"
	c.owner = c.textinput.Value()
	c.textinput, cmd = c.textinput.Update(msg)

	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter && c.owner != "" {
		c.textinput.Placeholder = cronPlaceholder
		c.textinput.SetValue("")
		c.state = stateAskTrigger
	}

	return c, cmd
}

func (c *createModel) updateTrigger(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// We press enter to select an item
	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter {
		c.triggerType = c.triggerList.SelectedItem().FilterValue()

		switch c.triggerType {
		case triggerManual:
			c.textinput.Placeholder = cronPlaceholder
			c.state = stateDone
			c.textinput.SetValue("")
		case triggerScheduled:
			c.textinput.Placeholder = startDatePlace
			c.state = stateAskStartDate
			today := time.Now().Format(dateFormat)
			c.textinput.SetValue(today)
		}
		return c, nil
	}

	c.triggerList, cmd = c.triggerList.Update(msg)
	cmds = append(cmds, cmd)
	return c, tea.Batch(cmds...)
}

func (c *createModel) updateStartDate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.textinput.Placeholder = startDatePlace
	startString := c.textinput.Value()
	c.textinput, cmd = c.textinput.Update(msg)

	var err error
	c.startDate, err = time.Parse(dateFormat, startString)
	if err != nil {
		c.startDateErr = fmt.Errorf("Invalid date: %s", c.startDate)
	} else if c.startDate.Year() < 2000 {
		c.startDateErr = fmt.Errorf("Date before 2000 are not allowed")
	} else {
		c.startDateErr = nil
	}

	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter && c.startDateErr == nil {
		c.textinput.Placeholder = cronPlaceholder
		c.textinput.SetValue("")
		c.state = stateAskCron
	}

	return c, cmd
}

func (c *createModel) updateCron(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.textinput.Placeholder = cronPlaceholder
	c.cron = c.textinput.Value()
	c.textinput, cmd = c.textinput.Update(msg)

	schedule, err := cron.ParseStandard(c.cron)
	if err != nil {
		c.cronError = fmt.Errorf("Cron expression is not valid")
		c.humanCron = ""
		c.nextCron = time.Time{}
	} else {
		c.cronError = nil
		if desc, err := humancron.NewDescriptor(); err == nil {
			c.humanCron, _ = desc.ToDescription(c.cron, humancron.Locale_en)
		}
		start := time.Now()
		if start.Before(c.startDate) {
			start = c.startDate
		}
		c.nextCron = schedule.Next(start)
	}

	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter && c.cron != "" && c.cronError == nil {
		c.textinput.Placeholder = cronPlaceholder
		c.textinput.SetValue("")
		c.state = stateDone
	}

	return c, cmd
}

func (c *createModel) View() string {
	b := &strings.Builder{}

	if c.height > 35 {
		b.WriteString(c.renderIntro())
	}

	switch c.state {
	case stateAskName:
		b.WriteString(c.renderName())
	case stateAskOwner:
		b.WriteString(c.renderOwner())
	case stateAskTrigger:
		b.WriteString(c.renderTrigger())
	case stateAskStartDate:
		b.WriteString(c.renderStartDate())
	case stateAskCron:
		b.WriteString(c.renderCron())
	case stateDone:
		b.WriteString("\n")
	}

	return b.String()
}

func (c *createModel) renderIntro() string {
	b := &strings.Builder{}

	b.WriteString("\n")
	b.WriteString(BoldStyle.Render("Let's get you set up with optimus job."))
	b.WriteString("\n")
	b.WriteString(TextStyle.Copy().Foreground(Feint).Render("Answer these questions to get started."))
	b.WriteString("\n\n")
	b.WriteString(c.renderState())
	return b.String()
}

func (c *createModel) renderState() string {
	if c.state == stateAskName {
		return ""
	}

	b := &strings.Builder{}
	n := 1
	write := func(s string) {
		b.WriteString(fmt.Sprintf("%d. %s", n, s))
		n++
	}

	write("Job name: " + BoldStyle.Render(c.name) + "\n")

	if c.owner != "" && c.state != stateAskOwner {
		write("Owner: " + BoldStyle.Render(c.owner) + "\n")
	}

	if c.triggerType != "" {
		write("Job trigger: " + BoldStyle.Render(c.triggerType) + "\n")
	}
	if c.startDate.Year() > 2000 && c.state != stateAskStartDate {
		write("Start Date: " + BoldStyle.Render(c.startDate.Format(dateFormat)) + "\n")
	}

	if c.cron != "" && c.state != stateAskCron {
		write("Cron schedule: " + BoldStyle.Render(c.cron) + " (" + c.humanCron + ")\n")
	}

	return b.String()
}

func (c *createModel) renderName() string {
	b := &strings.Builder{}
	b.WriteString(BoldStyle.Render(fmt.Sprintf("%d. Job name:", c.questions)) + "\n")
	b.WriteString(c.textinput.View())
	return b.String()
}

func (c *createModel) renderOwner() string {
	b := &strings.Builder{}
	b.WriteString(BoldStyle.Render(fmt.Sprintf("%d. Owner:", c.questions)) + "\n")
	b.WriteString(c.textinput.View())
	return b.String()
}

func (c *createModel) renderTrigger() string {
	b := &strings.Builder{}
	b.WriteString(BoldStyle.Render(fmt.Sprintf("%d. How should the job trigger?", c.questions)) + "\n\n")
	b.WriteString(c.triggerList.View())
	return b.String()
}

func (c *createModel) renderStartDate() string {
	b := &strings.Builder{}
	b.WriteString(BoldStyle.Render(fmt.Sprintf("%d. Start Date:", c.questions)) + "\n")
	b.WriteString(c.textinput.View())
	if c.startDateErr != nil {
		b.WriteString("\n")
		b.WriteString(RenderWarning(c.startDateErr.Error()))
	}

	return b.String()
}

func (c *createModel) renderCron() string {
	b := &strings.Builder{}
	b.WriteString(BoldStyle.Render(fmt.Sprintf("%d. Cron schedule:", c.questions)) + "\n")
	b.WriteString(c.textinput.View())
	if c.cronError != nil {
		b.WriteString("\n")
		b.WriteString(RenderWarning(c.cronError.Error()))
	}
	if !c.nextCron.IsZero() {
		b.WriteString("\n")
		dur := humanizeDuration(time.Until(c.nextCron))

		if c.humanCron != "" {
			b.WriteString(TextStyle.Copy().Foreground(Feint).Bold(true).Render(c.humanCron) + ". ")
		}
		b.WriteString(TextStyle.Copy().Foreground(Feint).Render("This would next run at: " + c.nextCron.Format(time.RFC3339) + " (in " + dur + ")\n"))
	}
	return b.String()
}

func humanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"day", days},
		{"hour", hours},
		{"minute", minutes},
		{"second", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		case 1:
			parts = append(parts, fmt.Sprintf("%d %s", chunk.amount, chunk.singularName))
		default:
			parts = append(parts, fmt.Sprintf("%d %ss", chunk.amount, chunk.singularName))
		}
	}

	return strings.Join(parts, " ")
}

type listItem struct {
	name        string
	description string
}

func (i listItem) Title() string       { return i.name }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.name }
