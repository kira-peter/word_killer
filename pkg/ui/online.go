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
	onlineInnerWidth    = contentWidth - 4

	// liveRows is one row per player showing what they are typing
	liveRows = 4
	// dropAreaRows leaves room under the falling word for the live rows
	dropAreaRows = 5
)

var (
	onlineErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	onlineSelfStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	onlineDisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// playerColors identify players across the score bar, board and results.
	// None is red or green, which mark wrong and right letters.
	playerColors = []lipgloss.Color{"75", "221", "177", "87"} // blue, yellow, purple, cyan

	// Wheel steps fade the further they are from the current value
	wheelNearStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	wheelFarStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
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

// playerStyle is the bold color a player is shown in
func playerStyle(b *online.Battle, id int) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(playerColors[b.ColorSlot(id)%len(playerColors)]).Bold(true)
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

// searchingFrames animates the dots after "Searching for rooms": one, two,
// three, then none, so the line visibly breathes while nothing is found
var searchingFrames = []string{".  ", ".. ", "...", "   "}

// searchingFrameTicks is how many 33ms ticks each frame stays up
const searchingFrameTicks = 12

// menuListRows is the height of the room list area, which keeps the notice
// on the bottom line whatever the list holds
const menuListRows = maxListedRooms + 1

// RenderOnlineMenu renders Create Room followed by the rooms found on the LAN.
// selected is 0 for Create Room, or 1+i for rooms[i]. The notice, if any,
// sits on the bottom line of the content box.
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

	var list []string
	if len(rooms) == 0 {
		// Centered in the list area, with dots that count up and clear
		dots := searchingFrames[(animFrame/searchingFrameTicks)%len(searchingFrames)]
		for len(list) < menuListRows/2 {
			list = append(list, "")
		}
		list = append(list, centerLine(statsStyle.Render("Searching for rooms"+dots)))
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
			list = append(list, centerLine(selectedStyle.Render("> "+text+" <")))
		case !r.Joinable():
			list = append(list, centerLine(onlineDisabledStyle.Render("  "+text+"  ")))
		default:
			list = append(list, centerLine(menuNormalStyle.Render("  "+text+"  ")))
		}
	}
	if hidden := len(rooms) - (last - first); hidden > 0 {
		list = append(list, centerLine(statsStyle.Render(fmt.Sprintf("%d more room(s), use ↑↓ to scroll", hidden))))
	}
	for len(list) < menuListRows {
		list = append(list, "")
	}
	lines = append(lines, list...)

	// The bottom line is reserved for the status or a notice
	for len(lines) < onlineContentHeight-1 {
		lines = append(lines, "")
	}
	switch {
	case connecting:
		lines = append(lines, centerLine(statsStyle.Render("Connecting...")))
	case notice != "":
		lines = append(lines, centerLine(onlineErrorStyle.Render(notice)))
	}

	return renderOnlineScreen("Online", lines, "[↑↓] Select  │  [Enter] Confirm  │  [ESC] Back")
}

// LobbyItem is one row of the lobby menu
type LobbyItem struct {
	Label string
	Value string // empty for Start Game
}

// LobbyItems lists the lobby menu: the mode, the mode's settings, then Start Game
func LobbyItems(b *online.Battle) []LobbyItem {
	items := []LobbyItem{{Label: "Mode", Value: online.ModeName(b.Mode)}}
	for _, o := range online.ModeOptions(b.Mode) {
		items = append(items, LobbyItem{Label: o.Label, Value: o.Format(b.Settings.Get(o.Key))})
	}
	return append(items, LobbyItem{Label: "Start Game"})
}

// LobbyView is what the lobby is currently showing on top of the room
type LobbyView struct {
	Selected    int    // highlighted menu row, host only
	Notice      string // shown under the settings
	EditingName bool   // the local player is renaming
	NameInput   string //
	Wheel       *Wheel // the selected setting is being adjusted in place
	KickPicker  []int  // player ids offered for kicking, nil when closed
	KickIndex   int    //
	Address     string // host:port
}

