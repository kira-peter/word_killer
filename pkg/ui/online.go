package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/word-killer/word-killer/pkg/online"
)

const (
	onlineContentHeight = 12
	dropAreaRows        = 9
	onlineInnerWidth    = contentWidth - 4
)

var (
	onlineErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	onlineSelfStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	onlineDisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// maxListedRooms keeps the menu at the standard height; the list scrolls to
// keep the selection visible, with one line each reserved for the scroll hint
// and the notice
const maxListedRooms = 6

// renderOnlineScreen lays out the standard header, content box and hints
func renderOnlineScreen(header string, lines []string, hints string) string {
	var s strings.Builder

	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(headerStyle.Render(header)))
	s.WriteString("\n")

	for len(lines) < onlineContentHeight {
		lines = append(lines, "")
	}
	s.WriteString(wordBoxStyle.Render(strings.Join(lines, "\n")))
	s.WriteString("\n")

	s.WriteString(inputBoxStyle.Render(hints))
	s.WriteString("\n")
	return s.String()
}

func centerLine(text string) string {
	return lipgloss.NewStyle().Width(onlineInnerWidth).Align(lipgloss.Center).Render(text)
}

// roomStatusLabel describes whether a room can be joined
func roomStatusLabel(r online.RoomInfo) string {
	switch {
	case r.InGame:
		return "In game"
	case r.Full():
		return "Full"
	}
	return "Waiting"
}

// RenderOnlineMenu renders Create Room followed by the rooms found on the LAN.
// selected is 0 for Create Room, or 1+i for rooms[i].
func RenderOnlineMenu(rooms []online.RoomInfo, selected int, notice string, connecting bool, animFrame int) string {
	selectedStyle := lipgloss.NewStyle().Foreground(getRandomMenuColor(animFrame)).Bold(true)

	lines := []string{""}
	if selected == 0 {
		lines = append(lines, centerLine(selectedStyle.Render("> Create Room <")))
	} else {
		lines = append(lines, centerLine(menuNormalStyle.Render("  Create Room  ")))
	}

	lines = append(lines, "", centerLine(separatorStyle.Render("──────────── ")+
		titleStyle.Render("Rooms on your network")+separatorStyle.Render(" ────────────")))

	if len(rooms) == 0 {
		lines = append(lines, centerLine(statsStyle.Render("Searching for rooms...")))
	}

	// Scroll so the selected room stays inside the visible window
	first := 0
	if selected > maxListedRooms {
		first = selected - maxListedRooms
	}
	last := min(first+maxListedRooms, len(rooms))
	for i := first; i < last; i++ {
		r := rooms[i]
		text := fmt.Sprintf("%-12s %d/%d  %-8s %-15s", r.HostName, r.Players, r.MaxPlayers, roomStatusLabel(r), r.IP)
		switch {
		case selected == i+1:
			lines = append(lines, centerLine(selectedStyle.Render("> "+text+" <")))
		case !r.Joinable():
			lines = append(lines, centerLine(onlineDisabledStyle.Render("  "+text+"  ")))
		default:
			lines = append(lines, centerLine(menuNormalStyle.Render("  "+text+"  ")))
		}
	}
	if hidden := len(rooms) - (last - first); hidden > 0 {
		lines = append(lines, centerLine(statsStyle.Render(fmt.Sprintf("%d more room(s), use ↑↓ to scroll", hidden))))
	}

	switch {
	case connecting:
		lines = append(lines, centerLine(statsStyle.Render("Connecting...")))
	case notice != "":
		lines = append(lines, centerLine(onlineErrorStyle.Render(notice)))
	}

	return renderOnlineScreen("Online", lines, "[↑↓] Select  │  [Enter] Confirm  │  [ESC] Back")
}

// RenderLobby renders the room lobby
func RenderLobby(b *online.Battle, address, notice string, editingName bool, nameInput string) string {
	lines := []string{
		statItemStyle.Render("Host: ") + statValueStyle.Render(address),
		statItemStyle.Render("Mode: ") + statValueStyle.Render(online.ModeName(b.Mode)),
		"",
		titleStyle.Render(fmt.Sprintf("Players (%d/%d):", len(b.Players), online.MaxPlayers)),
	}
	for _, p := range b.Players {
		line := "  " + wordStyle.Render(p.Name)
		if p.ID == b.MyID {
			name := p.Name
			if editingName {
				name = nameInput + "_"
			}
			line = "  " + onlineSelfStyle.Render(name)
		}
		if p.IsHost {
			line += statsStyle.Render(" (host)")
		}
		lines = append(lines, line)
	}

	var hints string
	if b.IsHost() {
		hints = "[↑↓] Mode  │  [Tab] Rename  │  [Enter] Start  │  [ESC] Close Room"
		// The only lobby notice is about player count, so it clears once enough have joined
		if notice != "" && len(b.Players) < online.MinPlayers {
			lines = append(lines, "", onlineErrorStyle.Render(notice))
		}
	} else {
		lines = append(lines, "", statsStyle.Render("Waiting for host..."))
		hints = "[Tab] Rename  │  [ESC] Leave"
	}
	if editingName {
		hints = "Type your name  │  [Enter] Save  │  [ESC] Cancel"
	}

	return renderOnlineScreen("Room", lines, hints)
}

