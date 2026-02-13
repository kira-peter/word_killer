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
	speedIncrement   float64 // 每完成一个单词的速度增量（未导出，仅内部使用）

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

	// 单词队列管理（多单词显示）
	WordQueue        []string // 固定长度5: [history-2, history-1, current, next+1, next+2]
	CurrentWordIndex int      // 永远是2（中间位置）
	// Deprecated: 使用 WordQueue[CurrentWordIndex] 替代
	CurrentWord string // 当前显示的单词

	// 最大连击数
	CurrentCombo int
	MaxCombo     int

	// 最近判定（用于显示特效）
	LastJudgment         string    // "Perfect", "Nice", "OK", "Miss"
	LastJudgmentTime     time.Time // 上次判定时间
	LastJudgmentPosition float64   // 上次判定的指针位置 [0.0, 1.0]（用于显示箭头）

	// 判定历史记录（按顺序记录每次判定结果）
	JudgmentHistory []string // 存储每次判定的结果："Perfect", "Nice", "OK", "Miss"

	// 舞蹈动画状态
	DanceAnimState *DanceAnimationState
}

// StartRhythmDanceMode 启动节奏舞蹈模式
func (g *Game) StartRhythmDanceMode(duration int, initialSpeed float64, speedIncrement float64) error {
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
		PointerPosition:  0.0,                                 // 从起点开始
		PointerDirection: 1,                                   // 向右
		PointerSpeed:     initialSpeed,                        // 使用传入的初始速度
		GoldenRatio:      0.618,                               // 黄金分割点
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
		JudgmentHistory:  []string{},               // 初始化判定历史
		DanceAnimState:   NewDanceAnimationState(), // 初始化动画状态
		WordQueue:        make([]string, 5),        // 初始化单词队列（固定长度5）
		CurrentWordIndex: 2,                        // 当前单词在中间位置
	}

	// 保存速度增量配置，用于CompleteRhythmWord
	g.RhythmDanceState.speedIncrement = speedIncrement

	// 初始化单词队列
	// 前2个位置（索引0-1）设为空字符串作为历史区占位符
	g.RhythmDanceState.WordQueue[0] = ""
	g.RhythmDanceState.WordQueue[1] = ""

	// 后3个位置（索引2-4）填充不重复的随机单词
	usedWords := make(map[string]bool)
	for i := 2; i < 5; i++ {
		word := g.pickRandomWord()
		// 确保单词不重复
		for usedWords[word] {
			word = g.pickRandomWord()
		}
		usedWords[word] = true
		g.RhythmDanceState.WordQueue[i] = word
	}

	// 保持向后兼容，设置 CurrentWord（已废弃，使用 WordQueue[2] 替代）
	g.RhythmDanceState.CurrentWord = g.RhythmDanceState.WordQueue[2]

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

	// 检查边界
	if state.PointerPosition >= 1.0 {
		state.PointerPosition = 0.0
	}
}

