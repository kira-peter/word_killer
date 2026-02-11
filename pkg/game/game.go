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
	ModeCountdown           // 倒计时模式 - 60秒限时
	ModeSpeedRun            // 极速模式 - 固定25词
	ModeRhythmMaster        // 节奏大师 - 每词限时
	ModeUnderwaterCountdown // 水下倒计时模式
	ModeRhythmDance         // 节奏舞蹈模式 - 打字+节奏判定
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

	// 倒计时模式专属字段
	CountdownDuration  time.Duration // 总倒计时时长（如60秒）
	CountdownStartTime time.Time     // 倒计时开始时间

	// 极速模式专属字段
	SpeedRunTargetWords int       // 固定单词数（如25个）
	SpeedRunStartTime   time.Time // 用于毫秒级计时

	// 节奏大师模式专属字段
	CurrentWordStartTime time.Time     // 当前单词开始时间
	WordTimeLimit        time.Duration // 每个单词的时间限制（初始2秒）
	ConsecutiveSuccesses int           // 连击计数
	DifficultyLevel      int           // 当前难度等级（每10词递增）

	// 节奏大师配置参数
	RhythmInitialTimeLimit float64 // 初始时间限制（秒）
	RhythmMinTimeLimit     float64 // 最小时间限制（秒）
	RhythmDifficultyStep   float64 // 难度递增步长（秒）
	RhythmWordsPerLevel    int     // 每级所需单词数

	// Underwater countdown mode fields
	UnderwaterState       *UnderwaterState
	CountdownDurationSecs int // 从配置读取，默认60秒

	// Rhythm Dance mode fields
	RhythmDanceState *RhythmDanceState
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

// StartCountdownMode 启动倒计时模式
func (g *Game) StartCountdownMode(duration time.Duration) error {
	// 检查词库是否加载
	if len(g.shortPool) == 0 && len(g.mediumPool) == 0 && len(g.longPool) == 0 {
		return fmt.Errorf("word dictionaries not loaded")
	}

	// 重置游戏状态
	g.Status = StatusRunning
	g.Mode = ModeCountdown
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()
	g.usedWords = make(map[string]bool)

	// 初始化倒计时字段
	g.CountdownDuration = duration
	g.CountdownStartTime = time.Now()

	// 生成初始单词（30个，后续动态补充）
	g.Words = g.generateWordsFromMultiPools(30)

	return nil
}

// StartSpeedRunMode 启动极速模式
func (g *Game) StartSpeedRunMode(targetWords int) error {
	// 检查词库是否加载
	if len(g.shortPool) == 0 && len(g.mediumPool) == 0 && len(g.longPool) == 0 {
		return fmt.Errorf("word dictionaries not loaded")
	}

	// 重置游戏状态
	g.Status = StatusRunning
	g.Mode = ModeSpeedRun
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()
	g.usedWords = make(map[string]bool)

	// 初始化极速模式字段
	g.SpeedRunTargetWords = targetWords
	g.SpeedRunStartTime = time.Now()

	// 生成固定数量的单词（25个）
	g.Words = g.generateWordsFromMultiPools(targetWords)

	return nil
}

