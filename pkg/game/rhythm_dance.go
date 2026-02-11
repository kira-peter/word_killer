package game

import (
	"fmt"
	"math"
	"time"
)

// RhythmDanceState 节奏舞蹈模式的状态
type RhythmDanceState struct {
	// 指针位置和移动
	PointerPosition  float64 // 指针当前位置 [0.0, 1.0]
	PointerDirection int     // 摆动方向: 1=右, -1=左
	PointerSpeed     float64 // 摆动速度（每帧移动距离）

	// 黄金分割点
	GoldenRatio float64 // 黄金分割点位置（约 0.618）

	// 统计数据
	CompletedWords int // 已完成单词数
	TotalScore     int // 总分

	// 判定计数
	PerfectCount int
	NiceCount    int
	OKCount      int
	MissCount    int

	// 倒计时
	Duration  time.Duration // 总时长
	StartTime time.Time     // 开始时间

	// 当前单词
	CurrentWord string // 当前显示的单词

	// 最大连击数
	CurrentCombo int
	MaxCombo     int

	// 最近判定（用于显示特效）
	LastJudgment     string    // "Perfect", "Nice", "OK", "Miss"
	LastJudgmentTime time.Time // 上次判定时间

	// 舞蹈动画状态
	DanceAnimState *DanceAnimationState
}

// StartRhythmDanceMode 启动节奏舞蹈模式
func (g *Game) StartRhythmDanceMode(duration int) error {
	// 检查词库是否加载
	if len(g.shortPool) == 0 && len(g.mediumPool) == 0 && len(g.longPool) == 0 {
		return fmt.Errorf("word dictionaries not loaded")
	}

	// 重置游戏状态
	g.Status = StatusRunning
	g.Mode = ModeRhythmDance
	g.InputBuffer = ""
	g.Aborted = false
	g.Stats.Reset()
	g.Stats.Start()

	// 初始化节奏舞蹈状态
	g.RhythmDanceState = &RhythmDanceState{
		PointerPosition:  0.5,                                   // 从中间开始
		PointerDirection: 1,                                     // 向右
		PointerSpeed:     0.01,                                  // 初始速度（将从配置读取）
		GoldenRatio:      0.618,                                 // 黄金分割点
		CompletedWords:   0,
		TotalScore:       0,
		PerfectCount:     0,
		NiceCount:        0,
		OKCount:          0,
		MissCount:        0,
		Duration:         time.Duration(duration) * time.Second,
		StartTime:        time.Now(),
		CurrentCombo:     0,
		MaxCombo:         0,
		DanceAnimState:   NewDanceAnimationState(), // 初始化动画状态
	}

	// 随机选择第一个单词
	g.RhythmDanceState.CurrentWord = g.pickRandomWord()

	return nil
}

// UpdateRhythmPointer 更新指针位置
func (g *Game) UpdateRhythmPointer() {
	if g.RhythmDanceState == nil {
		return
	}

	state := g.RhythmDanceState

	// 更新位置
	state.PointerPosition += state.PointerSpeed * float64(state.PointerDirection)

	// 检查边界并反转方向
	if state.PointerPosition >= 1.0 {
		state.PointerPosition = 1.0
		state.PointerDirection = -1
	} else if state.PointerPosition <= 0.0 {
		state.PointerPosition = 0.0
		state.PointerDirection = 1
	}
}

// JudgeRhythmTiming 判定节奏时机
// 返回判定等级 ("Perfect", "Nice", "OK", "Miss") 和得分
func (g *Game) JudgeRhythmTiming() (string, int) {
	if g.RhythmDanceState == nil {
		return "Miss", 0
	}

	state := g.RhythmDanceState

	// 计算距离黄金分割点的距离
	distance := math.Abs(state.PointerPosition - state.GoldenRatio)

	var judgment string
	var score int

	// 根据距离判定等级
	if distance <= 0.05 {
		judgment = "Perfect"
		score = 5
		state.PerfectCount++
		state.CurrentCombo++
	} else if distance <= 0.15 {
		judgment = "Nice"
		score = 3
		state.NiceCount++
		state.CurrentCombo++
	} else if distance <= 0.30 {
		judgment = "OK"
		score = 1
		state.OKCount++
		state.CurrentCombo = 0 // OK 重置连击
	} else {
		judgment = "Miss"
		score = 0
		state.MissCount++
		state.CurrentCombo = 0 // Miss 重置连击
	}

	// 更新最大连击
	if state.CurrentCombo > state.MaxCombo {
		state.MaxCombo = state.CurrentCombo
	}

	// 更新总分
	state.TotalScore += score

	// 记录最近判定（用于特效显示）
	state.LastJudgment = judgment
	state.LastJudgmentTime = time.Now()

	return judgment, score
}

// CompleteRhythmWord 完成当前单词并切换到下一个
func (g *Game) CompleteRhythmWord() {
	if g.RhythmDanceState == nil {
		return
	}

	state := g.RhythmDanceState

	// 增加完成计数
	state.CompletedWords++

	// 增加速度（每完成1个单词增加0.001）
	state.PointerSpeed += 0.001

	// 清空输入缓冲区
	g.InputBuffer = ""

	// 随机选择下一个单词
	state.CurrentWord = g.pickRandomWord()
}

// CheckRhythmTimeout 检查倒计时是否结束
func (g *Game) CheckRhythmTimeout() {
	if g.RhythmDanceState == nil {
		return
	}

	elapsed := time.Since(g.RhythmDanceState.StartTime)
	if elapsed >= g.RhythmDanceState.Duration {
		g.finish(false) // 时间到，游戏结束
	}
}

// GetRhythmRemainingTime 获取剩余时间（秒）
func (g *Game) GetRhythmRemainingTime() int {
	if g.RhythmDanceState == nil {
		return 0
	}

	elapsed := time.Since(g.RhythmDanceState.StartTime)
	remaining := g.RhythmDanceState.Duration - elapsed

	if remaining < 0 {
		return 0
	}

	return int(remaining.Seconds())
}

// pickRandomWord 从词库中随机选择一个单词
func (g *Game) pickRandomWord() string {
	// 构建可用单词池
	allWords := make([]string, 0)
	allWords = append(allWords, g.shortPool...)
	allWords = append(allWords, g.mediumPool...)
	allWords = append(allWords, g.longPool...)

	if len(allWords) == 0 {
		return "word" // 默认单词
	}

	// 随机选择
	idx := g.rng.Intn(len(allWords))
	return allWords[idx]
}

// TryRhythmJudgment 尝试进行节奏判定（按空格键触发）
func (g *Game) TryRhythmJudgment() {
	if g.Mode != ModeRhythmDance || g.RhythmDanceState == nil {
		return
	}

	// 检查单词是否完全正确
	if g.InputBuffer != g.RhythmDanceState.CurrentWord {
		// 单词不正确，判定为 Miss
		g.RhythmDanceState.MissCount++
		g.RhythmDanceState.CurrentCombo = 0
		g.RhythmDanceState.LastJudgment = "Miss"
		g.RhythmDanceState.LastJudgmentTime = time.Now()
		g.InputBuffer = "" // 清空输入，重新输入
		return
	}

	// 单词正确，执行节奏判定
	judgment, score := g.JudgeRhythmTiming()

	// 触发对应的舞蹈动画
	g.TriggerJudgmentAnimation(judgment)

	// 记录统计
	g.Stats.AddCompletedWord(len(g.RhythmDanceState.CurrentWord))

	// 完成单词并切换到下一个
	g.CompleteRhythmWord()

	// 可以在这里添加舞蹈动画触发逻辑
	_ = judgment
	_ = score
}
