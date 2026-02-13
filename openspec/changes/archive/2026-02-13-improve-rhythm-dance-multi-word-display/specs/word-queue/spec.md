## ADDED Requirements

### Requirement: Word Queue Management
系统 SHALL 维护一个固定大小的单词队列，用于节奏舞蹈模式的多单词显示。

#### Scenario: Initialize word queue on game start
- **WHEN** 节奏舞蹈模式启动时
- **THEN** 系统应生成 5 个单词位置的队列
- **AND** 前2个位置（索引0-1）设为空字符串作为占位符（历史区初始为空）
- **AND** 后3个位置（索引2-4）填充不重复的随机单词
- **AND** 设置当前可输入单词索引为 2（中间位置）

#### Scenario: Maintain fixed queue size
- **WHEN** 游戏进行中
- **THEN** 单词队列的长度应始终保持为 5
- **AND** 队列中实际单词部分（非空字符串）不应出现重复

### Requirement: Word Completion and Queue Update
系统 SHALL 在玩家完成单词后更新队列，移除已完成单词并补充新单词。

#### Scenario: Remove completed word from queue
- **WHEN** 玩家成功完成当前可输入单词（队列中间位置，索引2）
- **AND** 节奏判定已执行
- **THEN** 系统应将整个队列上移（丢弃索引0，其他单词索引减1）
- **AND** 当前可输入单词位置保持在索引2，但内容变为原索引3的单词

#### Scenario: Append new word to queue end
- **WHEN** 队列上移后末尾需要补充新单词
- **THEN** 系统应立即生成一个新的随机单词
- **AND** 将新单词追加到队列末尾（索引 4）
- **AND** 确保队列长度恢复为 5

#### Scenario: Prevent duplicate words in queue
- **WHEN** 生成新单词准备追加到队列时
- **AND** 该单词已存在于当前队列中
- **THEN** 系统应重新生成另一个单词
- **AND** 重复此过程直到找到不重复的单词

### Requirement: Current Word Identification
系统 SHALL 明确标识当前可输入的单词，并限制输入仅对该单词有效。

#### Scenario: Identify current inputtable word
- **WHEN** 渲染游戏界面或处理输入时
- **THEN** 系统应将队列中间位置单词（WordQueue[2]）标记为当前可输入单词
- **AND** 其他队列中的单词标记为"仅显示，不可输入"

#### Scenario: Reject input for non-current words
- **WHEN** 玩家输入字母
- **AND** 输入的字母不匹配当前可输入单词（WordQueue[2]）的前缀
- **THEN** 系统应拒绝该输入
- **AND** 不更新输入缓冲区
