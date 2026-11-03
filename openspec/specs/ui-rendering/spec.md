# ui-rendering Specification

## Purpose
Defines terminal rendering for every screen (welcome, mode selection, game, pause, results and online), covering layout, highlighting, statistics display and refresh rate.

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

### Requirement: Online Interface Rendering
系统 SHALL 为联机流程提供与现有风格一致的界面（80 字符宽、顶部标题、中间内容、底部操作提示），房间只能从局域网列表加入。

#### Scenario: Welcome menu with Online option
- **WHEN** 显示欢迎界面
- **THEN** 菜单选项应为 Start、Online、About

#### Scenario: Online menu rendering
- **WHEN** 显示联机菜单
- **THEN** 界面上方显示 Create Room 选项
- **AND** 下方的 “Rooms on your network” 区域列出发现的房间，每行显示房主昵称、人数（例如 2/4）、状态（Waiting / Full / In game）和 IP
- **AND** 对局中或已满的房间用灰色显示
- **AND** 没有房间时显示 “Searching for rooms...”
- **AND** 有错误或正在连接时，在列表下方显示提示
- **AND** 底部提示为 “[↑↓] Select │ [Enter] Confirm │ [ESC] Back”

#### Scenario: Lobby rendering
- **WHEN** 显示房间大厅
- **THEN** 内容区分为左右两半：左半部分竖排显示菜单（玩法、该玩法的各项设置的名称和当前值、Start Game），房主选中的一项高亮并带 “>” 标记
- **AND** 右半部分每行一名玩家，每名玩家显示在一个圆角边框的卡片中，名字用该玩家的颜色，并标记 host 和 you；本人的卡片边框使用自己的颜色
- **AND** 顶部标题显示房主 IP:端口和人数（例如 Players 3/4）
- **AND** 房主看到的底部提示为 “[↑↓] Select │ [Enter] Open │ [K] Kick │ [Tab] Rename │ [ESC] Close”
- **AND** 非房主看到 “Waiting for host...” 和 “[Tab] Rename │ [ESC] Leave”

#### Scenario: Setting wheel rendering
- **WHEN** 房主选中某一项并按回车
- **THEN** 界面保持大厅布局不变，看起来像在原视图上操作：该项的名称留在原来的行和列，其他菜单项隐藏（位置留空），右半部分的玩家卡片照常显示
- **AND** 滚轮出现在该项取值的位置，当前值与名称同一行、与原取值同一列，高亮加粗并用 ▶ ◀ 标出
- **AND** 上下各显示两个相邻档位，离当前值越远颜色越淡；靠近两端时空出的位置留白
- **AND** 底部提示为 “[↑↓] Scroll │ [Enter] Confirm │ [ESC] Cancel”

#### Scenario: Kick picker rendering
- **WHEN** 房主打开踢人选择框
- **THEN** 标题变为 “Kick which player?”，右半部分的玩家卡片中，当前选中的玩家卡片边框变红并在名字前显示 ✕，房主自己的卡片变灰
- **AND** 底部提示为 “[↑↓] Select │ [Enter] Kick │ [ESC] Cancel”

#### Scenario: Word drop battle rendering
- **WHEN** 处于单词掉落对战中
- **THEN** 标题显示当前单词进度（例如 Word 3/10）和剩余时间，内容区顶部一行显示所有玩家的分数
- **AND** 掉落区按掉落进度在对应的行显示当前单词，并将与本人输入匹配的前缀高亮
- **AND** 掉落区下方再显示一次当前单词作为对齐基准，其下每名玩家一行（最多 4 行），输入与单词逐字对齐（规则见 Live input rows）
- **AND** 单词被拿走或作废时，在掉落区短暂显示结果提示

#### Scenario: Sentence race rendering
- **WHEN** 处于整句竞速中
- **THEN** 标题显示当前句子进度（例如 Sentence 2/5）和剩余时间，内容区顶部一行显示所有玩家的分数
- **AND** 句子按单词折成最多 2 行；每一行句子下面紧跟每名玩家对应这一行的输入，与句子逐字对齐（规则见 Live input rows），不显示进度条或百分比
- **AND** 名字超过 5 个字符时截断为 “前 3 个字符 + * + 后 2 个字符”，例如 AVeryLongNam 显示为 AVe*am
- **AND** 本人的光标显示在下一个字符所在的那一行
- **AND** 一句被拿走时，显示是谁拿到以及得分

#### Scenario: Word pool rendering
- **WHEN** 处于单词池对战中
- **THEN** 标题显示剩余时间和是否允许掠夺
- **AND** 上方以 6 行 × 6 列的网格显示盘面：空闲单词为白色，被占有的单词显示为占有者的颜色，与本人输入匹配的单词前缀高亮
- **AND** 下方每名玩家一行（最多 4 行），先显示该玩家正在输入的目标单词（盘面上与其输入开头相同部分最长的单词），再显示 “→” 和输入，按目标单词逐字比对（规则见 Live input rows）；还没输入时只显示 “...”
- **AND** 本模式中，其他玩家打对的字母用该玩家自己的颜色，而不是灰色

#### Scenario: Live input rows
- **WHEN** 对战中显示每名玩家的实时输入
- **THEN** 每行显示该玩家的昵称（本人显示为 You，昵称用该玩家的颜色）和当前输入
- **AND** 输入与目标逐字对齐，第 N 个字符显示在目标第 N 个字符的正下方
- **AND** 逐字比对：与上方目标字符不同的字符单独显示为红色，之后打对的字符照常显示；打错的空格显示为红色的 “·”
- **AND** 其他玩家打对的字符为灰色，本人打对的字符为绿色，本人带光标
- **AND** 本人的一行固定在最下面，其他玩家按进入房间的顺序排在上方；人数不足 4 人时空行留在顶部
- **AND** 其他玩家还没输入时显示 “...”

#### Scenario: Player colors
- **WHEN** 为玩家分配颜色
- **THEN** 4 种玩家颜色依次为蓝、黄、紫、青，不使用红色和绿色，避免与对错标记混淆

#### Scenario: Battle result rendering
- **WHEN** 显示对战结算
- **THEN** 界面应按排名列出玩家的昵称、分数、拿到的单词或句子数，并标记胜者
- **AND** 房主的提示为 “[Enter] Back to Lobby │ [ESC] Close Room”，非房主的提示为 “Waiting for host... │ [ESC] Leave”