// Wheel is the setting being adjusted: its steps, with Index the step
// currently centered
type Wheel struct {
	Label string
	Steps []string
	Index int
}

// Lobby layout: settings on the left half, one framed card per player on the
// right. Four cards of three lines fill the content box exactly.
const (
	lobbyHalfWidth   = onlineInnerWidth / 2
	lobbyLabelWidth  = 12
	lobbyMenuTop     = 2 // blank rows above the first item, room for the wheel's upper steps
	lobbyCardWidth   = lobbyHalfWidth - 2
	lobbyWheelReach  = 2 // steps shown above and below the centered one
	lobbyValueColumn = 2 + lobbyLabelWidth + 2
)

// RenderLobby renders the room lobby. Adjusting a setting keeps the same
// layout: every other setting is hidden, the setting's name stays where it
// was, and a wheel opens in place of its value.
func RenderLobby(b *online.Battle, v LobbyView) string {
	left := renderLobbySettings(b, v)
	right := renderLobbyPlayers(b, v)

	rows := max(len(left), len(right))
	lines := make([]string, rows)
	half := lipgloss.NewStyle().Width(lobbyHalfWidth)
	for i := range lines {
		var l, r string
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		lines[i] = half.Render(l) + r
	}

	hints := "[Tab] Rename  │  [ESC] Leave"
	switch {
	case v.EditingName:
		hints = "Type your name  │  [Enter] Save  │  [ESC] Cancel"
	case v.Wheel != nil:
		hints = "[↑↓] Scroll  │  [Enter] Confirm  │  [ESC] Cancel"
	case v.KickPicker != nil:
		hints = "[↑↓] Select  │  [Enter] Kick  │  [ESC] Cancel"
	case b.IsHost():
		hints = "[↑↓] Select  │  [Enter] Open  │  [K] Kick  │  [Tab] Rename  │  [ESC] Close"
	}

	// The player count lives in the header so four cards fill the right half
	header := fmt.Sprintf("Room  │  %s  │  Players %d/%d", v.Address, len(b.Players), online.MaxPlayers)
	if v.KickPicker != nil {
		header = "Kick which player?"
	}
	return renderOnlineScreen(header, lines, hints)
}

// renderLobbySettings renders the left half: the menu, or the wheel in place
func renderLobbySettings(b *online.Battle, v LobbyView) []string {
	lines := make([]string, lobbyMenuTop)

	for i, item := range LobbyItems(b) {
		selected := b.IsHost() && i == v.Selected
		if v.Wheel != nil && !selected {
			// Hidden while adjusting; the blank row keeps the others in place
			lines = append(lines, "")
			continue
		}
		marker, labelStyle := "  ", statItemStyle
		if selected {
			marker, labelStyle = highlightStyle.Render("> "), highlightStyle
		}
		if item.Value == "" {
			lines = append(lines, marker+labelStyle.Render(item.Label))
			continue
		}
		label := marker + labelStyle.Render(fmt.Sprintf("%-*s", lobbyLabelWidth, item.Label))
		if selected && v.Wheel != nil {
			// The wheel's ▶ takes the gap, so the value keeps its column
			lines = append(lines, label)
			continue
		}
		lines = append(lines, label+"  "+statValueStyle.Render(item.Value))
	}

	if v.Wheel != nil {
		overlayWheel(lines, lobbyMenuTop+v.Selected, *v.Wheel)
	}

	lines = append(lines, "")
	if !b.IsHost() {
		lines = append(lines, "  "+statsStyle.Render("Waiting for host..."))
	}
	if v.Notice != "" {
		lines = append(lines, "  "+onlineErrorStyle.Render(v.Notice))
	}
	return lines
}

