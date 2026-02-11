# statistics Specification

## Purpose
TBD - created by archiving change add-classic-game-mode. Update Purpose after archive.
## Requirements
### Requirement: Keystroke Counting
系统 SHALL 准确统计各类按键操作。

#### Scenario: Count total keystrokes
- **WHEN** 用户按下任意键（包括字母、退格、回车）
- **THEN** 系统增加总敲击计数器

#### Scenario: Count valid keystrokes
- **WHEN** 用户输入的字符能够匹配屏幕中单词的下一个字符
- **THEN** 系统增加有效敲击计数器

#### Scenario: Count correct characters
- **WHEN** 用户输入字符与目标单词对应位置字符完全匹配
- **THEN** 系统增加正确字符计数器

#### Scenario: Backspace not counted as valid
- **WHEN** 用户按下退格键
- **THEN** 总敲击数增加
- **AND** 有效敲击数不增加

### Requirement: Word and Letter Counting
系统 SHALL 统计完成的单词数量和字母数量。

#### Scenario: Count completed words
- **WHEN** 用户成功消除一个单词
- **THEN** 系统增加完成单词计数器

#### Scenario: Count total letters
- **WHEN** 用户成功消除一个单词
- **THEN** 系统将该单词的字母数量累加到总字母计数器

#### Scenario: Initial counts
- **WHEN** 游戏开始时
- **THEN** 单词计数和字母计数初始化为 0

### Requirement: Time Tracking
系统 SHALL 精确跟踪游戏时间，排除暂停时长。

#### Scenario: Record start time
- **WHEN** 游戏状态变为 Running
- **THEN** 系统记录当前时间戳作为开始时间

#### Scenario: Record pause time
- **WHEN** 游戏状态变为 Paused
- **THEN** 系统记录当前时间戳作为暂停开始时间

#### Scenario: Accumulate pause duration
- **WHEN** 游戏从 Paused 恢复到 Running
- **THEN** 系统计算本次暂停时长（当前时间 - 暂停开始时间）
- **AND** 累加到总暂停时长

#### Scenario: Record end time
- **WHEN** 游戏状态变为 Finished
- **THEN** 系统记录当前时间戳作为结束时间

#### Scenario: Calculate effective elapsed time
- **WHEN** 需要显示用时时
- **THEN** 系统计算有效时长 = (当前时间 - 开始时间) - 总暂停时长
- **AND** 以秒为单位返回结果

#### Scenario: Pause does not affect statistics
- **WHEN** 游戏处于 Paused 状态
- **THEN** 所有统计计数器不变化
- **AND** 时间不计入有效游戏时长

### Requirement: Speed Calculation
系统 SHALL 计算打字速度指标。

#### Scenario: Calculate letters per second
- **WHEN** 游戏进行中或结束时
- **THEN** 系统计算速度 = 总字母数 / 耗时（秒）
- **AND** 返回保留 2 位小数的结果

#### Scenario: Calculate words per second
- **WHEN** 游戏进行中或结束时
- **THEN** 系统计算速度 = 完成单词数 / 耗时（秒）
- **AND** 返回保留 2 位小数的结果

#### Scenario: Handle zero elapsed time
- **WHEN** 耗时为 0 或极小值（< 0.1 秒）
- **THEN** 系统返回速度为 0 或 "N/A"
- **AND** 避免除零错误

#### Scenario: Real-time speed update
- **WHEN** 游戏进行中
- **THEN** 系统每秒更新一次速度统计
- **AND** 用于实时显示

### Requirement: Accuracy Calculation
系统 SHALL 计算打字准确率。

#### Scenario: Calculate accuracy percentage
- **WHEN** 计算准确率时
- **THEN** 系统计算 准确率 = (正确字符数 / 总敲击数) × 100%
- **AND** 返回保留 2 位小数的百分比

#### Scenario: Handle zero keystrokes
- **WHEN** 总敲击数为 0
- **THEN** 系统返回准确率为 0% 或 "N/A"

#### Scenario: Perfect accuracy
- **WHEN** 所有输入都是有效且正确的
- **THEN** 系统返回 100% 准确率

### Requirement: Statistics Export
系统 SHALL 支持导出完整的统计数据。

#### Scenario: Get final statistics
- **WHEN** 游戏结束时
- **THEN** 系统返回包含所有统计指标的数据结构：
  - 总敲击数 (TotalKeystrokes)
  - 有效敲击数 (ValidKeystrokes)
  - 正确字符数 (CorrectChars)
  - 完成单词数 (WordsCompleted)
  - 总字母数 (TotalLetters)
  - 总耗时秒数 (ElapsedSeconds)
  - 字母速度 (LettersPerSecond)
  - 单词速度 (WordsPerSecond)
  - 准确率 (AccuracyPercent)

#### Scenario: Get real-time statistics
- **WHEN** 游戏进行中请求统计数据时
- **THEN** 系统返回当前实时统计
- **AND** 基于当前时间计算速度

