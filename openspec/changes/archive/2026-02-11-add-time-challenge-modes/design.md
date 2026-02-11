# Design: Time Challenge Game Modes

## Context

Word Killer 当前有两个游戏模式：
- **经典模式 (Classic)**: 屏幕中随机产生单词，用户输入匹配并消除所有单词
- **句子模式 (Sentence)**: 用户逐字符输入完整句子，支持标点符号和空格

这两个模式的共同特征是：**没有时间限制，用户可以按照自己的节奏完成**。

用户提出需求：添加三个**时间挑战类**模式，引入时间压力和限制，提供不同类型的挑战体验：
1. **倒计时模式** - 固定时间内尽可能多消除单词
2. **极速模式** - 固定单词数，追求最快完成时间
3. **节奏大师** - 每个单词限时，逐步提升难度

这些模式需要在现有架构基础上实现，复用统计系统、UI组件、暂停功能等。

## Goals / Non-Goals

### Goals
- 添加三个时间挑战模式，每个模式有独特的游戏机制和时间管理策略
- 完全复用现有的统计系统（`pkg/stats/statistics.go`）
- 复用现有的 UI 组件和样式系统
- 复用现有的暂停/恢复机制
- 为极速模式提供最佳记录持久化存储
- 保持与现有模式的一致性（菜单、暂停、重启等）

### Non-Goals
- 不修改现有的经典模式和句子模式
- 不添加新的外部依赖库
- 不重构现有的代码结构
- 第一版不提供在线排行榜或云端记录同步
- 第一版不提供自定义时间参数（使用固定值）

## Decisions

### 决策 1: 时间检查机制
**选择**: 双重时间检查 - 按键时检查 + 100ms tick 定时检查

**原因**:
- **按键时检查**: 在 `AddChar` 和 `TryEliminate` 中检查，确保用户操作时立即响应超时
- **定时 tick 检查**: 在主循环的 100ms tick 中调用 `CheckTimeouts`，处理用户无操作时的超时（如倒计时耗尽）

**替代方案考虑**:
- ❌ 仅依赖 tick 检查：响应延迟最高100ms，体验不够实时
- ❌ 仅依赖按键检查：用户无操作时无法检测超时（如倒计时归零）
- ✅ 双重检查：兼顾实时性和无操作场景，代码冗余度可接受

### 决策 2: 统计系统复用策略
**选择**: 完全复用现有的 `stats.Statistics`，不做任何修改

**原因**:
- 现有统计系统已经支持：总按键数、有效按键数、正确字符数、单词数、字母速度、单词速度、准确率
- 统计系统已内置暂停/恢复的时间处理（`Pause()`、`Resume()` 方法）
- 三个新模式的统计需求与现有指标完全兼容

**验证**:
- 倒计时模式：需要"消除单词数"和"字母速度" → 已有 `WordsCompleted` 和 `GetLettersPerSecond()`
- 极速模式：需要"总耗时（毫秒）"和"单词速度" → 已有 `GetElapsedSeconds()` 和 `GetWordsPerSecond()`
- 节奏大师：需要"连击数"和"字母速度" → 连击数用 `Game.ConsecutiveSuccesses` 存储，速度已有

### 决策 3: 模式特定字段的存储位置
**选择**: 直接在 `Game` 结构体中添加模式特定字段

**新增字段**:
```go
// 倒计时模式
CountdownDuration  time.Duration
CountdownStartTime time.Time

// 极速模式
SpeedRunTargetWords int
SpeedRunStartTime   time.Time

// 节奏大师
CurrentWordStartTime time.Time
WordTimeLimit        time.Duration
ConsecutiveSuccesses int
DifficultyLevel      int
```

**原因**:
- Go 不支持继承或接口字段，无法用多态优雅地处理
- 使用结构体组合会增加代码复杂度，访问路径变长（`g.countdown.StartTime` vs `g.CountdownStartTime`）
- 字段数量可控（8个），对结构体大小影响不大
- 模式判断已经用 `g.Mode` 枚举，字段冗余可接受

**替代方案考虑**:
- ❌ 结构体组合：`type Countdown struct {...}; Game.Countdown *Countdown` - 增加空指针检查，代码更复杂
- ❌ 接口抽象：`type ModeHandler interface {...}` - Go 接口无字段，需要额外的状态管理
- ✅ 直接字段：简单直接，字段数量可控