// overlayWheel draws the wheel in the value column, centered on row: the
// current step bracketed and bright, the two steps on each side fading out
func overlayWheel(lines []string, row int, w Wheel) {
	accent := lipgloss.NewStyle().Foreground(getRandomMenuColor(0)).Bold(true)
	current := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	pad := strings.Repeat(" ", lobbyValueColumn)

	lines[row] += accent.Render("▶ ") + current.Render(w.Steps[w.Index]) + accent.Render(" ◀")
	for offset := 1; offset <= lobbyWheelReach; offset++ {
		style := wheelNearStyle
		if offset == lobbyWheelReach {
			style = wheelFarStyle
		}
		for _, sign := range []int{-1, 1} {
			i, r := w.Index+sign*offset, row+sign*offset
			if i < 0 || i >= len(w.Steps) || r < 0 || r >= len(lines) {
				continue
			}
			lines[r] = pad + style.Render(w.Steps[i])
		}
	}
}

// renderLobbyPlayers renders the right half: one framed card per player.
// While the kick picker is open the chosen card turns red.
func renderLobbyPlayers(b *online.Battle, v LobbyView) []string {
	var lines []string

	kickable := make(map[int]int) // player id -> position in the picker
	for i, id := range v.KickPicker {
		kickable[id] = i
	}

	for _, p := range b.Players {
		name := p.Name
		if p.ID == b.MyID && v.EditingName {
			name = v.NameInput + "_"
		}
		tags := ""
		if p.IsHost {
			tags += " host"
		}
		if p.ID == b.MyID {
			tags += " you"
		}

		color := playerColors[b.ColorSlot(p.ID)%len(playerColors)]
		border := lipgloss.Color("240")
		content := playerStyle(b, p.ID).Render(name) + statsStyle.Render(tags)
		if i, ok := kickable[p.ID]; ok && i == v.KickIndex {
			border = lipgloss.Color("196")
			content = onlineErrorStyle.Render("✕ "+name) + statsStyle.Render(tags)
		} else if v.KickPicker != nil && !ok {
			content = onlineDisabledStyle.Render(name + tags)
		} else if p.ID == b.MyID {
			border = color
		}

		card := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Width(lobbyCardWidth-2).
			Padding(0, 1).
			Render(content)
		lines = append(lines, strings.Split(card, "\n")...)
	}
	return lines
}

// battleHeader names the mode with its progress and the time left
func battleHeader(b *online.Battle, now time.Time) string {
	parts := []string{online.ModeName(b.Mode)}
	switch b.Mode {
	case online.ModeWordDrop:
		parts = append(parts, fmt.Sprintf("Word %d/%d", b.WordIndex+1, b.TotalWords))
	case online.ModeSentenceRace:
		parts = append(parts, fmt.Sprintf("Sentence %d/%d", b.WordIndex+1, b.TotalWords))
	case online.ModeWordPool:
		if b.Settings.AllowSteal {
			parts = append(parts, "Steal ON")
		} else {
			parts = append(parts, "Steal OFF")
		}
	}
	parts = append(parts, fmt.Sprintf("%ds left", b.RemainingSeconds(now)))
	return strings.Join(parts, "  │  ")
}

// battleMessage is the countdown or the last result, shown mid-screen
func battleMessage(b *online.Battle, now time.Time) string {
	switch b.Phase {
	case online.BattleCountdown:
		return hitEffectStyle.Render(fmt.Sprintf("Get ready: %d", b.CountdownSeconds(now)))
	case online.BattleBetween:
		if b.FeedbackHit {
			return highlightStyle.Render(b.Feedback)
		}
		return onlineErrorStyle.Render(b.Feedback)
	}
	return ""
}

// renderFallingWord highlights the part of the word the input already matches
func renderFallingWord(word, input string) string {
	if input != "" && strings.HasPrefix(word, input) {
		return highlightStyle.Render(input) + wordStyle.Render(word[len(input):])
	}
	return wordStyle.Render(word)
}

// liveInput returns what a player is typing: the local input for this
// player, the relayed input for everyone else
func liveInput(b *online.Battle, id int) string {
	if id == b.MyID {
		return b.Input
	}
	return b.Typing[id]
}

// liveOrder lists the other players first and the local player last, so
// your own row always sits right above the hints
func liveOrder(b *online.Battle) []online.PlayerInfo {
	order := make([]online.PlayerInfo, 0, len(b.Players))
	var me *online.PlayerInfo
	for i, p := range b.Players {
		if p.ID == b.MyID {
			me = &b.Players[i]
			continue
		}
		order = append(order, p)
	}
	if me != nil {
		order = append(order, *me)
	}
	return order
}

