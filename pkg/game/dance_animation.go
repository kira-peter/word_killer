package game

import "time"

// DanceAnimationType 舞蹈动画类型
type DanceAnimationType int

const (
	AnimIdle DanceAnimationType = iota
	AnimPerfect
	AnimNice
	AnimOK
	AnimMiss
)

// DanceAnimationState 舞蹈动画状态
type DanceAnimationState struct {
	CurrentAnimation DanceAnimationType // 当前动画类型
	FrameIndex       int                // 当前帧索引
	FrameCount       int                // 总帧数
	LastUpdate       time.Time          // 上次更新时间
	AnimationStart   time.Time          // 动画开始时间
}

// Idle 动画帧（2帧循环）
var idleFrames = []string{
	`  o
 /|\
 / \`,
	`  o
 \|/
 / \`,
}

// Perfect 动画帧（3帧播放一次）
var perfectFrames = []string{
	` \o/
  |
 / \`,
	`  o
 /|\
  ^`,
	` \o/
  |
 / \`,
}

// Nice 动画帧（2帧播放一次）
var niceFrames = []string{
	`  o
 /|
 / \`,
	`  o
  |\
 / \`,
}

// OK 动画帧（2帧播放一次）
var okFrames = []string{
	`  o
 /|\
 / \`,
	`  .
 /|\
 / \`,
}

// Miss 动画帧（2帧播放一次）
var missFrames = []string{
	` _o_
  |
 / \`,
	`  o
_/|\_
 / \`,
}

// NewDanceAnimationState 创建新的舞蹈动画状态
func NewDanceAnimationState() *DanceAnimationState {
	return &DanceAnimationState{
		CurrentAnimation: AnimIdle,
		FrameIndex:       0,
		FrameCount:       len(idleFrames),
		LastUpdate:       time.Now(),
		AnimationStart:   time.Now(),
	}
}

// UpdateDanceAnimation 更新舞蹈动画
func (g *Game) UpdateDanceAnimation() {
	if g.RhythmDanceState == nil || g.RhythmDanceState.DanceAnimState == nil {
		return
	}

	state := g.RhythmDanceState.DanceAnimState
	now := time.Now()

	// 每200ms切换一帧
	if now.Sub(state.LastUpdate) < 200*time.Millisecond {
		return
	}

	state.LastUpdate = now

	// 检查判定动画是否播放完毕（1秒后回到 Idle）
	if state.CurrentAnimation != AnimIdle {
		if now.Sub(state.AnimationStart) >= 1*time.Second {
			// 回到 Idle 动画
			state.CurrentAnimation = AnimIdle
			state.FrameIndex = 0
			state.FrameCount = len(idleFrames)
			return
		}
	}

	// 更新帧索引
	state.FrameIndex++
	if state.FrameIndex >= state.FrameCount {
		// Idle 动画循环，其他动画停在最后一帧
		if state.CurrentAnimation == AnimIdle {
			state.FrameIndex = 0
		} else {
			state.FrameIndex = state.FrameCount - 1
		}
	}
}

// TriggerJudgmentAnimation 触发判定动画
func (g *Game) TriggerJudgmentAnimation(judgment string) {
	if g.RhythmDanceState == nil {
		return
	}

	// 如果还没有动画状态，创建一个
	if g.RhythmDanceState.DanceAnimState == nil {
		g.RhythmDanceState.DanceAnimState = NewDanceAnimationState()
	}

	state := g.RhythmDanceState.DanceAnimState

	// 根据判定类型设置动画
	switch judgment {
	case "Perfect":
		state.CurrentAnimation = AnimPerfect
		state.FrameCount = len(perfectFrames)
	case "Nice":
		state.CurrentAnimation = AnimNice
		state.FrameCount = len(niceFrames)
	case "OK":
		state.CurrentAnimation = AnimOK
		state.FrameCount = len(okFrames)
	case "Miss":
		state.CurrentAnimation = AnimMiss
		state.FrameCount = len(missFrames)
	default:
		return
	}

	state.FrameIndex = 0
	state.AnimationStart = time.Now()
}

// GetCurrentDanceFrame 获取当前舞蹈帧
func (g *Game) GetCurrentDanceFrame() string {
	if g.RhythmDanceState == nil || g.RhythmDanceState.DanceAnimState == nil {
		return idleFrames[0]
	}

	state := g.RhythmDanceState.DanceAnimState

	switch state.CurrentAnimation {
	case AnimIdle:
		if state.FrameIndex < len(idleFrames) {
			return idleFrames[state.FrameIndex]
		}
		return idleFrames[0]
	case AnimPerfect:
		if state.FrameIndex < len(perfectFrames) {
			return perfectFrames[state.FrameIndex]
		}
		return perfectFrames[len(perfectFrames)-1]
	case AnimNice:
		if state.FrameIndex < len(niceFrames) {
			return niceFrames[state.FrameIndex]
		}
		return niceFrames[len(niceFrames)-1]
	case AnimOK:
		if state.FrameIndex < len(okFrames) {
			return okFrames[state.FrameIndex]
		}
		return okFrames[len(okFrames)-1]
	case AnimMiss:
		if state.FrameIndex < len(missFrames) {
			return missFrames[state.FrameIndex]
		}
		return missFrames[len(missFrames)-1]
	default:
		return idleFrames[0]
	}
}
