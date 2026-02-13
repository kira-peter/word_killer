## ADDED Requirements

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