// renderScoreBar renders every player's name and score on one line
func renderScoreBar(b *online.Battle) string {
	parts := make([]string, 0, len(b.Players))
	for _, p := range liveOrder(b) {
		parts = append(parts, playerStyle(b, p.ID).Render(displayName(b, p))+" "+statValueStyle.Render(fmt.Sprintf("%d", b.Scores[p.ID])))
	}
	return strings.Join(parts, separatorStyle.Render(" │ "))
}

// displayName shows the local player as "You"
func displayName(b *online.Battle, p online.PlayerInfo) string {
	if p.ID == b.MyID {
		return "You"
	}
	return p.Name
}

// shortName fits a long name into six characters, keeping both ends:
// AVeryLongNam becomes AVe*am, five letters around the star
func shortName(name string) string {
	runes := []rune(name)
	if len(runes) <= sentenceNameWidth {
		return name
	}
	return string(runes[:3]) + "*" + string(runes[len(runes)-2:])
}

// typedStyle is how a player's correct letters look: green for you, gray
// for the others, or their own color where told to (the word pool)
func typedStyle(b *online.Battle, id int, ownColors bool) lipgloss.Style {
	switch {
	case id == b.MyID:
		return highlightStyle
	case ownColors:
		return lipgloss.NewStyle().Foreground(playerColors[b.ColorSlot(id)%len(playerColors)])
	}
	return wheelNearStyle
}

// alignedInput renders input letter by letter under target: each letter
// that matches the target letter above it in good, each one that does not
// in red. A wrong space shows as a red dot so the mistake stays visible.
func alignedInput(target, input string, good lipgloss.Style) string {
	var out strings.Builder
	run, runOK := "", true
	flush := func() {
		if run == "" {
			return
		}
		if runOK {
			out.WriteString(good.Render(run))
		} else {
			out.WriteString(onlineErrorStyle.Render(run))
		}
		run = ""
	}
	for i := 0; i < len(input); i++ {
		ok := i < len(target) && input[i] == target[i]
		if ok != runOK {
			flush()
			runOK = ok
		}
		ch := input[i]
		if !ok && ch == ' ' {
			run += "·"
			continue
		}
		run += string(ch)
	}
	flush()
	return out.String()
}

// prefixMatch counts the leading input characters that match target
func prefixMatch(target, input string) int {
	n := 0
	for n < len(input) && n < len(target) && input[n] == target[n] {
		n++
	}
	return n
}

// liveNameWidth is the name column of the word modes: a full name and a gap
const liveNameWidth = 14

// maxShownInput keeps a relayed input, which may be long, inside its row
const maxShownInput = 40

// liveNameCell renders a name in the player's color, padded to width
func liveNameCell(b *online.Battle, p online.PlayerInfo, name string, width int) string {
	return playerStyle(b, p.ID).Render(name) + strings.Repeat(" ", max(1, width-len([]rune(name))))
}

// shownInput is what a player has typed, as far as it fits a row
func shownInput(b *online.Battle, id int) string {
	input := liveInput(b, id)
	if len(input) > maxShownInput {
		input = input[:maxShownInput]
	}
	return input
}

// padTop adds blank rows above, so your row stays on the bottom line
// whoever is in the room
func padTop(rows []string, n int) []string {
	for len(rows) < n {
		rows = append([]string{""}, rows...)
	}
	return rows
}

// RenderBattle renders the running game in the layout of its mode
func RenderBattle(b *online.Battle, now time.Time) string {
	switch b.Mode {
	case online.ModeSentenceRace:
		return renderSentenceRace(b, now)
	case online.ModeWordPool:
		return renderWordPool(b, now)
	}
	return renderWordDrop(b, now)
}

