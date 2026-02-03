package game

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/word-killer/word-killer/pkg/stats"
)

// GameStatus game status
type GameStatus int

const (
	StatusIdle GameStatus = iota
	StatusRunning
	StatusPaused
	StatusFinished
)

// Word represents a word in the game
type Word struct {
	Text         string
	Completed    bool
	CompletedAt  time.Time // when the word was completed (for animation)
}

// Game core game logic
type Game struct {
	Status           GameStatus
	Words            []Word
	InputBuffer      string
	Stats            *stats.Statistics
	PauseMenuIndex   int // pause menu selected index (0=resume, 1=quit)
	Aborted          bool
	wordPool         []string
	usedWords        map[string]bool
	rng              *rand.Rand
}

// New creates a new game instance
func New() *Game {
	return &Game{
		Status:         StatusIdle,
		Stats:          stats.New(),
		PauseMenuIndex: 0,
		usedWords:      make(map[string]bool),
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// LoadWordDict loads word dictionary
func (g *Game) LoadWordDict(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open word dictionary: %w", err)
	}
	defer file.Close()

	g.wordPool = make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && isValidWord(word) {
			g.wordPool = append(g.wordPool, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read word dictionary: %w", err)
	}

	if len(g.wordPool) == 0 {
		return fmt.Errorf("word dictionary is empty")
	}

	return nil
}

// Start starts the game
func (g *Game) Start(wordCount int) error {
	if len(g.wordPool) == 0 {
		return fmt.Errorf("word dictionary not loaded")
	}

	// Reset game state
	g.Status = StatusRunning
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()
	g.usedWords = make(map[string]bool)

	// Generate game words
	g.Words = g.generateWords(wordCount)

	return nil
}

// generateWords 生成游戏单词
func (g *Game) generateWords(count int) []Word {
	// 如果词库数量不足，使用全部
	if count <= 0 || count > len(g.wordPool) {
		count = len(g.wordPool)
	}

	words := make([]Word, 0, count)
	availableWords := make([]string, len(g.wordPool))
	copy(availableWords, g.wordPool)

	for i := 0; i < count && len(availableWords) > 0; i++ {
		// 随机选择一个单词
		idx := g.rng.Intn(len(availableWords))
		word := availableWords[idx]

		words = append(words, Word{Text: word, Completed: false})
		g.usedWords[word] = true

		// 移除已选单词
		availableWords = append(availableWords[:idx], availableWords[idx+1:]...)
	}

	return words
}

// AddChar 添加字符到输入缓冲区
func (g *Game) AddChar(ch rune) {
	if g.Status != StatusRunning {
		return
	}

	g.InputBuffer += string(ch)
	g.Stats.AddKeystroke()

	// 检查是否匹配
	if g.hasMatch() {
		g.Stats.AddValidKeystroke()
		g.Stats.AddCorrectChar()
	}
}

// Backspace 删除最后一个字符
func (g *Game) Backspace() {
	if g.Status != StatusRunning {
		return
	}

	if len(g.InputBuffer) > 0 {
		g.InputBuffer = g.InputBuffer[:len(g.InputBuffer)-1]
		g.Stats.AddKeystroke()
	}
}

// TryEliminate tries to eliminate a word
func (g *Game) TryEliminate() {
	if g.Status != StatusRunning {
		return
	}

	g.Stats.AddKeystroke()

	if g.InputBuffer == "" {
		return
	}

	// Find completely matched word
	for i := range g.Words {
		if !g.Words[i].Completed && g.Words[i].Text == g.InputBuffer {
			// Eliminate word - this Enter key should be counted as correct
			g.Stats.AddCorrectChar()
			g.Words[i].Completed = true
			g.Words[i].CompletedAt = time.Now() // record completion time for animation
			g.Stats.AddCompletedWord(len(g.Words[i].Text))
			g.InputBuffer = ""

			// Check if all completed
			if g.isAllCompleted() {
				g.finish(false)
			}
			return
		}
	}
}

// Pause 暂停游戏
func (g *Game) Pause() {
	if g.Status == StatusRunning {
		g.Status = StatusPaused
		g.PauseMenuIndex = 0
		g.Stats.Pause()
	}
}

// Resume 恢复游戏
func (g *Game) Resume() {
	if g.Status == StatusPaused {
		g.Status = StatusRunning
		g.Stats.Resume()
	}
}

// MovePauseMenu 移动暂停菜单选项
func (g *Game) MovePauseMenu(delta int) {
	if g.Status != StatusPaused {
		return
	}

	g.PauseMenuIndex += delta
	if g.PauseMenuIndex < 0 {
		g.PauseMenuIndex = 0
	} else if g.PauseMenuIndex > 1 {
		g.PauseMenuIndex = 1
	}
}

// ConfirmPauseMenu 确认暂停菜单选择
func (g *Game) ConfirmPauseMenu() {
	if g.Status != StatusPaused {
		return
	}

	if g.PauseMenuIndex == 0 {
		// 继续游戏
		g.Resume()
	} else {
		// 结束游戏
		g.finish(true)
	}
}

// Abort 中止游戏
func (g *Game) Abort() {
	if g.Status == StatusRunning {
		g.finish(true)
	}
}

// finish 结束游戏
func (g *Game) finish(aborted bool) {
	g.Status = StatusFinished
	g.Aborted = aborted
	g.Stats.Finish()
}

// GetAllWords returns all words (including completed ones)
func (g *Game) GetAllWords() []Word {
	return g.Words
}

// GetActiveWords gets uncompleted words (for stats display)
func (g *Game) GetActiveWords() []string {
	words := make([]string, 0)
	for _, w := range g.Words {
		if !w.Completed {
			words = append(words, w.Text)
		}
	}
	return words
}

// GetMatchedIndices gets matched word indices (only for active words)
func (g *Game) GetMatchedIndices() []int {
	if g.InputBuffer == "" {
		return nil
	}

	indices := make([]int, 0)
	for i, w := range g.Words {
		if w.Completed {
			continue
		}

		if strings.HasPrefix(w.Text, g.InputBuffer) {
			indices = append(indices, i)
		}
	}

	return indices
}

// hasMatch 检查是否有匹配
func (g *Game) hasMatch() bool {
	if g.InputBuffer == "" {
		return false
	}

	for _, w := range g.Words {
		if !w.Completed && strings.HasPrefix(w.Text, g.InputBuffer) {
			return true
		}
	}

	return false
}

// isAllCompleted 检查是否全部完成
func (g *Game) isAllCompleted() bool {
	for _, w := range g.Words {
		if !w.Completed {
			return false
		}
	}
	return true
}

// isValidWord 验证单词格式
func isValidWord(word string) bool {
	if len(word) == 0 {
		return false
	}

	for _, ch := range word {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			return false
		}
	}

	return true
}
