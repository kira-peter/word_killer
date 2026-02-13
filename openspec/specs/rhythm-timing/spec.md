# rhythm-timing Specification

## Purpose
TBD - created by archiving change add-rhythm-dance-mode. Update Purpose after archive.
## Requirements
### Requirement: Rhythm Timing Pointer Control
系统 SHALL 管理节奏指针的位置和摆动。

#### Scenario: Update pointer position per frame
- **WHEN** 游戏处于 Running 状态且为节奏舞蹈模式
- **THEN** 系统应每帧更新指针位置
- **AND** 新位置 = 当前位置 + 速度 × 方向
- **AND** 位置值应保持在 [0.0, 1.0] 范围内

#### Scenario: Pointer boundary loop
- **WHEN** 指针位置达到或超过 1.0（右边界）
- **THEN** 系统应将指针位置重置为 0.0（左边界）
- **AND** 方向保持不变（继续向右）
- **AND** 形成单向循环移动效果

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

### Requirement: Golden Ratio Position
系统 SHALL 使用黄金分割点作为判定基准。

#### Scenario: Define golden ratio position
- **WHEN** 初始化节奏舞蹈模式
- **THEN** 系统应设置黄金分割点位置为 0.618
- **AND** 该值在游戏过程中保持不变

