# ui-rendering Specification

## Purpose
TBD - created by archiving change add-classic-game-mode. Update Purpose after archive.
## Requirements
### Requirement: Screen Initialization
系统 SHALL 正确初始化和清理命令行屏幕环境。

#### Scenario: Terminal setup on game start
- **WHEN** 游戏启动时
- **THEN** 系统清空终端屏幕
- **AND** 隐藏光标
- **AND** 切换到原始模式（raw mode）

#### Scenario: Terminal cleanup on game exit
- **WHEN** 游戏退出时
- **THEN** 系统显示光标
- **AND** 恢复终端正常模式
- **AND** 清空或保留最终输出（可配置）

### Requirement: Word List Rendering
系统 SHALL 在屏幕上清晰渲染当前所有单词。

#### Scenario: Render active words
- **WHEN** 渲染游戏画面时
- **AND** 存在未消除的单词
- **THEN** 系统显示所有单词列表
- **AND** 每个单词占一行或分列显示

#### Scenario: Update word display on elimination
- **WHEN** 某个单词被消除后
- **THEN** 系统从显示列表移除该单词
- **AND** 重新渲染剩余单词

### Requirement: Input Highlighting
系统 SHALL 高亮显示当前匹配的单词，提供即时视觉反馈。

#### Scenario: Highlight matched word
- **WHEN** 用户输入匹配某个单词前缀
- **THEN** 系统高亮显示该单词（如使用不同颜色）
- **AND** 高亮匹配的字符部分

#### Scenario: Multiple word highlighting
- **WHEN** 用户输入匹配多个单词
- **THEN** 系统高亮所有匹配的单词

#### Scenario: Remove highlight on mismatch
- **WHEN** 用户输入不再匹配任何单词
- **THEN** 系统移除所有高亮
- **AND** 恢复单词正常显示

#### Scenario: Display current input
- **WHEN** 渲染游戏画面时
- **THEN** 系统在固定位置显示当前输入缓冲区内容
- **AND** 使用明显的视觉样式（如下划线或边框）

### Requirement: Statistics Display
系统 SHALL 实时显示游戏统计信息。

#### Scenario: Display real-time stats during game
- **WHEN** 游戏进行中
- **THEN** 系统在屏幕固定区域显示：
  - 剩余单词数量
  - 已完成单词数量
  - 当前用时
  - 实时速度（字母/秒）

#### Scenario: Display final results
- **WHEN** 游戏完成时
- **THEN** 系统显示完整统计报告：
  - 总敲击数
  - 有效敲击数
  - 正确字符数
  - 完成单词数
  - 总字母数
  - 总耗时
  - 平均字母速度（字母/秒）
  - 平均单词速度（单词/秒）
  - 准确率

### Requirement: Frame Rate Control
系统 SHALL 控制屏幕刷新频率，保证流畅体验。

#### Scenario: Limit refresh rate
- **WHEN** 游戏循环运行时
- **THEN** 系统控制渲染频率不超过 60fps
- **AND** 避免不必要的重绘

#### Scenario: Immediate render on input
- **WHEN** 用户输入导致视觉变化时
- **THEN** 系统立即触发重绘
- **AND** 显示更新后的状态

### Requirement: Cross-platform Compatibility
系统 SHALL 在主流平台上正确渲染。

#### Scenario: Render on Windows
- **WHEN** 在 Windows 终端运行
- **THEN** 系统正确显示所有元素和颜色

#### Scenario: Render on macOS/Linux
- **WHEN** 在 Unix-like 终端运行
- **THEN** 系统正确显示所有元素和颜色

#### Scenario: Fallback for unsupported terminals
- **WHEN** 检测到终端不支持彩色或高级特性时
- **THEN** 系统降级到基本文本渲染
- **AND** 保证核心功能可用

### Requirement: Pause Menu Rendering
系统 SHALL 在游戏暂停时渲染暂停菜单界面。

#### Scenario: Render pause menu
- **WHEN** 游戏状态为 Paused
- **THEN** 系统在屏幕中央或固定位置显示暂停菜单
- **AND** 显示标题"游戏已暂停"或"PAUSED"
- **AND** 列出菜单选项（继续、结束）

#### Scenario: Highlight selected menu item
- **WHEN** 渲染暂停菜单时
- **THEN** 系统高亮显示当前选中的菜单项
- **AND** 使用箭头、颜色或其他视觉提示标识选中状态

#### Scenario: Hide game content during pause
- **WHEN** 暂停菜单显示时
- **THEN** 系统隐藏或淡化游戏内容（可选）
- **AND** 暂停菜单清晰可见，不被遮挡

### Requirement: Welcome Title Animation
系统 SHALL 在欢迎界面为 "Word Killer" 标题提供持续循环的动画效果。

#### Scenario: Word 抖动特效
- **WHEN** 用户查看欢迎界面
- **THEN** "Word" 文字 SHALL 显示随机抖动效果，每个字符独立进行随机位置偏移
- **AND** 抖动 SHALL 持续循环播放

#### Scenario: Killer 红色呼吸脉冲
- **WHEN** 用户查看欢迎界面
- **THEN** "Killer" 文字 SHALL 显示红色呼吸脉冲渐变效果
- **AND** 颜色 SHALL 在暗红色到亮红色之间循环渐变
- **AND** 渐变 SHALL 持续循环播放

