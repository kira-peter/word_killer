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
	onlineErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	onlineSelfStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
)

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

// RenderOnlineMenu renders the Create Room / Join Room menu
func RenderOnlineMenu(selected int, notice string, animFrame int) string {
	selectedStyle := lipgloss.NewStyle().Foreground(getRandomMenuColor(animFrame)).Bold(true)

	lines := []string{""}
	for i, opt := range []string{"Create Room", "Join Room", "Back"} {
		if i == selected {
			lines = append(lines, centerLine(selectedStyle.Render("> "+opt+" <")))
		} else {
			lines = append(lines, centerLine(menuNormalStyle.Render("  "+opt+"  ")))
		}
	}
	if notice != "" {
		lines = append(lines, "", centerLine(onlineErrorStyle.Render(notice)))
	}

	return renderOnlineScreen("Online", lines, "[↑↓] Select  │  [Enter] Confirm  │  [ESC] Back")
}

// RenderJoinForm renders the host IP and room number form
func RenderJoinForm(host, room string, focus int, notice string, connecting bool) string {
	field := func(label, value string, focused bool) string {
		text := fmt.Sprintf("%-14s", label)
		if focused {
			return centerLine(highlightStyle.Render("> ") + statItemStyle.Render(text) + inputStyle.Render(fmt.Sprintf("%-24s", value+"_")))
		}
		return centerLine("  " + statItemStyle.Render(text) + wordStyle.Render(fmt.Sprintf("%-24s", value)))
	}

	lines := []string{
		"",
		field("Host IP:", host, focus == 0),
		"",
		field("Room number:", room, focus == 1),
		"",
	}
	if connecting {
		lines = append(lines, centerLine(statsStyle.Render("Connecting...")))
	} else if notice != "" {
		lines = append(lines, centerLine(onlineErrorStyle.Render(notice)))
	}

	return renderOnlineScreen("Join Room", lines, "[Tab/↑↓] Switch  │  [Enter] Join  │  [ESC] Back")
}

// RenderLobby renders the room lobby
func RenderLobby(b *online.Battle, address, notice string, editingName bool, nameInput string) string {
	lines := []string{
		statItemStyle.Render("Room: ") + statValueStyle.Render(b.Room) +
			statItemStyle.Render("    Address: ") + statValueStyle.Render(address),
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

	return renderOnlineScreen("Room "+b.Room, lines, hints)
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
