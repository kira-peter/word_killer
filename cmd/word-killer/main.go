package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/word-killer/word-killer/pkg/config"
	"github.com/word-killer/word-killer/pkg/game"
	"github.com/word-killer/word-killer/pkg/ui"
)

// tickMsg is sent on every tick to update the display
type tickMsg time.Time

// model is the Bubble Tea model
type model struct {
	game             *game.Game
	cfg              *config.Config
	ready            bool
	showModeSelect   bool // true when showing mode selection screen
	showAbout        bool // true when showing about page
	selectedMode     int  // 0=Classic, 1=Sentence, 2=Underwater
	width            int
	height           int
	animFrame        int                       // animation frame counter for pause menu
	welcomeAnimState *ui.WelcomeAnimationState // welcome screen animation state
	tickCount        int                       // tick 计数器，用于控制游戏逻辑更新频率
}

func initialModel(cfg *config.Config, g *game.Game) model {
	return model{
		game:             g,
		cfg:              cfg,
		ready:            false,
		showModeSelect:   false,
		showAbout:        false,
		selectedMode:     0,
		welcomeAnimState: &ui.WelcomeAnimationState{},
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*33, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Increment tick counter and animation frame (30 FPS)
		m.tickCount++
		m.animFrame++

		// Update game logic only every 3 ticks (~100ms, 10 updates/second)
		// This keeps animation speeds the same while rendering at 30 FPS
		if m.tickCount%3 == 0 {
			// Update welcome animation if on welcome screen
			if !m.ready && !m.showModeSelect && !m.showAbout {
				ui.UpdateWelcomeAnimation(m.welcomeAnimState)
			}

			// Update underwater animations if in underwater mode
			if m.ready && m.game.Mode == game.ModeUnderwaterCountdown && m.game.Status == game.StatusRunning {
				m.game.UpdateFishPositions()
				m.game.UpdateBackgroundAnimation()
				m.game.UpdateCountdown()
			}
		}

		// Always return tick command to keep rendering at 30 FPS
		return m, tickCmd()
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// About screen
	if !m.ready && m.showAbout {
		switch msg.String() {
		case "esc", "enter", "ctrl+c":
			// Go back to welcome screen
			m.showAbout = false
			return m, nil
		}
		return m, nil
	}

	// Welcome screen
	if !m.ready && !m.showModeSelect && !m.showAbout {
		switch msg.String() {
		case "up", "k":
			// Move selection up (3 options now)
			m.welcomeAnimState.SelectedOption = (m.welcomeAnimState.SelectedOption - 1 + 3) % 3
			return m, nil
		case "down", "j":
			// Move selection down (3 options now)
			m.welcomeAnimState.SelectedOption = (m.welcomeAnimState.SelectedOption + 1) % 3
			return m, nil
		case "enter":
			// Confirm selection
			if m.welcomeAnimState.SelectedOption == 0 {
				// Start selected - show mode selection
				m.showModeSelect = true
				m.selectedMode = 0
			} else if m.welcomeAnimState.SelectedOption == 1 {
				// About selected
				m.showAbout = true
			} else {
				// Quit selected
				return m, tea.Quit
			}
			return m, nil
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}

	// Mode selection screen
	if !m.ready && m.showModeSelect {
		switch msg.String() {
		case "up", "k":
			// Move selection up (3 modes)
			m.selectedMode = (m.selectedMode - 1 + 3) % 3
			return m, nil
		case "down", "j":
			// Move selection down (3 modes)
			m.selectedMode = (m.selectedMode + 1) % 3
			return m, nil
		case "enter":
			// Start game with selected mode
			var err error
			if m.selectedMode == 0 {
				// Classic mode
				err = m.game.Start(m.cfg.WordCount)
			} else if m.selectedMode == 1 {
				// Sentence mode
				err = m.game.StartSentenceMode()
			} else {
				// Underwater countdown mode
				err = m.game.StartUnderwaterCountdown(m.cfg.CountdownDuration)
			}
			if err != nil {
				return m, tea.Quit
			}
			m.ready = true
			return m, nil
		case "esc":
			// Go back to welcome screen
			m.showModeSelect = false
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}

	// Game running
	if m.game.Status == game.StatusRunning {
		switch msg.String() {
		case "esc":
			m.game.Pause()
		case "enter":
			m.game.TryEliminate()
		case "backspace":
			m.game.Backspace()
		default:
			// Handle input based on game mode
			runes := []rune(msg.String())
			if len(runes) == 1 {
				r := runes[0]
				if m.game.Mode == game.ModeSentence {
					// Sentence mode: accept all printable characters
					if r >= 32 && r <= 126 {
						m.game.AddChar(r)
					}
				} else {
					// Classic mode: only letters
					if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
						if r >= 'A' && r <= 'Z' {
							r = r + 32
						}
						m.game.AddChar(r)
					}
				}
			}
		}

		// Check game status
		if m.game.Status == game.StatusFinished {
			// Game ended, wait for user to press a key then exit
			return m, nil
		}

		return m, nil
	} else if m.game.Status == game.StatusPaused {
		// Pause menu - 4 options: Resume, Restart, Select Mode, Main Menu
		switch msg.String() {
		case "up", "k":
			m.game.MovePauseMenu(-1)
		case "down", "j":
			m.game.MovePauseMenu(1)
		case "enter":
			idx := m.game.PauseMenuIndex
			if idx == 0 {
				// Resume Game
				m.game.Resume()
			} else if idx == 1 {
				// Restart - same mode
				if m.game.Mode == game.ModeSentence {
					m.game.StartSentenceMode()
				} else if m.game.Mode == game.ModeUnderwaterCountdown {
					m.game.StartUnderwaterCountdown(m.cfg.CountdownDuration)
				} else {
					m.game.Start(m.cfg.WordCount)
				}
			} else if idx == 2 {
				// Select Mode - go back to mode selection
				m.ready = false
				m.showModeSelect = true
				m.selectedMode = 0
			} else if idx == 3 {
				// Main Menu - go back to welcome
				m.ready = false
				m.showModeSelect = false
				m.showAbout = false
				m.welcomeAnimState.SelectedOption = 0
			}
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	} else if m.game.Status == game.StatusFinished {
		// Results page - 3 options: Restart, Select Mode, Main Menu
		switch msg.String() {
		case "up", "k":
			m.game.MoveResultsMenu(-1)
		case "down", "j":
			m.game.MoveResultsMenu(1)
		case "enter":
			idx := m.game.ResultsMenuIndex
			if idx == 0 {
				// Restart - same mode
				if m.game.Mode == game.ModeSentence {
					m.game.StartSentenceMode()
				} else if m.game.Mode == game.ModeUnderwaterCountdown {
					m.game.StartUnderwaterCountdown(m.cfg.CountdownDuration)
				} else {
					m.game.Start(m.cfg.WordCount)
				}
			} else if idx == 1 {
				// Select Mode - go back to mode selection
				m.ready = false
				m.showModeSelect = true
				m.selectedMode = 0
			} else if idx == 2 {
				// Main Menu - go back to welcome
				m.ready = false
				m.showModeSelect = false
				m.showAbout = false
				m.welcomeAnimState.SelectedOption = 0
			}
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	// About screen
	if !m.ready && m.showAbout {
		return ui.RenderAbout()
	}

	// Welcome screen
	if !m.ready && !m.showModeSelect {
		return ui.RenderWelcome(m.welcomeAnimState, m.animFrame)
	}

	// Mode selection screen
	if !m.ready && m.showModeSelect {
		return ui.RenderModeSelection(m.selectedMode, m.animFrame)
	}

	if m.game.Status == game.StatusRunning {
		// Render based on game mode
		if m.game.Mode == game.ModeUnderwaterCountdown {
			// Underwater countdown mode rendering
			return ui.RenderUnderwaterGame(m.game)
		} else if m.game.Mode == game.ModeSentence {
			// Sentence mode rendering
			stats := ui.GameStats{
				TotalKeystrokes:  m.game.Stats.TotalKeystrokes,
				ValidKeystrokes:  m.game.Stats.ValidKeystrokes,
				CorrectChars:     m.game.Stats.CorrectChars,
				WordsCompleted:   m.game.Stats.WordsCompleted,
				TotalLetters:     m.game.Stats.TotalLetters,
				ElapsedSeconds:   m.game.Stats.GetElapsedSeconds(),
				LettersPerSecond: m.game.Stats.GetLettersPerSecond(),
				WordsPerSecond:   m.game.Stats.GetWordsPerSecond(),
				AccuracyPercent:  m.game.Stats.GetAccuracyPercent(),
			}
			return ui.RenderSentenceGame(m.game.TargetSentence, m.game.InputBuffer, stats)
		}

		// Classic mode rendering (existing code)
		// Get all words (including completed ones)
		allWords := m.game.GetAllWords()
		wordInfos := make([]ui.WordInfo, len(allWords))
		for i, w := range allWords {
			wordInfos[i] = ui.WordInfo{
				Text:        w.Text,
				Completed:   w.Completed,
				CompletedAt: w.CompletedAt,
			}
		}

		highlighted := m.game.GetMatchedIndices()
		activeWords := m.game.GetActiveWords()

		stats := ui.GameStats{
			TotalKeystrokes:  m.game.Stats.TotalKeystrokes,
			ValidKeystrokes:  m.game.Stats.ValidKeystrokes,
			CorrectChars:     m.game.Stats.CorrectChars,
			WordsCompleted:   m.game.Stats.WordsCompleted,
			TotalLetters:     m.game.Stats.TotalLetters,
			ElapsedSeconds:   m.game.Stats.GetElapsedSeconds(),
			LettersPerSecond: m.game.Stats.GetLettersPerSecond(),
			WordsPerSecond:   m.game.Stats.GetWordsPerSecond(),
			AccuracyPercent:  m.game.Stats.GetAccuracyPercent(),
		}

		return ui.RenderGame(wordInfos, highlighted, m.game.InputBuffer, stats, len(activeWords))
	} else if m.game.Status == game.StatusPaused {
		// Pass stats and animation frame to pause menu
		activeWords := m.game.GetActiveWords()
		stats := ui.GameStats{
			TotalKeystrokes:  m.game.Stats.TotalKeystrokes,
			ValidKeystrokes:  m.game.Stats.ValidKeystrokes,
			CorrectChars:     m.game.Stats.CorrectChars,
			WordsCompleted:   m.game.Stats.WordsCompleted,
			TotalLetters:     m.game.Stats.TotalLetters,
			ElapsedSeconds:   m.game.Stats.GetElapsedSeconds(),
			LettersPerSecond: m.game.Stats.GetLettersPerSecond(),
			WordsPerSecond:   m.game.Stats.GetWordsPerSecond(),
			AccuracyPercent:  m.game.Stats.GetAccuracyPercent(),
		}
		return ui.RenderPauseMenu(m.game.PauseMenuIndex, stats, len(activeWords), m.animFrame)
	} else if m.game.Status == game.StatusFinished {
		stats := ui.GameStats{
			TotalKeystrokes:  m.game.Stats.TotalKeystrokes,
			ValidKeystrokes:  m.game.Stats.ValidKeystrokes,
			CorrectChars:     m.game.Stats.CorrectChars,
			WordsCompleted:   m.game.Stats.WordsCompleted,
			TotalLetters:     m.game.Stats.TotalLetters,
			ElapsedSeconds:   m.game.Stats.GetElapsedSeconds(),
			LettersPerSecond: m.game.Stats.GetLettersPerSecond(),
			WordsPerSecond:   m.game.Stats.GetWordsPerSecond(),
			AccuracyPercent:  m.game.Stats.GetAccuracyPercent(),
		}

		return ui.RenderResults(stats, m.game.Aborted, m.game.ResultsMenuIndex, m.animFrame)
	}

	return ""
}

func main() {
	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create game instance
	g := game.New()

	// Normalize difficulty ratios
	shortRatio, mediumRatio, longRatio, err := cfg.NormalizeRatios()
	if err != nil {
		fmt.Printf("Invalid difficulty ratios: %v\n", err)
		os.Exit(1)
	}

	// Load word dictionaries
	if err := g.LoadWordDictionaries(
		cfg.ShortDictPath,
		cfg.MediumDictPath,
		cfg.LongDictPath,
		shortRatio,
		mediumRatio,
		longRatio,
	); err != nil {
		fmt.Printf("Failed to load word dictionaries: %v\n", err)
		os.Exit(1)
	}

	// Load sentences for sentence mode
	if err := g.LoadSentences(cfg.SentenceDictPath); err != nil {
		fmt.Printf("Warning: Failed to load sentences: %v\n", err)
		// Continue anyway - classic mode will still work
	}

	// Create Bubble Tea program
	p := tea.NewProgram(
		initialModel(cfg, g),
		tea.WithAltScreen(),       // use alternate screen buffer
		tea.WithMouseCellMotion(), // enable mouse support (optional)
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Failed to run: %v\n", err)
		os.Exit(1)
	}
}
