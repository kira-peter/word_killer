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

// GameMode game mode
type GameMode int

const (
	ModeClassic GameMode = iota
	ModeSentence
	ModeUnderwaterCountdown
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
	Mode             GameMode
	Words            []Word
	InputBuffer      string
	Stats            *stats.Statistics
	PauseMenuIndex   int // pause menu selected index (0=resume, 1=restart, 2=select mode, 3=main menu)
	ResultsMenuIndex int // results menu selected index (0=restart, 1=select mode, 2=main menu)
	Aborted          bool
	shortPool        []string
	mediumPool       []string
	longPool         []string
	usedWords        map[string]bool
	rng              *rand.Rand
	// Normalized difficulty ratios (0-1 range)
	shortRatio       float64
	mediumRatio      float64
	longRatio        float64
	// Sentence mode fields
	TargetSentence   string
	sentences        []string
	// Underwater countdown mode fields
	UnderwaterState  *UnderwaterState
	CountdownDuration int  // 从配置读取，默认60秒
}

// New creates a new game instance
func New() *Game {
	return &Game{
		Status:           StatusIdle,
		Stats:            stats.New(),
		PauseMenuIndex:   0,
		ResultsMenuIndex: 0,
		usedWords:        make(map[string]bool),
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// LoadWordDictionaries loads multiple difficulty-based word dictionaries
func (g *Game) LoadWordDictionaries(shortPath, mediumPath, longPath string, shortRatio, mediumRatio, longRatio float64) error {
	var hasError bool
	errorMsg := ""

	// Load short dictionary
	if shortPath != "" && shortRatio > 0 {
		if err := g.loadDictToPool(shortPath, &g.shortPool); err != nil {
			errorMsg += fmt.Sprintf("short dictionary: %v; ", err)
			hasError = true
		}
	}

	// Load medium dictionary
	if mediumPath != "" && mediumRatio > 0 {
		if err := g.loadDictToPool(mediumPath, &g.mediumPool); err != nil {
			errorMsg += fmt.Sprintf("medium dictionary: %v; ", err)
			hasError = true
		}
	}

	// Load long dictionary
	if longPath != "" && longRatio > 0 {
		if err := g.loadDictToPool(longPath, &g.longPool); err != nil {
			errorMsg += fmt.Sprintf("long dictionary: %v; ", err)
			hasError = true
		}
	}

	// Check if at least one pool is loaded
	if len(g.shortPool) == 0 && len(g.mediumPool) == 0 && len(g.longPool) == 0 {
		if hasError {
			return fmt.Errorf("failed to load any word dictionary: %s", errorMsg)
		}
		return fmt.Errorf("all word dictionaries are empty")
	}

	// Store normalized ratios
	g.shortRatio = shortRatio
	g.mediumRatio = mediumRatio
	g.longRatio = longRatio

	return nil
}

// loadDictToPool loads a dictionary file into a word pool
func (g *Game) loadDictToPool(path string, pool *[]string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer file.Close()

	*pool = make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && isValidWord(word) {
			*pool = append(*pool, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	if len(*pool) == 0 {
		return fmt.Errorf("dictionary is empty")
	}

	return nil
}

// LoadSentences loads sentences from a text file
func (g *Game) LoadSentences(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open sentences file: %w", err)
	}
	defer file.Close()

	g.sentences = make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			g.sentences = append(g.sentences, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read sentences file: %w", err)
	}

	if len(g.sentences) == 0 {
		return fmt.Errorf("sentences file is empty")
	}

	return nil
}

// Start starts the game
func (g *Game) Start(wordCount int) error {
	// Check if dictionaries are loaded
	if len(g.shortPool) == 0 && len(g.mediumPool) == 0 && len(g.longPool) == 0 {
		return fmt.Errorf("word dictionaries not loaded")
	}

	// Reset game state
	g.Status = StatusRunning
	g.Mode = ModeClassic
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()
	g.usedWords = make(map[string]bool)

	// Generate game words from multi-pools
	g.Words = g.generateWordsFromMultiPools(wordCount)

	return nil
}

// StartSentenceMode starts the game in sentence mode
func (g *Game) StartSentenceMode() error {
	// Check if sentences are loaded
	if len(g.sentences) == 0 {
		return fmt.Errorf("sentences not loaded")
	}

	// Reset game state
	g.Status = StatusRunning
	g.Mode = ModeSentence
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()

	// Randomly select a sentence
	idx := g.rng.Intn(len(g.sentences))
	g.TargetSentence = g.sentences[idx]

	return nil
}

// generateWordsFromMultiPools generates words from multiple difficulty pools based on ratios
func (g *Game) generateWordsFromMultiPools(count int) []Word {
	if count <= 0 {
		count = 20 // default count
	}

	// Calculate target counts for each difficulty
	shortCount := int(float64(count) * g.shortRatio)
	mediumCount := int(float64(count) * g.mediumRatio)
	longCount := int(float64(count) * g.longRatio)

	// Adjust for rounding errors to ensure total equals count
	total := shortCount + mediumCount + longCount
	if total < count {
		// Add remaining to the pool with highest ratio
		diff := count - total
		if g.mediumRatio >= g.shortRatio && g.mediumRatio >= g.longRatio {
			mediumCount += diff
		} else if g.shortRatio >= g.longRatio {
			shortCount += diff
		} else {
			longCount += diff
		}
	}

	words := make([]Word, 0, count)

	// Select words from each pool
	words = append(words, g.selectWordsFromPool(g.shortPool, shortCount)...)
	words = append(words, g.selectWordsFromPool(g.mediumPool, mediumCount)...)
	words = append(words, g.selectWordsFromPool(g.longPool, longCount)...)

	// If we don't have enough words, try to fill from other pools
	if len(words) < count {
		needed := count - len(words)
		allPools := make([]string, 0)
		allPools = append(allPools, g.shortPool...)
		allPools = append(allPools, g.mediumPool...)
		allPools = append(allPools, g.longPool...)

		// Filter out already used words
		available := make([]string, 0)
		for _, w := range allPools {
			if !g.usedWords[w] {
				available = append(available, w)
			}
		}

		words = append(words, g.selectWordsFromPool(available, needed)...)
	}

	// Shuffle the words to mix difficulties
	for i := len(words) - 1; i > 0; i-- {
		j := g.rng.Intn(i + 1)
		words[i], words[j] = words[j], words[i]
	}

	return words
}

// selectWordsFromPool selects random words from a pool without repetition
func (g *Game) selectWordsFromPool(pool []string, count int) []Word {
	if count <= 0 || len(pool) == 0 {
		return nil
	}

	// Limit count to available words
	if count > len(pool) {
		count = len(pool)
	}

	// Create a copy of available words (excluding already used)
	available := make([]string, 0)
	for _, w := range pool {
		if !g.usedWords[w] {
			available = append(available, w)
		}
	}

	// Adjust count if not enough available words
	if count > len(available) {
		count = len(available)
	}

	words := make([]Word, 0, count)
	for i := 0; i < count && len(available) > 0; i++ {
		// Randomly select a word
		idx := g.rng.Intn(len(available))
		word := available[idx]

		words = append(words, Word{Text: word, Completed: false})
		g.usedWords[word] = true

		// Remove selected word from available list
		available = append(available[:idx], available[idx+1:]...)
	}

	return words
}

// AddChar 添加字符到输入缓冲区
func (g *Game) AddChar(ch rune) {
	if g.Status != StatusRunning {
		return
	}

	// Handle based on game mode
	if g.Mode == ModeSentence {
		// Sentence mode: accept all printable characters
		if ch >= 32 && ch <= 126 { // ASCII printable range
			g.InputBuffer += string(ch)
			g.Stats.AddKeystroke()

			// Check if the character matches the target at this position
			pos := len(g.InputBuffer) - 1
			if pos < len(g.TargetSentence) {
				if g.InputBuffer[pos] == g.TargetSentence[pos] {
					g.Stats.AddCorrectChar()
				}
			}

			// Check if sentence is completed
			if len(g.InputBuffer) == len(g.TargetSentence) {
				// Sentence completed, but don't finish until Enter is pressed
			}
		}
	} else {
		// Classic mode: only accept letters
		g.InputBuffer += string(ch)
		g.Stats.AddKeystroke()

		// 检查是否匹配
		if g.hasMatch() {
			g.Stats.AddValidKeystroke()
			g.Stats.AddCorrectChar()
		}
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

// TryEliminate tries to eliminate a word or finish sentence
func (g *Game) TryEliminate() {
	if g.Status != StatusRunning {
		return
	}

	if g.Mode == ModeSentence {
		// Sentence mode: finish if input length matches target
		// Don't count Enter key as a keystroke in sentence mode
		if len(g.InputBuffer) == len(g.TargetSentence) {
			g.finish(false)
		}
		// Otherwise ignore Enter key
		return
	}

	if g.Mode == ModeUnderwaterCountdown {
		// 海底模式：检查是否匹配任何小鱼
		if g.InputBuffer == "" {
			return
		}

		for i := range g.UnderwaterState.Fishes {
			fish := &g.UnderwaterState.Fishes[i]
			if !fish.Completed && fish.Word == g.InputBuffer {
				// 抓到小鱼！
				fish.Completed = true
				fish.CompletedAt = time.Now()
				fish.Glowing = true
				g.Stats.AddCompletedWord(len(fish.Word))
				g.Stats.AddCorrectChar() // Enter键计为正确
				g.InputBuffer = ""
				return
			}
		}
		return
	}

	g.Stats.AddKeystroke()

	// Classic mode: eliminate matching word
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
	} else if g.PauseMenuIndex > 3 {
		g.PauseMenuIndex = 3
	}
}

// MoveResultsMenu 移动结果菜单选项
func (g *Game) MoveResultsMenu(delta int) {
	if g.Status != StatusFinished {
		return
	}

	g.ResultsMenuIndex += delta
	if g.ResultsMenuIndex < 0 {
		g.ResultsMenuIndex = 0
	} else if g.ResultsMenuIndex > 2 {
		g.ResultsMenuIndex = 2
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
	g.ResultsMenuIndex = 0
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

// StartUnderwaterCountdown 启动海底倒计时模式
func (g *Game) StartUnderwaterCountdown(durationSeconds int) error {
	g.Status = StatusRunning
	g.Mode = ModeUnderwaterCountdown
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()

	// 初始化海底状态
	g.UnderwaterState = &UnderwaterState{
		CountdownStart:   time.Now(),
		BackgroundFrame:  0,
		SeaweedPositions: generateSeaweedPositions(),
		BubbleStreams:    generateBubbleStreams(),
	}

	// 生成10条小鱼
	g.UnderwaterState.Fishes = g.GenerateFishes(10)
	g.CountdownDuration = durationSeconds

	return nil
}

// UpdateCountdown 更新倒计时
func (g *Game) UpdateCountdown() {
	if g.Mode != ModeUnderwaterCountdown || g.UnderwaterState == nil {
		return
	}

	elapsed := time.Since(g.UnderwaterState.CountdownStart).Seconds()
	remaining := float64(g.CountdownDuration) - elapsed

	if remaining <= 0 {
		g.finish(false) // 时间用尽，游戏结束
	}
}

// GetRemainingTime 获取剩余时间（秒）
func (g *Game) GetRemainingTime() int {
	if g.UnderwaterState == nil {
		return 0
	}

	elapsed := time.Since(g.UnderwaterState.CountdownStart).Seconds()
	remaining := float64(g.CountdownDuration) - elapsed
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}

// GetAvailableWords 获取可用单词列表（用于海底模式生成小鱼）
func (g *Game) GetAvailableWords() []string {
	words := make([]string, 0)

	// 从各个单词池收集单词
	words = append(words, g.shortPool...)
	words = append(words, g.mediumPool...)
	words = append(words, g.longPool...)

	return words
}
