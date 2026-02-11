## ADDED Requirements

### Requirement: Rhythm Timing Pointer Control
系统 SHALL 管理节奏指针的位置和摆动。

#### Scenario: Update pointer position per frame
- **WHEN** 游戏处于 Running 状态且为节奏舞蹈模式
- **THEN** 系统应每帧更新指针位置
- **AND** 新位置 = 当前位置 + 速度 × 方向
- **AND** 位置值应保持在 [0.0, 1.0] 范围内

#### Scenario: Pointer boundary reversal
- **WHEN** 指针位置达到 1.0（右边界）
- **THEN** 系统应将方向反转为向左（-1）
- **AND** 指针位置设置为 1.0

#### Scenario: Pointer boundary reversal at left
- **WHEN** 指针位置达到 0.0（左边界）
- **THEN** 系统应将方向反转为向右（1）
- **AND** 指针位置设置为 0.0

### Requirement: Rhythm Timing Judgment
系统 SHALL 基于指针位置计算节奏判定等级。

#### Scenario: Perfect judgment
- **WHEN** 玩家按下空格键
- **AND** 指针距离黄金分割点（0.618）的绝对距离 <= 0.05
- **THEN** 系统应判定为 Perfect
- **AND** 增加 5 分
- **AND** Perfect 计数加 1

#### Scenario: Nice judgment
- **WHEN** 玩家按下空格键
- **AND** 指针距离黄金分割点的绝对距离 > 0.05 且 <= 0.15
- **THEN** 系统应判定为 Nice
- **AND** 增加 3 分
- **AND** Nice 计数加 1

#### Scenario: OK judgment
- **WHEN** 玩家按下空格键
- **AND** 指针距离黄金分割点的绝对距离 > 0.15 且 <= 0.30
- **THEN** 系统应判定为 OK
- **AND** 增加 1 分
- **AND** OK 计数加 1

#### Scenario: Miss judgment
- **WHEN** 玩家按下空格键
- **AND** 指针距离黄金分割点的绝对距离 > 0.30
- **THEN** 系统应判定为 Miss
- **AND** 不增加分数
- **AND** Miss 计数加 1

#### Scenario: Judgment only on correct word
- **WHEN** 玩家按下空格键
- **AND** 当前输入的字母与目标单词不完全匹配
- **THEN** 系统应忽略判定请求
- **AND** 不更新分数和判定计数
- **AND** 不切换到下一个单词

### Requirement: Golden Ratio Position
系统 SHALL 使用黄金分割点作为判定基准。

#### Scenario: Define golden ratio position
- **WHEN** 初始化节奏舞蹈模式
- **THEN** 系统应设置黄金分割点位置为 0.618
- **AND** 该值在游戏过程中保持不变
