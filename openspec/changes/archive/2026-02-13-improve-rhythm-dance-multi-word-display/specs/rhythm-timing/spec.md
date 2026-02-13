## MODIFIED Requirements

### Requirement: Rhythm Timing Judgment
系统 SHALL 基于指针位置和当前队列第一个单词计算节奏判定等级。

#### Scenario: Judge timing for current queue word
- **WHEN** 玩家按下空格键或回车键
- **AND** 当前输入与单词队列中间位置单词（WordQueue[2]）完全匹配
- **THEN** 系统应基于指针位置执行节奏判定
- **AND** 根据判定等级更新分数和计数

#### Scenario: Miss judgment on incorrect queue word input
- **WHEN** 玩家按下空格键或回车键
- **AND** 当前输入与单词队列中间位置单词（WordQueue[2]）不完全匹配
- **THEN** 系统应判定为 Miss
- **AND** Miss 计数加 1
- **AND** 扣 1 分
- **AND** 清空输入缓冲区但不移除单词
