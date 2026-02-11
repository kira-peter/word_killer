## ADDED Requirements

### Requirement: Word Pool Management
系统 SHALL 管理游戏中使用的单词池，支持加载和随机选择单词。

#### Scenario: Load word dictionary
- **WHEN** 游戏初始化时
- **THEN** 系统从配置文件指定的词库文件加载单词列表
- **AND** 验证单词格式正确（仅字母，长度合理）

#### Scenario: Generate game words
- **WHEN** 开始新游戏时
- **THEN** 系统从单词池随机选择单词
- **AND** 保证每场游戏中没有重复单词
- **AND** 单词数量和长度不受限制

### Requirement: Word Matching
系统 SHALL 实现精确的前缀匹配算法，支持用户输入与屏幕单词的实时匹配。

#### Scenario: Prefix match success
- **WHEN** 用户输入 "hel"
- **AND** 屏幕中存在单词 "hello"
- **THEN** 系统识别为前缀匹配
- **AND** 返回匹配的单词引用

#### Scenario: Multiple word match
- **WHEN** 用户输入 "te"
- **AND** 屏幕中存在 "test" 和 "term"
- **THEN** 系统返回所有匹配的单词

#### Scenario: No match
- **WHEN** 用户输入不匹配任何单词前缀
- **THEN** 系统返回空匹配结果

### Requirement: Word Elimination
系统 SHALL 支持完全匹配单词的消除功能。

#### Scenario: Complete word elimination
- **WHEN** 用户输入完整匹配某个单词
- **AND** 用户按下回车键
- **THEN** 系统从屏幕移除该单词
- **AND** 清空输入缓冲区
- **AND** 更新单词完成计数

#### Scenario: Incomplete match on enter
- **WHEN** 用户输入仅为部分匹配
- **AND** 用户按下回车键
- **THEN** 系统不消除任何单词
- **AND** 保持当前输入状态

#### Scenario: Ambiguous complete match
- **WHEN** 用户输入可完全匹配多个单词（如 "test" 同时完全匹配屏幕中的两个 "test"）
- **THEN** 系统消除第一个匹配的单词

### Requirement: Game State Management
系统 SHALL 维护清晰的游戏状态，支持状态转换和查询。

#### Scenario: Game initialization
- **WHEN** 游戏启动时
- **THEN** 状态设置为 Idle（待开始）
- **AND** 初始化空统计数据

#### Scenario: Game start
- **WHEN** 用户触发开始命令
- **THEN** 状态转换为 Running（进行中）
- **AND** 记录开始时间
- **AND** 生成游戏单词

#### Scenario: Game pause
- **WHEN** 游戏处于 Running 状态
- **AND** 用户按下 P 键或 Space 键
- **THEN** 状态转换为 Paused（已暂停）
- **AND** 记录暂停开始时间

#### Scenario: Game resume
- **WHEN** 游戏处于 Paused 状态
- **AND** 用户在暂停菜单选择"继续"
- **THEN** 状态转换回 Running
- **AND** 累加暂停时长到总暂停时长

#### Scenario: Game completion detection
- **WHEN** 所有屏幕单词被消除
- **THEN** 状态自动转换为 Finished（已完成）
- **AND** 记录结束时间
- **AND** 计算最终统计数据（排除暂停时长）

#### Scenario: Game abort from pause menu
- **WHEN** 用户在暂停菜单选择"结束"
- **THEN** 状态转换为 Finished
- **AND** 标记为提前退出

#### Scenario: Game abort by ESC
- **WHEN** 用户按下 ESC 键（非暂停状态）
- **THEN** 状态转换为 Finished
- **AND** 标记为提前退出

### Requirement: Pause Menu Management
系统 SHALL 提供暂停菜单，支持用户选择继续或结束游戏。

#### Scenario: Display pause menu
- **WHEN** 游戏进入 Paused 状态
- **THEN** 系统显示暂停菜单
- **AND** 菜单包含"继续 (Resume)"和"结束 (Quit)"选项
- **AND** 默认选中"继续"

#### Scenario: Navigate pause menu
- **WHEN** 暂停菜单显示时
- **AND** 用户按下上/下方向键
- **THEN** 系统切换选中的菜单项

#### Scenario: Confirm menu selection
- **WHEN** 暂停菜单显示时
- **AND** 用户按下回车键
- **THEN** 系统执行选中的操作（继续或结束）

#### Scenario: Ignore game input during pause
- **WHEN** 游戏处于 Paused 状态
- **THEN** 系统忽略所有游戏相关输入（字母键）
- **AND** 仅响应菜单导航和确认操作
