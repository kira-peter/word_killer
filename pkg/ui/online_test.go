package ui

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/word-killer/word-killer/pkg/online"
)

func testRooms(n int) []online.RoomInfo {
	rooms := make([]online.RoomInfo, n)
	for i := range rooms {
		rooms[i] = online.RoomInfo{ID: fmt.Sprint(i), HostName: fmt.Sprintf("Host%02d", i), IP: "10.0.0.1", Port: 7788, Players: 1, MaxPlayers: 4}
	}
	return rooms
}

func TestOnlineMenuKeepsStandardHeight(t *testing.T) {
	want := lipgloss.Height(RenderOnlineMenu(nil, 0, "", false, 0))
	for _, n := range []int{0, 1, maxListedRooms, maxListedRooms + 1, 20} {
		rooms := testRooms(n)
		for _, selected := range []int{0, n} {
			view := RenderOnlineMenu(rooms, selected, "Cannot connect to host", false, 0)
			if got := lipgloss.Height(view); got != want {
				t.Errorf("%d rooms, selected %d: height %d, want %d", n, selected, got, want)
			}
		}
	}
}

func TestSearchingAnimation(t *testing.T) {
	// The dots count up, clear, and start over
	var seen []string
	for frame := 0; frame < len(searchingFrames)*searchingFrameTicks; frame += searchingFrameTicks {
		view := RenderOnlineMenu(nil, 0, "", false, frame)
		line := lineWith(view, "Searching for rooms")
		if line == "" {
			t.Fatalf("frame %d has no searching line:\n%s", frame, view)
		}
		seen = append(seen, strings.TrimSpace(strings.Trim(line, "│ ")))
	}
	want := []string{"Searching for rooms.", "Searching for rooms..", "Searching for rooms...", "Searching for rooms"}
	if strings.Join(seen, "|") != strings.Join(want, "|") {
		t.Errorf("frames = %q, want %q", seen, want)
	}

	// Every frame is the same width, so the text never shifts
	widths := map[int]bool{}
	for frame := 0; frame < len(searchingFrames)*searchingFrameTicks; frame++ {
		widths[lipgloss.Width(lineWith(RenderOnlineMenu(nil, 0, "", false, frame), "Searching"))] = true
	}
	if len(widths) != 1 {
		t.Errorf("searching line width varies: %v", widths)
	}
}

func TestMenuLayoutPositions(t *testing.T) {
	view := RenderOnlineMenu(nil, 0, "Host closed the room", false, 0)
	lines := strings.Split(view, "\n")
	search, notice, boxEnd := -1, -1, -1
	for i, l := range lines {
		switch {
		case strings.Contains(l, "Searching"):
			search = i
		case strings.Contains(l, "Host closed the room"):
			notice = i
		case strings.HasPrefix(l, "└"):
			boxEnd = i
		}
	}
	// Searching sits mid-list; the notice is on the last content line, just
	// above the box's bottom padding row
	if notice != boxEnd-2 {
		t.Errorf("notice on line %d, box ends on %d:\n%s", notice, boxEnd, view)
	}
	listTop := lineIndex(lines, "Rooms on your network") + 1
	if search-listTop != menuListRows/2 {
		t.Errorf("searching on list row %d, want %d:\n%s", search-listTop, menuListRows/2, view)
	}

	// With rooms listed the notice stays on that same bottom line
	withRooms := strings.Split(RenderOnlineMenu(testRooms(20), 20, "Room is full", false, 0), "\n")
	if lineIndex(withRooms, "Room is full") != notice {
		t.Errorf("notice moved with a full list:\n%s", strings.Join(withRooms, "\n"))
	}
}

func lineWith(view, text string) string {
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(l, text) {
			return l
		}
	}
	return ""
}

func lineIndex(lines []string, text string) int {
	for i, l := range lines {
		if strings.Contains(l, text) {
			return i
		}
	}
	return -1
}