// renderScoreBar renders every player's name and score on one line
func renderScoreBar(b *online.Battle) string {
	parts := make([]string, 0, len(b.Players))
	for _, p := range b.Players {
		nameStyle := statItemStyle
		if p.ID == b.MyID {
			nameStyle = onlineSelfStyle
		}
		parts = append(parts, nameStyle.Render(p.Name)+" "+statValueStyle.Render(fmt.Sprintf("%d", b.Scores[p.ID])))
	}
	return strings.Join(parts, separatorStyle.Render(" │ "))
}

// renderFallingWord highlights the part of the word the input already matches
func renderFallingWord(word, input string) string {
	if input != "" && strings.HasPrefix(word, input) {
		return highlightStyle.Render(input) + wordStyle.Render(word[len(input):])
	}
	return wordStyle.Render(word)
}

// RenderBattle renders the word drop battle: score bar, drop area and input box
func RenderBattle(b *online.Battle, now time.Time) string {
	lines := []string{centerLine(renderScoreBar(b)), ""}

	wordRow := -1
	message := ""
	switch b.Phase {
	case online.BattleCountdown:
		message = hitEffectStyle.Render(fmt.Sprintf("Get ready: %d", b.CountdownSeconds(now)))
	case online.BattleBetween:
		if b.FeedbackHit {
			message = highlightStyle.Render(b.Feedback)
		} else {
			message = onlineErrorStyle.Render(b.Feedback)
		}
	case online.BattleFalling:
		wordRow = int(b.FallProgress(now) * float64(dropAreaRows-1))
	}

	for row := 0; row < dropAreaRows; row++ {
		switch {
		case row == wordRow:
			lines = append(lines, centerLine(renderFallingWord(b.Word, b.Input)))
		case message != "" && row == dropAreaRows/2:
			lines = append(lines, centerLine(message))
		default:
			lines = append(lines, "")
		}
	}
	lines = append(lines, separatorStyle.Render(strings.Repeat("━", onlineInnerWidth)))

	inputText := inputStyle.Render(b.Input)
	if b.Input != "" && !strings.HasPrefix(b.Word, b.Input) {
		inputText = onlineErrorStyle.Render(b.Input)
	}

	header := fmt.Sprintf("%s  │  Word %d/%d", online.ModeName(b.Mode), b.WordIndex+1, b.TotalWords)
	return renderOnlineScreen(header, lines, "Input: "+inputText+"_")
}

// RenderBattleResult renders the final ranking
func RenderBattleResult(b *online.Battle) string {
	lines := []string{
		centerLine(titleStyle.Render("BATTLE RESULT")),
		"",
		centerLine(statItemStyle.Render(fmt.Sprintf("%-4s %-14s %6s %6s", "#", "Player", "Score", "Words")) + "  "),
		centerLine(separatorStyle.Render(strings.Repeat("━", 33)) + "  "),
	}

	for i, r := range b.Ranking {
		style := wordStyle
		if r.ID == b.MyID {
			style = onlineSelfStyle
		}
		line := style.Render(fmt.Sprintf("%-4d %-14s %6d %6d", i+1, r.Name, r.Score, r.Words))
		if b.IsWinner(r.ID) {
			line += highlightStyle.Render(" ★")
		} else {
			line += "  "
		}
		lines = append(lines, centerLine(line))
	}

	lines = append(lines, "")
	switch {
	case b.IsWinner(b.MyID) && len(b.Winners) > 1:
		lines = append(lines, centerLine(highlightStyle.Render("It's a tie!")))
	case b.IsWinner(b.MyID):
		lines = append(lines, centerLine(highlightStyle.Render("You win!")))
	default:
		lines = append(lines, centerLine(statsStyle.Render("Better luck next time!")))
	}

	hints := "Waiting for host...  │  [ESC] Leave"
	if b.IsHost() {
		hints = "[Enter] Back to Lobby  │  [ESC] Close Room"
	}
	return renderOnlineScreen("Battle Result", lines, hints)
}
