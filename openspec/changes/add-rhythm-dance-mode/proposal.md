# Change: Add Rhythm Dance Mode

## Why
用户需要一种更具挑战性和娱乐性的打字练习方式。节奏舞蹈模式结合了打字练习和节奏游戏元素（类似 QQ 炫舞），通过节奏判定和视觉反馈增加游戏趣味性，同时保持打字练习的核心价值。

这种模式不仅能提升打字速度，还能训练玩家的节奏感和时机把握能力，提供与经典模式完全不同的游戏体验。

## What Changes
- **新增游戏模式**: 节奏舞蹈模式（Rhythm Dance Mode）
- **节奏判定系统**: 基于指针位置的 Perfect/Nice/OK/Miss 四档判定
- **动态难度**: 随游戏进行指针摆动速度逐渐加快
- **舞蹈动画**: ASCII 小人根据判定结果展示不同动作反馈
- **视觉特效**: 不同判定触发不同的长条区域特效（边框闪烁、文字浮现等）
- **得分系统**: Perfect(5分) > Nice(3分) > OK(1分) > Miss(0分)
- **限时挑战**: 固定时长（如 60 秒）内完成尽可能多的单词并获取高分

这是全新功能，不涉及破坏性变更。

## Impact
- **新增能力规范**:
  - `rhythm-timing` - 节奏判定和指针控制
  - `dance-animation` - 舞蹈小人动画系统
  - `visual-effects` - 节奏特效渲染

- **受影响规范**:
  - `game-modes` - 添加 ModeRhythmDance 枚举和初始化逻辑
  - `input-handling` - 添加节奏模式下的空格键和回车键判定逻辑
  - `ui-rendering` - 添加节奏模式专用界面渲染
  - `statistics` - 添加节奏模式统计（Perfect/Nice/OK/Miss 计数、总分等）

- **受影响代码**:
  - `pkg/game/game.go` - 添加 ModeRhythmDance 常量和 StartRhythmDanceMode 方法
  - `pkg/game/rhythm_dance.go` - 新增文件，实现节奏判定和指针控制逻辑
  - `pkg/ui/rhythm_dance.go` - 新增文件，实现节奏模式专用渲染
  - `cmd/word-killer/main.go` - 添加模式选择选项和启动逻辑

- **数据依赖**:
  - 使用现有的单词词库（short/medium/long）
  - 添加配置项：`rhythm_dance_duration`（限时时长，默认 60 秒）
  - 添加配置项：`rhythm_dance_word_count`（初始单词数，默认 1 个在屏幕上）
