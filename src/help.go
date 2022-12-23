package main

import "github.com/charmbracelet/bubbles/key"

var keysTableView = keyMapTableView{
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "New invalidation"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "Refresh"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "Back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

type keyMapTableView struct {
	New     key.Binding
	Refresh key.Binding
	Back    key.Binding
	Quit    key.Binding
}

func (k keyMapTableView) ShortHelp() []key.Binding {
	return []key.Binding{k.New, k.Refresh, k.Back, k.Quit}
}

func (k keyMapTableView) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.New, k.Refresh, k.Back},
		{k.Quit},
	}
}
