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
	game           *game.Game
	cfg            *config.Config
	ready          bool
	width          int
	height         int
	animFrame      int // animation frame counter for pause menu
	welcomeAnimState *ui.WelcomeAnimationState // welcome screen animation state
}

func initialModel(cfg *config.Config, g *game.Game) model {
	return model{
		game:           g,
		cfg:            cfg,
		ready:          false,
		welcomeAnimState: &ui.WelcomeAnimationState{},
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
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
		// Increment animation frame
		m.animFrame++

		// Update welcome animation if on welcome screen
		if !m.ready {
			ui.UpdateWelcomeAnimation(m.welcomeAnimState)
		}

		// Always return tick command to keep animation running
		return m, tickCmd()
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Waiting to start
	if !m.ready {
		switch msg.String() {
		case "up", "k":
			// Move selection up
			m.welcomeAnimState.SelectedOption = (m.welcomeAnimState.SelectedOption - 1 + 2) % 2
			return m, nil
		case "down", "j":
			// Move selection down
			m.welcomeAnimState.SelectedOption = (m.welcomeAnimState.SelectedOption + 1) % 2
			return m, nil
		case "enter":
			// Confirm selection
			if m.welcomeAnimState.SelectedOption == 0 {
				// Start selected
				if err := m.game.Start(m.cfg.WordCount); err != nil {
					return m, tea.Quit
				}
				m.ready = true
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
			// Letter keys
			runes := []rune(msg.String())
			if len(runes) == 1 {
				r := runes[0]
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					if r >= 'A' && r <= 'Z' {
						r = r + 32
					}
					m.game.AddChar(r)
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
		// Pause menu
		switch msg.String() {
		case "up", "k":
			m.game.MovePauseMenu(-1)
		case "down", "j":
			m.game.MovePauseMenu(1)
		case "enter":
			m.game.ConfirmPauseMenu()
			// If quit was selected, check status
			if m.game.Status == game.StatusFinished {
				return m, nil
			}
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	} else if m.game.Status == game.StatusFinished {
		// Results page, press any key to exit
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return ui.RenderWelcome(m.welcomeAnimState)
	}

	if m.game.Status == game.StatusRunning {
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
			TotalKeystrokes:   m.game.Stats.TotalKeystrokes,
			ValidKeystrokes:   m.game.Stats.ValidKeystrokes,
			CorrectChars:      m.game.Stats.CorrectChars,
			WordsCompleted:    m.game.Stats.WordsCompleted,
			TotalLetters:      m.game.Stats.TotalLetters,
			ElapsedSeconds:    m.game.Stats.GetElapsedSeconds(),
			LettersPerSecond:  m.game.Stats.GetLettersPerSecond(),
			WordsPerSecond:    m.game.Stats.GetWordsPerSecond(),
			AccuracyPercent:   m.game.Stats.GetAccuracyPercent(),
		}

		return ui.RenderGame(wordInfos, highlighted, m.game.InputBuffer, stats, len(activeWords))
	} else if m.game.Status == game.StatusPaused {
		// Pass stats and animation frame to pause menu
		activeWords := m.game.GetActiveWords()
		stats := ui.GameStats{
			TotalKeystrokes:   m.game.Stats.TotalKeystrokes,
			ValidKeystrokes:   m.game.Stats.ValidKeystrokes,
			CorrectChars:      m.game.Stats.CorrectChars,
			WordsCompleted:    m.game.Stats.WordsCompleted,
			TotalLetters:      m.game.Stats.TotalLetters,
			ElapsedSeconds:    m.game.Stats.GetElapsedSeconds(),
			LettersPerSecond:  m.game.Stats.GetLettersPerSecond(),
			WordsPerSecond:    m.game.Stats.GetWordsPerSecond(),
			AccuracyPercent:   m.game.Stats.GetAccuracyPercent(),
		}
		return ui.RenderPauseMenu(m.game.PauseMenuIndex, stats, len(activeWords), m.animFrame)
	} else if m.game.Status == game.StatusFinished {
		stats := ui.GameStats{
			TotalKeystrokes:   m.game.Stats.TotalKeystrokes,
			ValidKeystrokes:   m.game.Stats.ValidKeystrokes,
			CorrectChars:      m.game.Stats.CorrectChars,
			WordsCompleted:    m.game.Stats.WordsCompleted,
			TotalLetters:      m.game.Stats.TotalLetters,
			ElapsedSeconds:    m.game.Stats.GetElapsedSeconds(),
			LettersPerSecond:  m.game.Stats.GetLettersPerSecond(),
			WordsPerSecond:    m.game.Stats.GetWordsPerSecond(),
			AccuracyPercent:   m.game.Stats.GetAccuracyPercent(),
		}

		return ui.RenderResults(stats, m.game.Aborted)
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
