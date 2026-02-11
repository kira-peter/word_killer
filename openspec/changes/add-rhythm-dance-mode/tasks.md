# Implementation Tasks

## 1. 配置扩展
- [ ] 1.1 在 `pkg/config/config.go` 中添加 `RhythmDanceDuration` 字段（默认 60 秒）
- [ ] 1.2 在 `pkg/config/config.go` 中添加 `RhythmDanceInitialSpeed` 字段（默认 0.01）
- [ ] 1.3 更新 `config.json` 添加节奏舞蹈模式配置示例

## 2. 游戏核心逻辑扩展
- [ ] 2.1 在 `pkg/game/game.go` 中添加 `ModeRhythmDance` 枚举常量
- [ ] 2.2 创建 `pkg/game/rhythm_dance.go` 文件
- [ ] 2.3 定义 `RhythmDanceState` 结构体（指针位置、速度、方向、黄金点等）
- [ ] 2.4 在 `Game` 结构体中添加 `RhythmDanceState *RhythmDanceState` 字段
- [ ] 2.5 实现 `StartRhythmDanceMode(duration int)` 方法
- [ ] 2.6 实现指针更新逻辑 `UpdateRhythmPointer()`
- [ ] 2.7 实现节奏判定逻辑 `JudgeRhythmTiming() (string, int)` 返回判定等级和得分
- [ ] 2.8 实现加速逻辑（每完成 1 个单词增加速度 0.001）
- [ ] 2.9 在 `Game` 结构体中添加判定计数字段（PerfectCount, NiceCount, OKCount, MissCount, TotalScore）
- [ ] 2.10 实现倒计时逻辑和超时检测

## 3. 输入处理扩展
- [ ] 3.1 修改 `AddChar()` 方法，支持节奏舞蹈模式的字母输入
- [ ] 3.2 修改 `Backspace()` 方法，支持节奏舞蹈模式
- [ ] 3.3 实现 `TryRhythmJudgment()` 方法，处理空格键触发的节奏判定
- [ ] 3.4 在 `TryRhythmJudgment()` 中检查单词是否完全正确
- [ ] 3.5 判定完成后切换到下一个随机单词

## 4. 舞蹈动画系统
- [ ] 4.1 创建 `pkg/game/dance_animation.go` 文件
- [ ] 4.2 定义 `DanceAnimationState` 结构体（当前状态、帧计数、动画类型）
- [ ] 4.3 定义动画常量（AnimIdle, AnimPerfect, AnimNice, AnimOK, AnimMiss）
- [ ] 4.4 设计并实现 Idle 动画的 2 帧 ASCII 小人
- [ ] 4.5 设计并实现 Perfect 动画的 3 帧 ASCII 小人（跳跃）
- [ ] 4.6 设计并实现 Nice 动画的 2 帧 ASCII 小人（摆臂）
- [ ] 4.7 设计并实现 OK 动画的 2 帧 ASCII 小人（点头）
- [ ] 4.8 设计并实现 Miss 动画的 2 帧 ASCII 小人（摔倒）
- [ ] 4.9 实现 `UpdateDanceAnimation()` 方法管理动画帧切换
- [ ] 4.10 实现 `TriggerJudgmentAnimation(judgment string)` 方法触发特定动画
- [ ] 4.11 在 `Game` 结构体中添加 `DanceAnimState *DanceAnimationState` 字段

## 5. UI 渲染 - 节奏舞蹈模式
- [ ] 5.1 创建 `pkg/ui/rhythm_dance.go` 文件
- [ ] 5.2 实现 `RenderRhythmDanceGame()` 主渲染函数
- [ ] 5.3 实现 `renderDanceCharacter()` 渲染舞蹈小人（上部区域）
- [ ] 5.4 实现 `renderWordArea()` 渲染单词和输入（中间左侧）
- [ ] 5.5 实现 `renderRhythmBar()` 渲染节奏条框架（40 字符宽）
- [ ] 5.6 实现 `renderGoldenRatioMarker()` 渲染黄金分割点标记
- [ ] 5.7 实现 `renderPointer()` 渲染指针并根据距离设置颜色渐变
- [ ] 5.8 实现 `renderJudgmentEffect()` 渲染判定特效（边框闪烁、文字上浮）
- [ ] 5.9 实现 `renderRhythmStats()` 渲染统计信息（时间、分数、各档计数）
- [ ] 5.10 实现特效状态管理（EffectState 结构体，记录特效类型、剩余帧数）
- [ ] 5.11 实现 `RenderRhythmDanceResults()` 渲染最终结果页面

## 6. 主程序集成
- [ ] 6.1 在 `cmd/word-killer/main.go` 的 `selectedMode` 注释中添加节奏舞蹈模式（索引 6）
- [ ] 6.2 在模式选择界面增加节奏舞蹈模式的上下选择逻辑（0-6 共 7 个模式）
- [ ] 6.3 在 `handleKey()` 的 enter 分支中添加 case 6，调用 `game.StartRhythmDanceMode(cfg.RhythmDanceDuration)`
- [ ] 6.4 在 `Update()` 的 tick 处理中添加节奏舞蹈模式的更新逻辑:
  - 更新指针位置
  - 更新舞蹈动画
  - 更新倒计时
  - 更新特效状态
- [ ] 6.5 在 `View()` 中添加节奏舞蹈模式的渲染分支
- [ ] 6.6 在 `handleKey()` 的 Running 状态下，添加空格键处理（节奏舞蹈模式专用）

## 7. 测试和验证
- [ ] 7.1 手动测试节奏舞蹈模式启动流程
- [ ] 7.2 手动测试指针摆动和边界反转
- [ ] 7.3 手动测试 Perfect/Nice/OK/Miss 各档判定阈值
- [ ] 7.4 手动测试单词输入和颜色变化
- [ ] 7.5 手动测试空格键判定触发
- [ ] 7.6 手动测试舞蹈动画切换（Idle -> 判定动画 -> Idle）
- [ ] 7.7 手动测试判定特效渲染（边框闪烁、文字上浮）
- [ ] 7.8 手动测试加速逻辑（每个单词完成后速度增加）
- [ ] 7.9 手动测试倒计时和自动结束
- [ ] 7.10 手动测试最终结果页面显示
- [ ] 7.11 验证统计数据准确性（各档计数、总分、准确率）

## 8. 文档和完善
- [ ] 8.1 更新 README 说明节奏舞蹈模式的玩法
- [ ] 8.2 添加节奏舞蹈模式的配置说明
- [ ] 8.3 添加代码注释（特别是判定算法和动画逻辑）
- [ ] 8.4 性能优化（如需要，确保特效不影响帧率）