### 决策 4: 动态单词生成策略
**选择**: 批量生成 + 阈值触发补充

**倒计时模式和节奏大师**:
- 初始生成 30/50 个单词
- 当剩余单词 < 10 时，自动生成 20 个新词
- 复用现有的 `generateWordsFromMultiPools` 方法

**极速模式**:
- 固定 25 个单词，不再生成

**原因**:
- 批量生成减少频繁调用的性能开销
- 阈值触发确保屏幕上始终有足够的单词可选
- 复用现有的多难度词库混合逻辑

### 决策 5: 极速模式记录存储格式
**选择**: JSON 文件本地存储（`speedrun_record.json`）

**数据结构**:
```go
type speedRunRecord struct {
    BestTime float64 `json:"best_time"` // 单位：秒（含小数）
}
```

**原因**:
- 简单轻量，无需数据库依赖
- JSON 便于读取和调试
- 只存储最佳时间，数据量极小
- 后续可扩展为多难度记录或用户配置

**替代方案考虑**:
- ❌ 二进制格式（gob）：不便于调试和人工查看
- ❌ SQLite：过度设计，仅存储一个浮点数
- ✅ JSON：简单、可读、可扩展

### 决策 6: 节奏大师难度递增算法
**选择**: 线性递减时间限制，带最小值保护

**参数**:
- 初始时间限制：2.0 秒/词
- 每 10 词递减：0.1 秒
- 最小时间限制：0.5 秒

**公式**:
```go
newLimit := 2.0 - float64(g.DifficultyLevel) * 0.1
if newLimit < 0.5 {
    newLimit = 0.5
}
```

**原因**:
- 线性递减简单易懂，玩家可预测难度增长
- 最小值 0.5 秒避免过于极端的挑战（人类反应时间约 200-300ms）
- 每 10 词递增给玩家适应时间

**替代方案考虑**:
- ❌ 指数递减：难度曲线陡峭，玩家容易挫败
- ❌ 无最小值限制：会出现物理上不可能完成的挑战
- ✅ 线性 + 最小值：平衡挑战性和可玩性

## Risks / Trade-offs

### 风险 1: 时间计算精度
**风险**: `time.Since()` 的精度在不同平台可能不同，Windows 上可能只有 1-15ms 精度

**影响**: 极速模式的毫秒级计时可能不够精确

**缓解措施**:
- 使用 Go 的 `time.Now()` 和 `time.Since()`，Go runtime 会自动选择平台最佳时钟源
- 极速模式的记录比较是相对值，精度一致即可
- 后续可考虑使用 `time.Tick` 或 `time.NewTicker` 优化

### 风险 2: 暂停时的时间处理
**风险**: 暂停后恢复，时间计算可能出错（如倒计时继续流逝）

**影响**: 用户体验差，可能导致不公平的挑战

**缓解措施**:
- **统计时间**: 已由 `stats.Statistics` 的 `Pause()`/`Resume()` 正确处理
- **模式特定时间**: 需要在暂停时记录快照，恢复时调整
  - 倒计时：记录剩余时间，恢复时重新计算 `CountdownStartTime`
  - 节奏大师：记录当前单词剩余时间，恢复时重新计算 `CurrentWordStartTime`

**实现**: 在 `Game.Pause()` 和 `Game.Resume()` 中添加模式特定逻辑（待验证是否已实现）

### 风险 3: 100ms tick 的性能影响
**风险**: 每 100ms 调用 `CheckTimeouts()` 可能影响性能

**影响**: CPU 占用略微增加

**缓解措施**:
- `CheckTimeouts()` 只做简单的时间比较（ns 级操作）
- 仅在 `StatusRunning` 时执行，暂停/结束时跳过
- 100ms 间隔在现代硬件上几乎无感知（60fps = 16ms 一帧）

### 风险 4: 模式切换时的状态清理
**风险**: 从一个模式切换到另一个模式时，状态字段可能残留

**影响**: 可能导致逻辑错误（如节奏大师的 `ConsecutiveSuccesses` 影响倒计时模式）

