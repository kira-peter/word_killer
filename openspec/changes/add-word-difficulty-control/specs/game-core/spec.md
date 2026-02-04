## MODIFIED Requirements

### Requirement: Word Dictionary Loading
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

## ADDED Requirements

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
