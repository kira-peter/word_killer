## MODIFIED Requirements

### Requirement: Input Validation
系统 SHALL 根据当前游戏模式验证和过滤用户输入。

#### Scenario: Accept letters in Rhythm Dance mode
- **WHEN** 游戏处于 Rhythm Dance 模式
- **AND** 用户输入 a-z 或 A-Z
- **THEN** 系统接受该输入
- **AND** 转换为小写
- **AND** 与目标单词对应位置字母比较

#### Scenario: Ignore non-letter input in Rhythm Dance mode
- **WHEN** 游戏处于 Rhythm Dance 模式
- **AND** 用户输入非字母字符（除空格外）
- **THEN** 系统忽略该输入

## ADDED Requirements

### Requirement: Space Key Timing in Rhythm Dance Mode
系统 SHALL 在节奏舞蹈模式下处理空格键作为节奏判定触发。

#### Scenario: Space triggers rhythm judgment
- **WHEN** 游戏处于 Rhythm Dance 模式且 Running 状态
- **AND** 用户按下空格键
- **AND** 当前输入与目标单词完全匹配
- **THEN** 系统应计算指针距离并执行节奏判定
- **AND** 根据判定等级更新分数
- **AND** 触发对应的动画和特效
- **AND** 切换到下一个单词

#### Scenario: Space ignored if word incomplete
- **WHEN** 游戏处于 Rhythm Dance 模式
- **AND** 用户按下空格键
- **AND** 当前输入与目标单词不完全匹配
- **THEN** 系统应忽略空格键
- **AND** 不执行节奏判定
- **AND** 不切换单词

#### Scenario: Space ignored if word wrong
- **WHEN** 游戏处于 Rhythm Dance 模式
- **AND** 用户按下空格键
- **AND** 当前输入包含错误字母
- **THEN** 系统应忽略空格键
- **AND** 判定为 Miss
- **AND** 清空输入缓冲区但不切换单词
