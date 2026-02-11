## ADDED Requirements

### Requirement: Keyboard Event Listening
系统 SHALL 实时监听键盘事件，无阻塞地捕获用户输入。

#### Scenario: Capture printable character
- **WHEN** 用户按下可打印字符键（a-z, A-Z）
- **THEN** 系统立即捕获该字符
- **AND** 将字符添加到输入缓冲区
- **AND** 触发匹配检测

#### Scenario: Capture backspace
- **WHEN** 用户按下退格键
- **AND** 输入缓冲区非空
- **THEN** 系统删除缓冲区最后一个字符
- **AND** 触发匹配检测更新

#### Scenario: Capture enter key
- **WHEN** 用户按下回车键
- **THEN** 系统触发单词消除逻辑
- **AND** 处理匹配结果

#### Scenario: Capture ESC key
- **WHEN** 用户按下 ESC 键（游戏进行中）
- **THEN** 系统进入暂停菜单

#### Scenario: Capture ESC key in pause menu
- **WHEN** 用户按下 ESC 键（暂停菜单中）
- **THEN** 系统退出游戏

#### Scenario: Non-blocking input
- **WHEN** 等待键盘输入时
- **THEN** 系统不阻塞主游戏循环
- **AND** UI 渲染继续正常工作

### Requirement: Input Buffer Management
系统 SHALL 管理用户输入缓冲区，支持实时更新和清空。

#### Scenario: Append character to buffer
- **WHEN** 捕获到有效字符
- **THEN** 字符追加到缓冲区末尾
- **AND** 更新总按键计数

#### Scenario: Remove character from buffer
- **WHEN** 用户按下退格键
- **AND** 缓冲区包含至少一个字符
- **THEN** 移除最后一个字符

#### Scenario: Clear buffer on word elimination
- **WHEN** 单词成功消除
- **THEN** 输入缓冲区完全清空

#### Scenario: Buffer size limit
- **WHEN** 输入缓冲区长度达到最大限制（如 50 字符）
- **THEN** 系统忽略新的字符输入
- **AND** （可选）提供视觉提示

### Requirement: Input Validation
系统 SHALL 验证和过滤用户输入，仅接受有效字符。

#### Scenario: Accept valid letters
- **WHEN** 用户输入 a-z 或 A-Z
- **THEN** 系统接受该输入
- **AND** 转换为小写（如需要）

#### Scenario: Ignore invalid characters
- **WHEN** 用户输入数字、符号或特殊字符
- **THEN** 系统忽略该输入
- **AND** 不更新缓冲区
- **AND** 不增加有效按键计数

#### Scenario: Ignore during non-running state
- **WHEN** 游戏状态不是 Running
- **AND** 用户按下字符键
- **THEN** 系统忽略该输入（除特殊控制键外）