// JudgeRhythmTiming 判定节奏时机
// 返回判定等级 ("Perfect", "Nice", "OK", "Miss") 和得分
func (g *Game) JudgeRhythmTiming() (string, int) {
	if g.RhythmDanceState == nil {
		return "Miss", 0
	}

	state := g.RhythmDanceState

	// 节奏条宽度（与UI渲染保持一致）
	const barWidth = 35

	// 计算指针和黄金点的字符位置
	pointerPos := int(state.PointerPosition * float64(barWidth))
	goldenPos := int(state.GoldenRatio * float64(barWidth))

	// 计算字符距离（绝对距离）
	charDistance := int(math.Abs(float64(pointerPos - goldenPos)))

	var judgment string
	var score int

	// 根据字符距离判定等级
	// Pattern: --4433223344--
	//          Miss OK Nice Perfect Nice OK Miss
	if charDistance == 0 {
		// Perfect: 黄金点本身（1个字符）
		judgment = "Perfect"
		score = 5
		state.PerfectCount++
		state.CurrentCombo++
	} else if charDistance <= 2 {
		// Nice: 距离1-2个字符（左右各2个）
		judgment = "Nice"
		score = 3
		state.NiceCount++
		state.CurrentCombo++
	} else if charDistance <= 4 {
		// OK: 距离3-4个字符（左右各2个）
		judgment = "OK"
		score = 1
		state.OKCount++
		state.CurrentCombo = 0 // OK 重置连击
	} else {
		// Miss: 距离>4个字符
		judgment = "Miss"
		score = -1 // Miss 扣1分
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
	state.LastJudgmentPosition = state.PointerPosition // 记录判定时的指针位置

	// 判定后将指针重置到起点
	state.PointerPosition = 0.0
	state.PointerDirection = 1 // 重新从左向右移动

	// 添加到判定历史记录
	state.JudgmentHistory = append(state.JudgmentHistory, judgment)

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

	// 增加速度（使用保存的速度增量）
	state.PointerSpeed += state.speedIncrement

	// 清空输入缓冲区
	g.InputBuffer = ""

	// 单词队列上移：丢弃索引0，其他单词索引减1
	state.WordQueue = state.WordQueue[1:]

	// 生成新单词追加到末尾（索引4）
	newWord := g.generateUniqueWord(state.WordQueue)
	state.WordQueue = append(state.WordQueue, newWord)

	// 队列长度保持为5，当前单词始终在索引2
	// （移除1个+追加1个，自动满足）

	// 保持向后兼容，更新 CurrentWord（已废弃，使用 WordQueue[2] 替代）
	state.CurrentWord = state.WordQueue[2]
}

// CheckRhythmTimeout 检查倒计时是否结束以及Miss次数
func (g *Game) CheckRhythmTimeout() {
	if g.RhythmDanceState == nil {
		return
	}

	// 检查Miss次数，达到10次提前结束游戏
	if g.RhythmDanceState.MissCount >= 10 {
		g.finish(false) // Miss过多，游戏结束
		return
	}

	// 检查时间是否到
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

// generateUniqueWord 生成一个不在现有队列中的随机单词
func (g *Game) generateUniqueWord(existingWords []string) string {
	// 创建已存在单词的映射（排除空字符串）
	used := make(map[string]bool)
	for _, word := range existingWords {
		if word != "" {
			used[word] = true
		}
	}

	// 生成新单词直到找到不重复的
	maxAttempts := 100 // 防止无限循环
	for i := 0; i < maxAttempts; i++ {
		word := g.pickRandomWord()
		if !used[word] {
			return word
		}
	}

	// 如果尝试多次仍未找到，直接返回（词库足够大时不应发生）
	return g.pickRandomWord()
}

// TryRhythmJudgment 尝试进行节奏判定（按空格键触发）
func (g *Game) TryRhythmJudgment() {
	if g.Mode != ModeRhythmDance || g.RhythmDanceState == nil {
		return
	}

	// 获取当前单词（队列中间位置，索引2）
	currentWord := g.RhythmDanceState.WordQueue[g.RhythmDanceState.CurrentWordIndex]

	// 检查单词是否完全正确
	if g.InputBuffer != currentWord {
		// 单词不正确或不完整，判定为 Miss，扣1分
		g.RhythmDanceState.MissCount++
		g.RhythmDanceState.CurrentCombo = 0
		g.RhythmDanceState.TotalScore -= 1 // Miss 扣1分
		g.RhythmDanceState.LastJudgment = "Miss"
		g.RhythmDanceState.LastJudgmentTime = time.Now()

		// 添加到判定历史记录
		g.RhythmDanceState.JudgmentHistory = append(g.RhythmDanceState.JudgmentHistory, "Miss")

		// 触发Miss动画
		g.TriggerJudgmentAnimation("Miss")

		g.InputBuffer = "" // 清空输入，重新输入
		return
	}

	// 单词正确，执行节奏判定
	judgment, score := g.JudgeRhythmTiming()

	// 触发对应的舞蹈动画
	g.TriggerJudgmentAnimation(judgment)

	// 记录统计
	g.Stats.AddCompletedWord(len(currentWord))

	// 完成单词并切换到下一个
	g.CompleteRhythmWord()

	// 可以在这里添加舞蹈动画触发逻辑
	_ = judgment
	_ = score
}
