package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Constants

var stations = []Station{
	ChristianRock,
	ChristianHits,
	ChristianLofi,
	GospelMix,
	Melodia,
}

// Model

type model struct {
	cursor   int
	selected int
	song     modelSong
	player   *Player
}

type modelSong struct {
	isLoading bool
	error     error
	data      Song
}

func initialModel() model {
	selected := 0
	player, err := NewPlayer(stations[selected].stream)
	if err != nil {
		fmt.Printf("Error creating player: %v", err)
		os.Exit(1)
	}
	return model{
		cursor:   0,
		selected: selected,
		song: modelSong{
			isLoading: true,
			error:     nil,
			data:      Song{},
		},
		player: player,
	}
}

func (m model) Init() tea.Cmd {
	return getSong(stations[m.cursor])
}

// Msg

type gotSongMsg Song

type errMsg struct{ err error }

func (e errMsg) Error() string {
	return e.err.Error()
}

// Update

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case gotSongMsg:
		m.song = modelSong{
			isLoading: false,
			error:     nil,
			data:      Song(msg),
		}
	case errMsg:
		m.song = modelSong{
			isLoading: false,
			error:     msg,
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.selected = m.cursor
			m.song = modelSong{
				isLoading: true,
			}
			m.player.Play(stations[m.selected].stream)
			cmd = getSong(stations[m.cursor])
		case "j":
			if m.cursor == len(stations)-1 {
				m.cursor = 0
			} else {
				m.cursor++
			}
		case "G":
			m.cursor = len(stations) - 1
		case "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(stations) - 1
			}
		case "g":
			m.cursor = 0
		case "q", "Q":
			m.player.Quit()
			return m, tea.Quit
		case "r":
			m.song = modelSong{
				isLoading: true,
			}
			cmd = getSong(stations[m.selected])
		case " ":
			if m.player.IsPlaying {
				m.player.Stop()
			} else {
				m.player.Resume()
			}
		}
	}

	return m, cmd
}

// View

func (m model) View() string {
	sectionStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

	stationsSection := viewStations(m)
	songSection := viewSongSection(
		m,
		lipgloss.Height(stationsSection),
		lipgloss.Width(stationsSection),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		sectionStyle.Render(stationsSection),
		sectionStyle.Render(songSection),
	)

}

func viewStations(m model) string {
	var stationsSection string
	for index, station := range stations {
		cursor := " "
		if m.cursor == index {
			cursor = ">"
		}

		newLine := "\n"
		if index == len(stations)-1 {
			newLine = ""
		}

		isSelected := m.selected == index
		stationLabel := lipgloss.NewStyle().
			Bold(isSelected).
			Underline(isSelected).
			Render(station.Name)

		stationsSection += fmt.Sprintf("%s %s %s", cursor, stationLabel, newLine)
	}

	return stationsSection
}

func viewSongSection(m model, height int, width int) string {
	var title string
	var artist string
	if m.song.isLoading {
		title = "Loading"
	} else if m.song.error != nil {
		title = "Error"
		artist = m.song.error.Error()
	} else if m.song.data.Title == "" {
		title = "Unknown"
		artist = "No song data"
	} else {
		title = m.song.data.Title
		artist = m.song.data.Artist
	}

	icon := "󰐊"
	if m.player.IsPlaying {
		icon = "󰏤"
	}
	content := fmt.Sprintf(
		"%s\n%s\n%s",
		icon,
		lipgloss.NewStyle().Bold(true).Render(title),
		artist,
	)

	actualWidth := width
	if contentWidth := lipgloss.Width(content) + 2; contentWidth > actualWidth {
		actualWidth = contentWidth
	}

	return lipgloss.NewStyle().
		Height(height).
		Width(actualWidth).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// Helpers

func getSong(station Station) tea.Cmd {
	return func() tea.Msg {
		song, err := station.GetSong()
		if err != nil {
			return errMsg{err}
		}

		return gotSongMsg(song)
	}
}

// Main

func main() {
	program := tea.NewProgram(initialModel())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