// liveBattle is a running game with the local player (Alice, the host)
// listed first, so tests can check that "You" still renders last
func liveBattle(mode string) *online.Battle {
	now := time.Now()
	b := online.NewBattle(1)
	settings := online.DefaultSettings(mode)
	b.Apply(online.Message{Type: online.TypeLobby, Mode: mode, Settings: &settings,
		Players: []online.PlayerInfo{{ID: 1, Name: "Alice", IsHost: true}, {ID: 2, Name: "Bob"}, {ID: 3, Name: "AVeryLongNam"}}}, now)
	b.Apply(online.Message{Type: online.TypeGameStart, Mode: mode, Settings: &settings, TotalWords: 5}, now)
	return b
}

func TestAlignedInput(t *testing.T) {
	good := wheelNearStyle

	// Only the wrong letters are red; the letters after a mistake are judged
	// on their own against the letter above them
	got := alignedInput("keyboard", "kexboard", good)
	if want := good.Render("ke") + onlineErrorStyle.Render("x") + good.Render("board"); got != want {
		t.Errorf("one wrong letter: got %q, want %q", got, want)
	}

	// Letters past the end of the target are wrong too
	got = alignedInput("go", "gone", good)
	if want := good.Render("go") + onlineErrorStyle.Render("ne"); got != want {
		t.Errorf("too long: got %q", got)
	}

	// A wrong space shows as a red dot instead of vanishing
	got = alignedInput("ab", "a ", good)
	if want := good.Render("a") + onlineErrorStyle.Render("·"); got != want {
		t.Errorf("wrong space: got %q", got)
	}
}

func TestShortName(t *testing.T) {
	for in, want := range map[string]string{"Bob": "Bob", "Alices": "Alices", "AVeryLongNam": "AVe*am", "Abcdefg": "Abc*fg"} {
		if got := shortName(in); got != want {
			t.Errorf("shortName(%q) = %q, want %q", in, got, want)
		}
	}
}

// col returns the display column of text in line, or -1. Box borders are
// multi-byte, so byte offsets would not line up.
func col(line, text string) int {
	i := strings.Index(line, text)
	if i < 0 {
		return -1
	}
	return lipgloss.Width(line[:i])
}

// rowFor finds a player's live row: the first line below the board or word
// separator that mentions them, skipping the score bar
func rowFor(view, who string) string {
	lines := strings.Split(view, "\n")
	for i := lineIndex(lines, "━━"); i >= 0 && i < len(lines); i++ {
		if strings.Contains(lines[i], who) {
			return lines[i]
		}
	}
	return ""
}

func TestWordDropAlignedRows(t *testing.T) {
	b := liveBattle(online.ModeWordDrop)
	b.Apply(online.Message{Type: online.TypeWordSpawn, Word: "keyboard", FallMs: 8000}, time.Now())
	b.Apply(online.Message{Type: online.TypeTyping, PlayerID: 2, Word: "kez"}, time.Now())
	b.AddChar('k')
	b.AddChar('x')
	view := RenderBattle(b, time.Now())

	// The word is printed again with every input starting in its column
	lines := strings.Split(view, "\n")
	wordLine := lines[lineIndex(lines, "━━")+1]
	wordCol := col(wordLine, "keyboard")
	for _, who := range []string{"Bob ", "You "} {
		line := rowFor(view, who)
		if col(line, "k") != wordCol {
			t.Errorf("%s input not under the word:\n%s\n%s", who, wordLine, line)
		}
	}

	// Others gray, you green, each wrong letter red
	if !strings.Contains(view, wheelNearStyle.Render("ke")+onlineErrorStyle.Render("z")) {
		t.Errorf("Bob should be gray with a red z:\n%s", view)
	}
	if !strings.Contains(view, highlightStyle.Render("k")+onlineErrorStyle.Render("x")+inputStyle.Render("_")) {
		t.Errorf("you should be green with a red x:\n%s", view)
	}

	// You are on the last row, after the others
	if lineIndex(lines, rowFor(view, "You ")) < lineIndex(lines, rowFor(view, "AVeryLongNam ")) {
		t.Errorf("you should be last:\n%s", view)
	}
}