func renderWordDrop(b *online.Battle, now time.Time) string {
	lines := []string{centerLine(renderScoreBar(b))}

	wordRow := -1
	message := battleMessage(b, now)
	if b.Phase == online.BattleFalling {
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

	// The word again, with everyone's input lined up under it letter by letter
	indent := "  " + strings.Repeat(" ", liveNameWidth)
	lines = append(lines, indent+titleStyle.Render(b.Word))
	var rows []string
	for _, p := range liveOrder(b) {
		input := shownInput(b, p.ID)
		text := alignedInput(b.Word, input, typedStyle(b, p.ID, false))
		switch {
		case p.ID == b.MyID:
			text += inputStyle.Render("_")
		case input == "":
			text = wheelFarStyle.Render("...")
		}
		rows = append(rows, "  "+liveNameCell(b, p, displayName(b, p), liveNameWidth)+text)
	}
	lines = append(lines, padTop(rows, liveRows)...)
	return renderOnlineScreen(battleHeader(b, now), lines, "Type the word  │  [Enter] Submit  │  [ESC] Leave")
}

// Sentence rows: a name of up to six characters (five letters and the star
// of a shortened name), a space, then the text
const (
	sentenceNameWidth = 6
	sentenceIndent    = 2 + sentenceNameWidth + 1
	// sentenceWidth leaves room for the space kept at the end of a wrapped
	// line and the cursor; every sentence in the data fits in two lines
	sentenceWidth = onlineInnerWidth - sentenceIndent - 2
)

// wrapRanges splits text at spaces into lines of at most width characters,
// returning [start, end) byte ranges. A line keeps its trailing space, so
// the ranges cover every character and input can be cut the same way.
func wrapRanges(text string, width int) [][2]int {
	var ranges [][2]int
	start, end := 0, 0
	for end < len(text) {
		next := strings.IndexByte(text[end:], ' ')
		wordEnd := len(text)
		if next >= 0 {
			wordEnd = end + next
		}
		if wordEnd-start > width && end > start {
			ranges = append(ranges, [2]int{start, end})
			start = end
		}
		end = wordEnd
		if end < len(text) {
			end++ // keep the space with the word before it
		}
	}
	if start < len(text) {
		ranges = append(ranges, [2]int{start, len(text)})
	}
	return ranges
}

func renderSentenceRace(b *online.Battle, now time.Time) string {
	hints := "Type the sentence  │  [Enter] Submit  │  [ESC] Leave"
	if b.Hint != "" {
		hints = b.Hint + "  │  [Backspace] Fix  │  [ESC] Leave"
	}
	lines := []string{centerLine(renderScoreBar(b))}

	if message := battleMessage(b, now); message != "" {
		lines = append(lines, "", "", centerLine(message))
		return renderOnlineScreen(battleHeader(b, now), lines, hints)
	}

	players := liveOrder(b)
	ranges := wrapRanges(b.Word, sentenceWidth)

	// Start from the line you are on if the whole sentence would not fit
	first := 0
	for first < len(ranges)-1 && len(lines)+(len(ranges)-first)*(1+len(players)) > onlineContentHeight &&
		len(b.Input) >= ranges[first][1] {
		first++
	}

	indent := strings.Repeat(" ", sentenceIndent)
	for _, r := range ranges[first:] {
		lines = append(lines, indent+titleStyle.Render(b.Word[r[0]:r[1]]))
		last := r[1] == len(b.Word)
		for _, p := range players {
			input := liveInput(b, p.ID)
			if len(input) > len(b.Word) {
				input = input[:len(b.Word)]
			}
			text := ""
			if input := clip(input, r[0], r[1]); input != "" {
				text = alignedInput(b.Word[r[0]:r[1]], input, typedStyle(b, p.ID, false))
			}
			// The cursor sits on the line your next letter goes to
			if p.ID == b.MyID && len(input) >= r[0] && (len(input) < r[1] || last) {
				text += inputStyle.Render("_")
			}
			name := shortName(displayName(b, p))
			lines = append(lines, "  "+liveNameCell(b, p, name, sentenceNameWidth+1)+text)
		}
	}
	if len(lines) > onlineContentHeight {
		lines = lines[:onlineContentHeight]
	}
	return renderOnlineScreen(battleHeader(b, now), lines, hints)
}

// clip returns the part of input that falls in [start, end)
func clip(input string, start, end int) string {
	if start >= len(input) {
		return ""
	}
	return input[start:min(end, len(input))]
}

// poolTarget picks the board word a player is most likely typing: the one
// sharing the longest prefix with the input, the first slot on ties
func poolTarget(b *online.Battle, input string) string {
	best, bestN := "", 0
	for _, c := range b.Pool {
		if c.Word == "" {
			continue
		}
		if n := prefixMatch(c.Word, input); n > bestN {
			best, bestN = c.Word, n
		}
	}
	return best
}

func renderWordPool(b *online.Battle, now time.Time) string {
	const cellWidth = onlineInnerWidth / online.PoolColumns
	lines := []string{centerLine(renderScoreBar(b))}

	if message := battleMessage(b, now); message != "" {
		lines = append(lines, "", "", centerLine(message))
		for len(lines) < onlineContentHeight-liveRows {
			lines = append(lines, "")
		}
	} else {
		// Rows sit flush so the board and the live rows share the screen
		cell := lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center)
		for row := 0; row < online.PoolSlots/online.PoolColumns; row++ {
			var cells []string
			for col := 0; col < online.PoolColumns; col++ {
				c := b.Pool[row*online.PoolColumns+col]
				cells = append(cells, cell.Render(renderPoolWord(b, c)))
			}
			lines = append(lines, strings.Join(cells, ""))
		}
		lines = append(lines, separatorStyle.Render(strings.Repeat("━", onlineInnerWidth)))
	}

	// Each row names the word that player seems to be typing, then their
	// input checked against it letter by letter, in the player's own color
	var rows []string
	for _, p := range liveOrder(b) {
		input := shownInput(b, p.ID)
		target := ""
		if input != "" {
			target = poolTarget(b, input)
		}
		name := "  " + liveNameCell(b, p, displayName(b, p), liveNameWidth)
		if input == "" && p.ID != b.MyID {
			// Nothing typed yet: no target to show
			rows = append(rows, name+wheelFarStyle.Render("..."))
			continue
		}
		targetCell := wordStyle.Render(target) + strings.Repeat(" ", online.MaxPoolWordLength-len(target))
		text := alignedInput(target, input, typedStyle(b, p.ID, true))
		if p.ID == b.MyID {
			text += inputStyle.Render("_")
		}
		rows = append(rows, name+targetCell+separatorStyle.Render(" → ")+text)
	}
	lines = append(lines, padTop(rows, liveRows)...)
	return renderOnlineScreen(battleHeader(b, now), lines, "Type a word  │  [Enter] Claim  │  [ESC] Leave")
}

