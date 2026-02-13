## MODIFIED Requirements

### Requirement: Rhythm Dance Mode Initialization
系统 SHALL 维护节奏舞蹈模式的完整状态，包括单词队列。

#### Scenario: Initialize rhythm dance state with word queue
- **WHEN** 启动节奏舞蹈模式
- **THEN** 系统应初始化 RhythmDanceState 结构
- **AND** 创建包含 5 个单词的单词队列（WordQueue）
- **AND** 前2个位置（索引0-1）设为空字符串或占位符（历史区初始为空）
- **AND** 后3个位置（索引2-4）填充随机单词（当前+预览区）
- **AND** 设置当前可输入单词索引为 2（中间位置）
- **AND** 初始化指针位置、速度等其他状态

#### Scenario: Update word queue on word completion
- **WHEN** 玩家成功完成当前单词（WordQueue[2]）
- **THEN** 系统应将整个 WordQueue 上移（丢弃索引0，1→0, 2→1, 3→2, 4→3）
- **AND** 生成新单词追加到 WordQueue 末尾（索引4）
- **AND** 保持 WordQueue 长度为 5
- **AND** 当前可输入单词始终在索引 2
- **AND** 清空输入缓冲区