// StartRhythmMasterMode 启动节奏大师模式
func (g *Game) StartRhythmMasterMode() error {
	// 检查词库是否加载
	if len(g.shortPool) == 0 && len(g.mediumPool) == 0 && len(g.longPool) == 0 {
		return fmt.Errorf("word dictionaries not loaded")
	}

	// 重置游戏状态
	g.Status = StatusRunning
	g.Mode = ModeRhythmMaster
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()
	g.usedWords = make(map[string]bool)

	// 初始化节奏大师字段
	g.WordTimeLimit = time.Duration(g.RhythmInitialTimeLimit * float64(time.Second)) // 使用配置的初始时间限制
	g.ConsecutiveSuccesses = 0
	g.DifficultyLevel = 0

	// 生成大量单词（50个起步）
	g.Words = g.generateWordsFromMultiPools(50)

	// 标记第一个单词的开始时间
	if len(g.Words) > 0 {
		g.CurrentWordStartTime = time.Now()
	}

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

	// 新增：模式特定的时间检查
	switch g.Mode {
	case ModeCountdown:
		// 检查倒计时是否耗尽
		elapsed := time.Since(g.CountdownStartTime)
		if elapsed >= g.CountdownDuration {
			g.finish(false) // 时间到，游戏结束
			return
		}
	case ModeRhythmMaster:
		// 检查当前单词是否超时
		wordElapsed := time.Since(g.CurrentWordStartTime)
		if wordElapsed >= g.WordTimeLimit {
			g.finish(false) // 超时，节奏失败
			return
		}
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
	} else if g.Mode == ModeRhythmDance {
		// Rhythm Dance mode: 只接受字母，检查是否匹配当前单词
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			g.InputBuffer += string(ch)
			g.Stats.AddKeystroke()

			// 检查是否匹配当前单词
			if g.RhythmDanceState != nil {
				currentWord := g.RhythmDanceState.CurrentWord
				if len(g.InputBuffer) <= len(currentWord) &&
					strings.HasPrefix(currentWord, g.InputBuffer) {
					g.Stats.AddValidKeystroke()
					g.Stats.AddCorrectChar()
				}
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

			// 新增：模式特定的完成后处理
			switch g.Mode {
			case ModeCountdown:
				// 动态补充单词
				remainingWords := 0
				for _, w := range g.Words {
					if !w.Completed {
						remainingWords++
					}
				}
				if remainingWords < 10 {
					// 剩余单词少于10个，生成20个新词
					newWords := g.generateWordsFromMultiPools(20)
					g.Words = append(g.Words, newWords...)
				}

			case ModeSpeedRun:
				// 检查是否全部完成
				if g.isAllCompleted() {
					g.finish(false) // 全部完成，游戏结束
					return
				}

			case ModeRhythmMaster:
				// 增加连击数
				g.ConsecutiveSuccesses++

				// 根据配置的每级单词数增加难度
				if g.ConsecutiveSuccesses > 0 && g.ConsecutiveSuccesses%g.RhythmWordsPerLevel == 0 {
					g.DifficultyLevel++
					// 减少时间限制，使用配置的步长和最小值
					newLimit := g.RhythmInitialTimeLimit - float64(g.DifficultyLevel)*g.RhythmDifficultyStep
					if newLimit < g.RhythmMinTimeLimit {
						newLimit = g.RhythmMinTimeLimit
					}
					g.WordTimeLimit = time.Duration(newLimit * float64(time.Second))
				}

				// 为下一个单词启动计时器
				for _, w := range g.Words {
					if !w.Completed {
						g.CurrentWordStartTime = time.Now()
						break
					}
				}

				// 动态补充单词
				remainingWords := 0
				for _, w := range g.Words {
					if !w.Completed {
						remainingWords++
					}
				}
				if remainingWords < 10 {
					newWords := g.generateWordsFromMultiPools(20)
					g.Words = append(g.Words, newWords...)
				}
			}

			// Check if all completed
			if g.Mode == ModeClassic && g.isAllCompleted() {
				g.finish(false)
			}
			return
		}
	}
}

// CheckTimeouts 检查模式特定的超时条件
// 应在主循环的每个tick（100ms）调用
func (g *Game) CheckTimeouts() {
	if g.Status != StatusRunning {
		return
	}

	switch g.Mode {
	case ModeCountdown:
		// 检查倒计时
		elapsed := time.Since(g.CountdownStartTime)
		if elapsed >= g.CountdownDuration {
			g.finish(false) // 时间到
		}

	case ModeRhythmMaster:
		// 检查当前活动单词是否超时
		wordElapsed := time.Since(g.CurrentWordStartTime)
		if wordElapsed >= g.WordTimeLimit {
			g.finish(false) // 节奏失败
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
	g.CountdownDurationSecs = durationSeconds

	return nil
}

// UpdateCountdown 更新倒计时
func (g *Game) UpdateCountdown() {
	if g.Mode != ModeUnderwaterCountdown || g.UnderwaterState == nil {
		return
	}

	elapsed := time.Since(g.UnderwaterState.CountdownStart).Seconds()
	remaining := float64(g.CountdownDurationSecs) - elapsed

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
	remaining := float64(g.CountdownDurationSecs) - elapsed
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