// renderPoolWord colors a board word by its owner and highlights the
// part the input matches. Words the server would never send, longer than a
// cell, are cut so a bad packet cannot wrap and break the board.
func renderPoolWord(b *online.Battle, c online.PoolCell) string {
	if c.Word == "" {
		return wheelFarStyle.Render("·")
	}
	word := c.Word
	if len(word) > online.MaxPoolWordLength {
		word = word[:online.MaxPoolWordLength]
	}
	style := wordStyle
	if c.Owner != 0 {
		style = playerStyle(b, c.Owner)
	}
	if b.MatchesPrefix(word) {
		return lipgloss.NewStyle().Underline(true).Inherit(style).Render(b.Input) + style.Render(word[len(b.Input):])
	}
	return style.Render(word)
}

// RenderBattleResult renders the final ranking
func RenderBattleResult(b *online.Battle) string {
	itemLabel := "Words"
	if b.Mode == online.ModeSentenceRace {
		itemLabel = "Lines"
	}
	lines := []string{
		centerLine(titleStyle.Render("BATTLE RESULT")),
		"",
		centerLine(statItemStyle.Render(fmt.Sprintf("%-4s %-14s %6s %6s", "#", "Player", "Score", itemLabel)) + "  "),
		centerLine(separatorStyle.Render(strings.Repeat("━", 33)) + "  "),
	}

	for i, r := range b.Ranking {
		style := playerStyle(b, r.ID)
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