func TestSentenceAlignedRows(t *testing.T) {
	b := liveBattle(online.ModeSentenceRace)
	sentence := "Nobody ever wrote down a plan to be broke, fat, lazy, or stupid. Those things are what happen when you don't have a plan."
	b.Apply(online.Message{Type: online.TypeSentence, Word: sentence}, time.Now())
	// The first wrapped line ends after "stupid. ", so Bob is into the second
	firstLineEnd := wrapRanges(sentence, sentenceWidth)[0][1]
	b.Apply(online.Message{Type: online.TypeTyping, PlayerID: 2, Word: sentence[:firstLineEnd] + "Thxse"}, time.Now())
	b.Apply(online.Message{Type: online.TypeTyping, PlayerID: 3, Word: "Nobody evre"}, time.Now())
	for _, r := range "Nobody ever" {
		b.AddChar(r)
	}
	view := RenderBattle(b, time.Now())
	lines := strings.Split(view, "\n")

	// Each wrapped sentence line is followed by every player's part of it,
	// starting in the same column
	first := lineIndex(lines, "Nobody ever wrote down")
	second := lineIndex(lines, "Those things are what")
	if first < 0 || second-first != 4 {
		t.Fatalf("sentence lines at %d and %d:\n%s", first, second, view)
	}
	sentenceCol := col(lines[first], "Nobody")
	for i, who := range []string{"Bob", "AVe*am", "You"} {
		line := lines[first+1+i]
		if !strings.Contains(line, who) || col(line, "Nobody") != sentenceCol {
			t.Errorf("row %d should be %s's input under the sentence:\n%s\n%s", i, who, lines[first], line)
		}
	}
	if col(lines[second+1], "Th") != col(lines[second], "Those") {
		t.Errorf("Bob's second line not aligned:\n%s\n%s", lines[second], lines[second+1])
	}

	// Single red letters: "evre" has two wrong letters, "Thxse" one
	if !strings.Contains(view, onlineErrorStyle.Render("re")) || !strings.Contains(view, onlineErrorStyle.Render("x")+wheelNearStyle.Render("se")) {
		t.Errorf("wrong letters should be red one by one:\n%s", view)
	}
	if lipgloss.Height(view) != lipgloss.Height(RenderOnlineMenu(nil, 0, "", false, 0)) {
		t.Errorf("height = %d", lipgloss.Height(view))
	}
}

func TestPoolAlignedRows(t *testing.T) {
	b := liveBattle(online.ModeWordPool)
	for i, w := range []string{"apple", "planet", "river"} {
		b.Apply(online.Message{Type: online.TypePoolSpawn, Index: i, Word: w}, time.Now())
	}
	b.Apply(online.Message{Type: online.TypeTyping, PlayerID: 2, Word: "plx"}, time.Now())
	for _, r := range "appk" {
		b.AddChar(r)
	}
	view := RenderBattle(b, time.Now())

	// Each row names the word being typed; others use their own color
	bobColor := lipgloss.NewStyle().Foreground(playerColors[b.ColorSlot(2)%len(playerColors)])
	if line := rowFor(view, "Bob "); !strings.Contains(line, "planet") || !strings.Contains(line, bobColor.Render("pl")+onlineErrorStyle.Render("x")) {
		t.Errorf("Bob's row = %q", line)
	}
	if line := rowFor(view, "You "); !strings.Contains(line, "apple") || !strings.Contains(line, highlightStyle.Render("app")+onlineErrorStyle.Render("k")) {
		t.Errorf("your row = %q", line)
	}
	// Nothing typed yet: no target and no arrow
	if line := rowFor(view, "AVeryLongNam "); line == "" || strings.Contains(line, "→") {
		t.Errorf("an idle player has no target: %q", line)
	}
}

func TestWrapRangesCoverEverySentence(t *testing.T) {
	data, err := os.ReadFile("../../data/sentences.txt")
	if err != nil {
		t.Skip("no sentence data")
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ranges := wrapRanges(line, sentenceWidth)
		// Two lines of five rows each, plus the score bar, fit the box
		if len(ranges) > 2 {
			t.Errorf("%d lines for %q", len(ranges), line)
		}
		joined := ""
		for _, r := range ranges {
			if r[1]-r[0] > sentenceWidth+1 {
				t.Errorf("line too wide: %q", line[r[0]:r[1]])
			}
			joined += line[r[0]:r[1]]
		}
		if joined != line {
			t.Errorf("ranges drop characters of %q", line)
		}
	}
}