**缓解措施**:
- 每个 `Start*Mode()` 函数中重置所有共享字段（`InputBuffer`, `Aborted`, `Stats.Reset()`）
- 模式特定字段在各自的启动函数中初始化
- 使用 `g.Mode` 枚举严格隔离逻辑分支

## Migration Plan

**不涉及数据迁移**，这是纯新增功能。

**发布流程**:
1. 合并代码到主分支
2. 用户更新后，新模式自动出现在模式选择菜单中
3. 极速模式的 `speedrun_record.json` 在首次完成游戏时自动创建

**回滚计划**:
- 代码变更集中在三个文件，回滚 commit 即可
- 删除 `speedrun_record.json` 无副作用

## Open Questions

### Q1: 暂停时的时间处理是否需要额外实现？
**现状**: 需要检查 `Game.Pause()` 和 `Game.Resume()` 是否已处理模式特定时间

**待验证**:
- 是否需要在暂停时保存倒计时的剩余时间？
- 是否需要在恢复时调整 `CountdownStartTime`？

**决策**: 在实现阶段检查现有代码，如果需要则添加

### Q2: 节奏大师的难度参数是否需要可配置？
**现状**: 使用硬编码的 2.0s 初始、0.1s 递减、0.5s 最小值

**考虑**: 后续版本可以考虑在配置文件中允许调整

**决策**: 第一版使用固定值，收集用户反馈后再决定是否开放配置

### Q3: 是否需要为倒计时模式添加时长配置？
**现状**: 固定 60 秒

**考虑**: 可以添加 30s/60s/90s 的选项

**决策**: 第一版固定 60 秒，后续可扩展难度选择

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/word-killer/main.go                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Bubble Tea Model                                         │   │
│  │ - 100ms Tick → CheckTimeouts()                           │   │
│  │ - 模式选择（5个选项）                                     │   │
│  │ - View 渲染分支（根据 g.Mode）                           │   │
│  │ - 记录存储（loadSpeedRunBestTime/saveSpeedRunBestTime） │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              ↓↑                                  │
└──────────────────────────────┼───────────────────────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ↓                      ↓                      ↓
┌────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│ pkg/game/      │   │ pkg/ui/          │   │ pkg/stats/       │
│ game.go        │   │ styles.go        │   │ statistics.go    │
│                │   │                  │   │                  │
│ Game 结构体:   │   │ 渲染函数:        │   │ Statistics:      │
│ - Mode 枚举    │   │ - RenderCountdown│   │ - 复用现有       │
│ - 8个新字段    │   │ - RenderSpeedRun │   │ - 无需修改       │
│                │   │ - RenderRhythm   │   │                  │
│ 启动函数:      │   │ - renderRhythm   │   │ 方法:            │
│ - StartCountdown   WordArea          │   │ - Pause()        │
│ - StartSpeedRun│   │                  │   │ - Resume()       │
│ - StartRhythm  │   │ 模式菜单:        │   │ - GetElapsed     │
│                │   │ - 5个选项        │   │ - GetLettersPer  │
│ 逻辑修改:      │   │                  │   │   Second()       │
│ - AddChar      │   └──────────────────┘   │                  │
│ - TryEliminate │                           └──────────────────┘
│ - CheckTimeouts│
└────────────────┘

数据流:
1. 用户按键 → AddChar → 时间检查 + 匹配检测
2. 回车键 → TryEliminate → 消除单词 + 模式特定逻辑（生成新词/难度递增/结束判定）
3. 100ms tick → CheckTimeouts → 倒计时/节奏模式超时检测
4. 游戏结束 → 极速模式保存最佳记录
```

## Implementation Sequence

1. **阶段 1**: 核心游戏逻辑（`game.go`）
   - 添加枚举、字段、启动函数
   - 修改 AddChar/TryEliminate
   - 添加 CheckTimeouts

2. **阶段 2**: UI 渲染（`styles.go`）
   - 实现三个渲染函数
   - 更新模式选择菜单

3. **阶段 3**: 主程序集成（`main.go`）
   - 更新 model
   - 扩展事件处理
   - 添加记录存储

4. **阶段 4**: 测试验证
   - 功能测试、边界测试、UI测试
