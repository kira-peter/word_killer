## MODIFIED Requirements

### Requirement: Input Validation
系统 SHALL 根据当前游戏模式验证和过滤用户输入。

#### Scenario: Accept valid letters in Classic mode
- **WHEN** 游戏处于 Classic 模式
- **AND** 用户输入 a-z 或 A-Z
- **THEN** 系统接受该输入
- **AND** 转换为小写

#### Scenario: Accept printable characters in Sentence mode
- **WHEN** 游戏处于 Sentence 模式
- **AND** 用户输入可打印字符（字母、数字、标点、空格）
- **THEN** 系统接受该输入
- **AND** 与目标句子对应字符比较
- **AND** 更新正确性统计

#### Scenario: Ignore invalid characters
- **WHEN** 用户输入非法字符（根据当前模式）
- **THEN** 系统忽略该输入
- **AND** 不更新缓冲区
- **AND** 不增加有效按键计数

#### Scenario: Ignore during non-running state
- **WHEN** 游戏状态不是 Running
- **AND** 用户按下字符键
- **THEN** 系统忽略该输入（除特殊控制键外）

## ADDED Requirements

### Requirement: Mode-Specific Enter Key Handling
系统 SHALL 根据游戏模式处理回车键。

#### Scenario: Enter in Classic mode
- **WHEN** 游戏处于 Classic 模式
- **AND** 用户按下回车键
- **THEN** 系统尝试消除匹配的单词
- **AND** 成功时清空输入缓冲区

#### Scenario: Enter in Sentence mode
- **WHEN** 游戏处于 Sentence 模式
- **AND** 用户按下回车键
- **AND** 输入长度等于目标句子长度
- **THEN** 系统结束游戏并显示结果

#### Scenario: Enter before sentence completion
- **WHEN** 游戏处于 Sentence 模式
- **AND** 用户按下回车键
- **AND** 输入长度小于目标句子长度
- **THEN** 系统忽略回车键
- **AND** 不结束游戏
