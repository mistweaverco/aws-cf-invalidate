package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
)

var DOCSTYLE = lipgloss.NewStyle().Margin(1, 2)
var CLIENT = &cloudfront.Client{}
var CF_DIST_ID = ""
var VIEW = "list"

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.desc }

type model struct {
	list          list.Model
	table         table.Model
	textInput     textinput.Model
	help          help.Model
	keysTableView keyMapTableView
}

func (m model) Init() tea.Cmd {
	return nil
}

func createInvalidation(m *model) {
	ti := textinput.New()
	ti.Placeholder = "/*"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	m.textInput = ti
}

func viewListInvalidations(m *model) {
	resp, err := CLIENT.ListInvalidations(context.TODO(), &cloudfront.ListInvalidationsInput{
		DistributionId: aws.String(CF_DIST_ID),
	})

	if err != nil {
		fmt.Printf("list.inval. err:%v", err.Error())
	}

	columns := []table.Column{
		{Title: "ID", Width: 18},
		{Title: "Datetime", Width: 22},
		{Title: "Status", Width: 14},
	}

	rows := []table.Row{}

	for _, invalidation := range resp.InvalidationList.Items {
		timeStr := invalidation.CreateTime.Local().Format("2006-01-02 15:04:05")
		rows = append(rows, table.Row{
			*invalidation.Id,
			timeStr,
			*invalidation.Status,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	m.table = t
}

func createInvalidationRequest(pattern string) error {
	unixTime := time.Now().Unix()
	unixTimeStr := strconv.FormatInt(unixTime, 10)
	invalidationBatch := types.InvalidationBatch{
		CallerReference: aws.String("aws-cf-invalidation-" + unixTimeStr),
		Paths: &types.Paths{
			Quantity: aws.Int32(1),
			Items:    []string{pattern},
		},
	}
	input := cloudfront.CreateInvalidationInput{}
	input.DistributionId = &CF_DIST_ID
	input.InvalidationBatch = &invalidationBatch
	_, err := CLIENT.CreateInvalidation(context.TODO(), &input)
	return err
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "esc" && VIEW == "textinput" {
			viewListInvalidations(&m)
			VIEW = "table"
			return m, cmd
		}
		if msg.String() == "backspace" && VIEW == "table" {
			VIEW = "list"
			return m, nil
		}
		if msg.String() == "enter" && VIEW == "list" {
			CF_DIST_ID = m.list.SelectedItem().(item).title
			viewListInvalidations(&m)
			VIEW = "table"
			return m, nil
		}
		if msg.String() == "r" && VIEW == "table" {
			viewListInvalidations(&m)
			VIEW = "table"
			return m, nil
		}
		if msg.String() == "o" && VIEW == "table" {
			invalidationId := m.table.SelectedRow()[0]
			browser.OpenURL("https://us-east-1.console.aws.amazon.com/cloudfront/v3/home#/distributions/" + CF_DIST_ID + "/invalidations/details/" + invalidationId)
		}
		if msg.String() == "enter" && VIEW == "textinput" {
			createInvalidationRequest(m.textInput.Value())
			viewListInvalidations(&m)
			VIEW = "table"
			return m, nil
		}
		if msg.String() == "n" && VIEW == "table" {
			createInvalidation(&m)
			VIEW = "textinput"
			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := DOCSTYLE.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.help.Width = msg.Width
	}

	switch VIEW {
	case "list":
		m.list, cmd = m.list.Update(msg)
	case "table":
		m.table, cmd = m.table.Update(msg)
	case "textinput":
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	switch VIEW {
	case "list":
		return DOCSTYLE.Render(m.list.View())
	case "table":
		helpView := "\n\n" + m.help.View(m.keysTableView)
		return DOCSTYLE.Render(m.table.View() + helpView)
	case "textinput":
		return fmt.Sprintf(
			"Path to invalidate?\n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to quit)",
		) + "\n"
	}
	return ""
}

func getDistClient() *cloudfront.Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return cloudfront.NewFromConfig(cfg)
}

func main() {
	browser.Stdout = ioutil.Discard
	browser.Stderr = ioutil.Discard
	CLIENT = getDistClient()
	dists, err := CLIENT.ListDistributions(context.TODO(), &cloudfront.ListDistributionsInput{})
	if err != nil {
		log.Fatalf("unable to list distributions, %v", err)
	}
	items := []list.Item{}
	for _, dist := range dists.DistributionList.Items {
		for _, alias := range dist.Aliases.Items {
			items = append(items, item{title: *dist.Id, desc: alias})
		}
	}

	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Datetime", Width: 22},
		{Title: "Status", Width: 10},
	}

	rows := []table.Row{
		{"2", "2022-12-21 18:11:12", "Pending"},
		{"1", "2022-12-21 17:16:03", "Completed"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		list:          list.New(items, list.NewDefaultDelegate(), 0, 0),
		table:         t,
		keysTableView: keysTableView,
		help:          help.New(),
	}

	m.list.Title = "Distribution List"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