func TestPlayerColorsAvoidRedAndGreen(t *testing.T) {
	for _, c := range playerColors {
		for _, banned := range []lipgloss.Color{"196", "203", "46", "82", "84"} {
			if c == banned {
				t.Errorf("player color %s looks like a right/wrong marker", c)
			}
		}
	}
}

func TestOnlineMenuScrollsToSelection(t *testing.T) {
	rooms := testRooms(20)

	view := RenderOnlineMenu(rooms, 0, "", false, 0)
	if !strings.Contains(view, "Host00") || strings.Contains(view, "Host19") || !strings.Contains(view, "14 more room(s)") {
		t.Errorf("top of list:\n%s", view)
	}

	view = RenderOnlineMenu(rooms, 20, "", false, 0)
	if !strings.Contains(view, "> Host19") || strings.Contains(view, "Host00") {
		t.Errorf("selection at the end must be visible:\n%s", view)
	}
}

// TestOnlineScreensKeepStandardHeight renders every online screen at its
// fullest and checks none grows taller than the room menu, so switching
// screens never makes the layout jump
func TestOnlineScreensKeepStandardHeight(t *testing.T) {
	want := lipgloss.Height(RenderOnlineMenu(nil, 0, "", false, 0))
	now := time.Now()
	players := []online.PlayerInfo{{ID: 1, Name: "AVeryLongNam", IsHost: true}, {ID: 2, Name: "Bob"}, {ID: 3, Name: "Carol"}, {ID: 4, Name: "Dan"}}

	views := map[string]string{}
	for _, mode := range online.Modes {
		b := online.NewBattle(1)
		settings := online.DefaultSettings(mode)
		b.Apply(online.Message{Type: online.TypeLobby, Mode: mode, Settings: &settings, Players: players}, now)
		views[mode+" lobby"] = RenderLobby(b, LobbyView{Notice: "Need at least 2 players", Address: "192.168.100.200:7788"})
		views[mode+" kick"] = RenderLobby(b, LobbyView{KickPicker: []int{2, 3, 4}})
		for i, o := range online.ModeOptions(mode) {
			var steps []string
			for _, v := range o.Values {
				steps = append(steps, o.Format(v))
			}
			views[mode+" wheel "+o.Key] = RenderLobby(b, LobbyView{Selected: i + 1, Wheel: &Wheel{Label: o.Label, Steps: steps, Index: len(steps) / 2}})
		}
		// The mode wheel opens on the top row, where its upper steps need room
		views[mode+" mode wheel"] = RenderLobby(b, LobbyView{Wheel: &Wheel{Label: "Mode", Steps: []string{"Word Drop", "Sentence Race", "Word Pool"}, Index: 2}})

		b.Apply(online.Message{Type: online.TypeGameStart, Mode: mode, Settings: &settings, TotalWords: 10}, now)
		long := strings.Repeat("word ", 24) + "end."
		switch mode {
		case online.ModeSentenceRace:
			b.Apply(online.Message{Type: online.TypeSentence, Word: long}, now)
			for _, r := range long + long {
				b.AddChar(r)
			}
			b.Submit()
		case online.ModeWordPool:
			for i := 0; i < online.PoolSlots; i++ {
				b.Apply(online.Message{Type: online.TypePoolSpawn, Index: i, Word: "internationalization"}, now)
			}
		default:
			b.Apply(online.Message{Type: online.TypeWordSpawn, Word: "keyboard", FallMs: 1000}, now)
		}
		views[mode+" battle"] = RenderBattle(b, now)

		b.Apply(online.Message{Type: online.TypeGameOver, Winners: []int{1}, Ranking: []online.RankEntry{
			{ID: 1, Name: "AVeryLongNam", Score: 99}, {ID: 2, Name: "Bob"}, {ID: 3, Name: "Carol"}, {ID: 4, Name: "Dan"}}}, now)
		views[mode+" result"] = RenderBattleResult(b)
	}

	for name, view := range views {
		if got := lipgloss.Height(view); got != want {
			t.Errorf("%s: height %d, want %d\n%s", name, got, want, view)
		}
		if got := lipgloss.Width(view); got != screenWidth {
			t.Errorf("%s: width %d, want %d", name, got, screenWidth)
		}
	}
}
