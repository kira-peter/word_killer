# game-core Specification

## Purpose
TBD - created by archiving change add-classic-game-mode. Update Purpose after archive.
## Requirements
### Requirement: Word Pool Management
游戏 SHALL 支持加载多个难度级别的单词词库,并根据配置的比例从不同难度池中随机选择单词。

#### Scenario: Load multiple difficulty dictionaries
- **WHEN** 游戏初始化时调用 LoadWordDict
- **THEN** 系统应同时加载 short、medium、long 三个难度的单词文件
- **AND** 每个词库的单词应存储在独立的内存池中
- **AND** 验证每个文件至少包含 1 个有效单词

#### Scenario: Handle missing difficulty files
- **WHEN** 某个难度文件不存在或为空
- **THEN** 系统应记录警告信息
- **AND** 继续加载其他可用的难度文件
- **AND** 如果所有文件都不可用,返回错误

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

### Requirement: Difficulty-Based Word Generation
游戏 SHALL 根据配置的难度比例从对应的词库中生成单词列表。

#### Scenario: Generate words with configured ratios
- **WHEN** 开始游戏时指定生成 20 个单词,配置比例为 short:30%, medium:50%, long:20%
- **THEN** 系统应生成约 6 个 short 单词、10 个 medium 单词、4 个 long 单词
- **AND** 单词应在三个词库中不重复选择
- **AND** 总单词数应等于请求的数量

#### Scenario: Handle insufficient words in a pool
- **WHEN** 某个难度池的单词数量不足以满足配置比例
- **THEN** 系统应从该池中选择所有可用单词
- **AND** 从其他难度池中补充剩余数量
- **AND** 优先保证总单词数量正确

#### Scenario: Cross-pool word deduplication
- **WHEN** 从多个词库中选择单词时
- **THEN** 系统应确保不会选择重复的单词
- **AND** 即使同一单词出现在多个难度文件中,也只能被选中一次

### Requirement: Ratio Normalization
系统 SHALL 自动归一化用户配置的难度比例值,支持任意正数形式。

#### Scenario: Normalize percentage-based ratios
- **WHEN** 用户配置 short_ratio=30, medium_ratio=50, long_ratio=20
- **THEN** 系统应识别总和为 100
- **AND** 直接使用这些比例值 (30%, 50%, 20%)

#### Scenario: Normalize arbitrary ratio values
- **WHEN** 用户配置 short_ratio=1, medium_ratio=2, long_ratio=1
- **THEN** 系统应计算总和为 4
- **AND** 归一化为 25%, 50%, 25%

#### Scenario: Handle zero ratio values
- **WHEN** 用户配置某个难度比例为 0 (如 short_ratio=0, medium_ratio=70, long_ratio=30)
- **THEN** 系统应从其他非零比例的词库中选择单词
- **AND** 不从比例为 0 的词库中选择任何单词

#### Scenario: Reject invalid ratio configuration
- **WHEN** 用户配置的所有比例值都 <= 0 (如 short_ratio=0, medium_ratio=0, long_ratio=0)
- **THEN** 系统应返回配置错误
- **AND** 提示用户至少一个比例值必须 > 0

