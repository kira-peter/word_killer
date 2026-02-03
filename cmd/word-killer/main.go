package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/word-killer/word-killer/pkg/config"
	"github.com/word-killer/word-killer/pkg/game"
	"github.com/word-killer/word-killer/pkg/ui"
)

// model 是 Bubble Tea 的模型
type model struct {
	game   *game.Game
	cfg    *config.Config
	ready  bool
	width  int
	height int
}

func initialModel(cfg *config.Config, g *game.Game) model {
	return model{
		game:  g,
		cfg:   cfg,
		ready: false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 等待开始
	if !m.ready {
		switch msg.String() {
		case "enter":
			if err := m.game.Start(m.cfg.WordCount); err != nil {
				return m, tea.Quit
			}
			m.ready = true
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}

	// 游戏进行中
	if m.game.Status == game.StatusRunning {
		switch msg.String() {
		case "esc":
			m.game.Pause()
		case "enter":
			m.game.TryEliminate()
		case "backspace":
			m.game.Backspace()
		default:
			// 字母键
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

		// 检查游戏状态
		if m.game.Status == game.StatusFinished {
			// 游戏结束，等待用户按键后退出
			return m, nil
		}

		return m, nil
	} else if m.game.Status == game.StatusPaused {
		// 暂停菜单
		switch msg.String() {
		case "up", "k":
			m.game.MovePauseMenu(-1)
		case "down", "j":
			m.game.MovePauseMenu(1)
		case "enter":
			m.game.ConfirmPauseMenu()
			// 如果选择了退出，检查状态
			if m.game.Status == game.StatusFinished {
				return m, nil
			}
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	} else if m.game.Status == game.StatusFinished {
		// 结果页面，按任意键退出
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return ui.RenderWelcome()
	}

	if m.game.Status == game.StatusRunning {
		words := m.game.GetActiveWords()
		highlighted := m.game.GetMatchedIndices()

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

		return ui.RenderGame(words, highlighted, m.game.InputBuffer, stats)
	} else if m.game.Status == game.StatusPaused {
		return ui.RenderPauseMenu(m.game.PauseMenuIndex)
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

	// Load word dictionary
	if err := g.LoadWordDict(cfg.WordDictPath); err != nil {
		fmt.Printf("Failed to load word dictionary: %v\n", err)
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
