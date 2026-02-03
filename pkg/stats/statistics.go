package stats

import (
	"time"
)

// Statistics 游戏统计数据
type Statistics struct {
	// 计数器
	TotalKeystrokes int // 总敲击数
	ValidKeystrokes int // 有效敲击数
	CorrectChars    int // 正确字符数
	WordsCompleted  int // 完成单词数
	TotalLetters    int // 总字母数

	// 时间跟踪
	StartTime           time.Time     // 开始时间
	EndTime             time.Time     // 结束时间
	PauseStartTime      time.Time     // 暂停开始时间
	TotalPausedDuration time.Duration // 累计暂停时长
	isPaused            bool          // 是否正在暂停
}

// New 创建统计对象
func New() *Statistics {
	return &Statistics{}
}

// Start 开始计时
func (s *Statistics) Start() {
	s.StartTime = time.Now()
}

// Pause 暂停计时
func (s *Statistics) Pause() {
	if !s.isPaused {
		s.PauseStartTime = time.Now()
		s.isPaused = true
	}
}

// Resume 恢复计时
func (s *Statistics) Resume() {
	if s.isPaused {
		pauseDuration := time.Since(s.PauseStartTime)
		s.TotalPausedDuration += pauseDuration
		s.isPaused = false
	}
}

// Finish 结束计时
func (s *Statistics) Finish() {
	// 如果还在暂停中，先恢复
	if s.isPaused {
		s.Resume()
	}
	s.EndTime = time.Now()
}

// AddKeystroke 增加总敲击数
func (s *Statistics) AddKeystroke() {
	s.TotalKeystrokes++
}

// AddValidKeystroke 增加有效敲击数
func (s *Statistics) AddValidKeystroke() {
	s.ValidKeystrokes++
}

// AddCorrectChar 增加正确字符数
func (s *Statistics) AddCorrectChar() {
	s.CorrectChars++
}

// AddCompletedWord 增加完成单词数
func (s *Statistics) AddCompletedWord(wordLength int) {
	s.WordsCompleted++
	s.TotalLetters += wordLength
}

// GetElapsedSeconds 获取有效耗时（秒）
func (s *Statistics) GetElapsedSeconds() float64 {
	var elapsed time.Duration

	if s.EndTime.IsZero() {
		// 游戏还在进行中
		elapsed = time.Since(s.StartTime)
		// 如果正在暂停，需要减去当前暂停时长
		if s.isPaused {
			currentPauseDuration := time.Since(s.PauseStartTime)
			elapsed -= (s.TotalPausedDuration + currentPauseDuration)
		} else {
			elapsed -= s.TotalPausedDuration
		}
	} else {
		// 游戏已结束
		elapsed = s.EndTime.Sub(s.StartTime) - s.TotalPausedDuration
	}

	if elapsed < 0 {
		elapsed = 0
	}

	return elapsed.Seconds()
}

// GetLettersPerSecond 获取字母速度
func (s *Statistics) GetLettersPerSecond() float64 {
	elapsed := s.GetElapsedSeconds()
	if elapsed < 0.1 {
		return 0.0
	}
	return float64(s.TotalLetters) / elapsed
}

// GetWordsPerSecond 获取单词速度
func (s *Statistics) GetWordsPerSecond() float64 {
	elapsed := s.GetElapsedSeconds()
	if elapsed < 0.1 {
		return 0.0
	}
	return float64(s.WordsCompleted) / elapsed
}

// GetAccuracyPercent 获取准确率
func (s *Statistics) GetAccuracyPercent() float64 {
	if s.TotalKeystrokes == 0 {
		return 0.0
	}
	return float64(s.CorrectChars) / float64(s.TotalKeystrokes) * 100.0
}

// Reset 重置所有统计数据
func (s *Statistics) Reset() {
	s.TotalKeystrokes = 0
	s.ValidKeystrokes = 0
	s.CorrectChars = 0
	s.WordsCompleted = 0
	s.TotalLetters = 0
	s.StartTime = time.Time{}
	s.EndTime = time.Time{}
	s.PauseStartTime = time.Time{}
	s.TotalPausedDuration = 0
	s.isPaused = false
}
