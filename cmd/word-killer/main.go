package main

import (
	"encoding/json"
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
	selectedMode     int  // 0=经典, 1=句子, 2=倒计时, 3=极速, 4=节奏大师, 5=水下倒计时, 6=节奏舞蹈
	width            int
	height           int
	animFrame        int                       // animation frame counter for pause menu
	welcomeAnimState *ui.WelcomeAnimationState // welcome screen animation state
	// 极速模式专用
	speedRunBestTime float64 // 最佳时间（秒），从文件加载
	tickCount        int     // tick 计数器，用于控制游戏逻辑更新频率
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

			// Update rhythm dance animations if in rhythm dance mode
			if m.ready && m.game.Mode == game.ModeRhythmDance && m.game.Status == game.StatusRunning {
				m.game.UpdateRhythmPointer()
				m.game.UpdateDanceAnimation()
			}
		}

		// 新增：检查时间模式的超时条件
		if m.ready && m.game.Status == game.StatusRunning {
			m.game.CheckTimeouts()
		}

		// Always return tick command to keep animation running at 30 FPS
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
			// Move selection up（现在有7个模式）
			m.selectedMode = (m.selectedMode - 1 + 7) % 7
			return m, nil
		case "down", "j":
			// Move selection down
			m.selectedMode = (m.selectedMode + 1) % 7
			return m, nil
		case "enter":
			// Start game with selected mode
			var err error
			switch m.selectedMode {
			case 0:
				// 经典模式
				err = m.game.Start(m.cfg.WordCount)
			case 1:
				// 句子模式
				err = m.game.StartSentenceMode()
			case 2:
				// 倒计时模式 - 从配置读取时长
				err = m.game.StartCountdownMode(time.Duration(m.cfg.CountdownDuration) * time.Second)
			case 3:
				// 极速模式 - 从配置读取单词数
				err = m.game.StartSpeedRunMode(m.cfg.SpeedRunWordCount)
				// 加载最佳时间记录
				m.speedRunBestTime = loadSpeedRunBestTime()
			case 4:
				// 节奏大师模式
				err = m.game.StartRhythmMasterMode()
			case 5:
				// 水下倒计时模式
				err = m.game.StartUnderwaterCountdown(m.cfg.CountdownDuration)
			case 6:
				// 节奏舞蹈模式
				err = m.game.StartRhythmDanceMode(m.cfg.RhythmDanceDuration)
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
			// 节奏舞蹈模式回车触发判定
			// 其他模式触发确认
			if m.game.Mode == game.ModeRhythmDance {
				m.game.TryRhythmJudgment()
			} else {
				m.game.TryEliminate()
			}
		case "backspace":
			m.game.Backspace()
		default:
			// Handle input based on game mode
			runes := []rune(msg.String())
			if len(runes) == 1 {
				r := runes[0]
				// 节奏舞蹈模式：空格键触发判定
				if m.game.Mode == game.ModeRhythmDance && r == ' ' {
					m.game.TryRhythmJudgment()
				} else if m.game.Mode == game.ModeSentence {
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
				switch m.game.Mode {
				case game.ModeSentence:
					m.game.StartSentenceMode()
				case game.ModeCountdown:
					m.game.StartCountdownMode(time.Duration(m.cfg.CountdownDuration) * time.Second)
				case game.ModeSpeedRun:
					m.game.StartSpeedRunMode(m.cfg.SpeedRunWordCount)
				case game.ModeRhythmMaster:
					m.game.StartRhythmMasterMode()
				case game.ModeUnderwaterCountdown:
					m.game.StartUnderwaterCountdown(m.cfg.CountdownDuration)
				case game.ModeRhythmDance:
					m.game.StartRhythmDanceMode(m.cfg.RhythmDanceDuration)
				default:
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
				switch m.game.Mode {
				case game.ModeSentence:
					m.game.StartSentenceMode()
				case game.ModeCountdown:
					m.game.StartCountdownMode(time.Duration(m.cfg.CountdownDuration) * time.Second)
				case game.ModeSpeedRun:
					m.game.StartSpeedRunMode(m.cfg.SpeedRunWordCount)
				case game.ModeRhythmMaster:
					m.game.StartRhythmMasterMode()
				case game.ModeUnderwaterCountdown:
					m.game.StartUnderwaterCountdown(m.cfg.CountdownDuration)
				case game.ModeRhythmDance:
					m.game.StartRhythmDanceMode(m.cfg.RhythmDanceDuration)
				default:
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
		switch m.game.Mode {
		case game.ModeUnderwaterCountdown:
			// Underwater countdown mode rendering
			return ui.RenderUnderwaterGame(m.game)
		case game.ModeSentence:
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

		case game.ModeCountdown:
			// 倒计时模式渲染
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
			elapsed := time.Since(m.game.CountdownStartTime)
			remaining := m.game.CountdownDuration - elapsed
			remainingSec := remaining.Seconds()
			if remainingSec < 0 {
				remainingSec = 0
			}
			return ui.RenderCountdownGame(wordInfos, highlighted, m.game.InputBuffer, stats,
				remainingSec, m.game.CountdownDuration.Seconds())

		case game.ModeSpeedRun:
			// 极速模式渲染
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
			currentTime := time.Since(m.game.SpeedRunStartTime).Seconds()
			return ui.RenderSpeedRunGame(wordInfos, highlighted, m.game.InputBuffer, stats,
				currentTime, m.speedRunBestTime)

		case game.ModeRhythmMaster:
			// 节奏大师模式渲染
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
			wordElapsed := time.Since(m.game.CurrentWordStartTime)
			wordRemaining := m.game.WordTimeLimit - wordElapsed
			wordRemainingSec := wordRemaining.Seconds()
			if wordRemainingSec < 0 {
				wordRemainingSec = 0
			}
			return ui.RenderRhythmMasterGame(wordInfos, highlighted, m.game.InputBuffer, stats,
				wordRemainingSec, m.game.WordTimeLimit.Seconds(),
				m.game.ConsecutiveSuccesses, m.game.DifficultyLevel)

		case game.ModeRhythmDance:
			// 节奏舞蹈模式渲染
			if m.game.RhythmDanceState == nil {
				return "Rhythm Dance mode not initialized"
			}
			state := m.game.RhythmDanceState

			// 获取舞蹈帧
			danceFrame := m.game.GetCurrentDanceFrame()

			// 构建节奏条信息
			rhythmBar := ui.RhythmBarInfo{
				PointerPosition: state.PointerPosition,
				GoldenRatio:     state.GoldenRatio,
			}

			// 构建统计信息
			stats := ui.RhythmDanceStats{
				RemainingTime:   m.game.GetRhythmRemainingTime(),
				CompletedWords:  state.CompletedWords,
				TotalScore:      state.TotalScore,
				CurrentCombo:    state.CurrentCombo,
				JudgmentHistory: state.JudgmentHistory,
			}

			// 构建判定特效信息
			judgmentEffect := ui.JudgmentEffectInfo{
				LastJudgment:     state.LastJudgment,
				LastJudgmentTime: state.LastJudgmentTime,
			}

			return ui.RenderRhythmDanceGame(
				danceFrame,
				state.CurrentWord,
				m.game.InputBuffer,
				rhythmBar,
				stats,
				judgmentEffect,
			)

		default:
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
		}
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

		// 如果是极速模式且未中止，检查是否创造新记录
		if m.game.Mode == game.ModeSpeedRun && !m.game.Aborted {
			completionTime := m.game.Stats.GetElapsedSeconds()
			if m.speedRunBestTime == 0 || completionTime < m.speedRunBestTime {
				// 新记录！
				saveSpeedRunBestTime(completionTime)
				m.speedRunBestTime = completionTime
			}
		}

		// 如果是节奏舞蹈模式，渲染专用结果界面
		if m.game.Mode == game.ModeRhythmDance && m.game.RhythmDanceState != nil {
			state := m.game.RhythmDanceState
			rhythmStats := ui.RhythmDanceStats{
				RemainingTime:  0, // 已结束
				CompletedWords: state.CompletedWords,
				TotalScore:     state.TotalScore,
				CurrentCombo:   state.CurrentCombo,
				MaxCombo:       state.MaxCombo,
				PerfectCount:   state.PerfectCount,
				NiceCount:      state.NiceCount,
				OKCount:        state.OKCount,
				MissCount:      state.MissCount,
			}
			return ui.RenderRhythmDanceResults(rhythmStats, m.game.ResultsMenuIndex, m.animFrame)
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

	// 设置节奏大师模式的配置参数
	g.RhythmInitialTimeLimit = cfg.RhythmInitialTimeLimit
	g.RhythmMinTimeLimit = cfg.RhythmMinTimeLimit
	g.RhythmDifficultyStep = cfg.RhythmDifficultyStep
	g.RhythmWordsPerLevel = cfg.RhythmWordsPerLevel

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

// speedRunRecord 存储极速模式的最佳时间
type speedRunRecord struct {
	BestTime float64 `json:"best_time"` // 单位：秒
}

// loadSpeedRunBestTime 从文件加载最佳时间
func loadSpeedRunBestTime() float64 {
	const recordFile = "speedrun_record.json"

	file, err := os.Open(recordFile)
	if err != nil {
		return 0 // 还没有记录文件
	}
	defer file.Close()

	var record speedRunRecord
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&record); err != nil {
		return 0
	}

	return record.BestTime
}

// saveSpeedRunBestTime 保存新的最佳时间到文件
func saveSpeedRunBestTime(newTime float64) error {
	const recordFile = "speedrun_record.json"

	record := speedRunRecord{
		BestTime: newTime,
	}

	file, err := os.Create(recordFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(&record)
}