#### Scenario: 动画帧管理
- **WHEN** 欢迎界面显示时
- **THEN** 系统 SHALL 维护动画帧计数器
- **AND** 帧计数 SHALL 随时间持续递增
- **AND** 动画效果 SHALL 基于当前帧数计算

### Requirement: Mode Selection Screen
UI SHALL 渲染模式选择界面，允许玩家选择游戏模式。

#### Scenario: Display rhythm dance mode option
- **WHEN** 渲染模式选择界面
- **THEN** 系统应显示"节奏舞蹈模式 (Rhythm Dance Mode)"选项
- **AND** 该选项应排在其他模式之后
- **AND** 选中时高亮显示
- **AND** 显示模式说明："限时节奏挑战 - 在恰当时机按空格获得高分"

### Requirement: Sentence Mode Game Rendering
The system SHALL render the sentence typing game interface.

#### Scenario: Display target sentence
- **GIVEN** the game is in Sentence mode and running
- **WHEN** the game screen is rendered
- **THEN** it SHALL display the complete target sentence
- **AND** use a distinct style (e.g., gray or muted color)

#### Scenario: Display user input with color coding
- **GIVEN** the game is in Sentence mode and user has typed characters
- **WHEN** the game screen is rendered
- **THEN** each typed character SHALL be displayed below or alongside the target
- **AND** correct characters SHALL be rendered in green
- **AND** incorrect characters SHALL be rendered in red
- **AND** untyped positions SHALL remain empty or show placeholders

#### Scenario: Show real-time statistics
- **GIVEN** the game is in Sentence mode
- **WHEN** the game screen is rendered
- **THEN** it SHALL display total characters typed
- **AND** display correct character count
- **AND** display current accuracy percentage
- **AND** display elapsed time

### Requirement: Sentence Mode Results Rendering
The system SHALL render completion results for Sentence mode.

#### Scenario: Display sentence mode results
- **GIVEN** the game in Sentence mode has finished
- **WHEN** the results screen is rendered
- **THEN** it SHALL show the target sentence
- **AND** show the user's typed sentence with color coding
- **AND** display total keystrokes
- **AND** display correct characters
- **AND** display accuracy percentage
- **AND** display typing speed (characters per second)
- **AND** display elapsed time

### Requirement: Rhythm Dance Game Rendering
系统 SHALL 渲染节奏舞蹈模式的多单词队列界面。

#### Scenario: Display word queue as vertical list
- **WHEN** 渲染节奏舞蹈模式游戏界面
- **THEN** 系统应在屏幕左半部分垂直显示 5 个单词
- **AND** 所有单词靠右对齐
- **AND** 每个单词占一行
- **AND** 单词列表从上到下排列

#### Scenario: Highlight current inputtable word
- **WHEN** 渲染单词队列时
- **THEN** 系统应高亮显示当前可输入的单词（队列中间位置，索引2）
- **AND** 使用白色 `#FFFFFF` 显示
- **AND** 可选：在单词左侧显示箭头或标记（如 "═══>"）

#### Scenario: Display non-inputtable words in muted style
- **WHEN** 渲染单词队列时
- **THEN** 系统应使用渐变色显示其他单词：
- **AND** 索引0（已完成-2）：深灰色 `#444444`（初始状态为空字符串时不显示任何内容，仅保留行高）
- **AND** 索引1（已完成-1）：浅灰色 `#888888`（初始状态为空字符串时不显示任何内容，仅保留行高）
- **AND** 索引3（待输入+1）：浅灰色 `#888888`
- **AND** 索引4（待输入+2）：深灰色 `#444444`
- **AND** 这些单词不应有高亮或特殊标记

#### Scenario: Show rhythm bar only for current word
- **WHEN** 渲染节奏条时
- **THEN** 系统应将节奏条显示在屏幕右半部分
- **AND** 节奏条与当前可输入单词（索引2）垂直对齐
- **AND** 其他单词右侧不显示节奏条

#### Scenario: Display input buffer in current word row
- **WHEN** 渲染游戏界面时
- **THEN** 系统应在当前单词行（索引2）的最开始显示 "Input: [完整单词]"
- **AND** 始终显示完整的目标单词字符
- **AND** 根据用户输入缓冲区进行颜色编码：
  - 已正确输入的字符显示为绿色
  - 错误输入的字符显示为红色
  - 未输入的字符显示为白色
- **AND** 不在底部单独显示输入框

#### Scenario: Update word list on completion
- **WHEN** 玩家完成当前单词后
- **THEN** 系统应将整个队列上移（索引0被丢弃，1→0, 2→1, 3→2, 4→3）
- **AND** 新单词出现在底部（索引4）
- **AND** 高亮和节奏条保持在中间位置（索引2），但显示的是新的当前单词

### Requirement: Word Queue Visual Positioning
系统 SHALL 为单词队列提供清晰的视觉定位和布局。

#### Scenario: Align rhythm bar with current word
- **WHEN** 渲染节奏条时
- **THEN** 节奏条应始终与中间位置单词（索引2）垂直对齐
- **AND** 显示在屏幕右半部分
- **AND** 即使队列内容上移，节奏条位置保持在屏幕中间高度

#### Scenario: Preserve vertical spacing
- **WHEN** 渲染单词队列时
- **THEN** 系统应在单词之间保持一致的垂直间距
- **AND** 确保单词列表不会超出屏幕高度
- **AND** 单词列表整体在屏幕中央垂直居中
- **AND** 左右分栏布局：左侧单词列表靠右对齐，右侧节奏条和特效

