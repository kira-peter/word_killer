# input-handling Specification

## Purpose
TBD - created by archiving change add-classic-game-mode. Update Purpose after archive.
## Requirements
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
系统 SHALL 根据当前单词队列验证和过滤用户输入。

#### Scenario: Accept letters matching current queue word
- **WHEN** 游戏处于 Rhythm Dance 模式
- **AND** 用户输入 a-z 或 A-Z
- **AND** 输入字母与单词队列中间位置单词（WordQueue[2]）的对应位置匹配
- **THEN** 系统接受该输入
- **AND** 转换为小写
- **AND** 添加到输入缓冲区

#### Scenario: Reject letters not matching current queue word
- **WHEN** 游戏处于 Rhythm Dance 模式
- **AND** 用户输入 a-z 或 A-Z
- **AND** 输入字母与单词队列中间位置单词（WordQueue[2]）的对应位置不匹配
- **THEN** 系统拒绝该输入
- **AND** 不更新输入缓冲区
- **AND** 可选：播放错误提示音或视觉反馈

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

### Requirement: Rhythm Judgment Trigger Keys
系统 SHALL 在节奏舞蹈模式下处理空格键和回车键作为节奏判定触发。

#### Scenario: Space or Enter triggers rhythm judgment
- **WHEN** 游戏处于 Rhythm Dance 模式且 Running 状态
- **AND** 用户按下空格键或回车键
- **AND** 当前输入与目标单词完全匹配
- **THEN** 系统应计算指针距离并执行节奏判定
- **AND** 根据判定等级更新分数
- **AND** 触发对应的动画和特效
- **AND** 切换到下一个单词

#### Scenario: Judgment keys mark Miss if word wrong or incomplete
- **WHEN** 游戏处于 Rhythm Dance 模式
- **AND** 用户按下空格键或回车键
- **AND** 当前输入与目标单词不完全匹配（包含错误字母或长度不足）
- **THEN** 系统应判定为 Miss
- **AND** Miss 计数增加
- **AND** 重置连击数
- **AND** 总分扣 1 分
- **AND** 触发 Miss 动画
- **AND** 添加 Miss 到判定历史记录
- **AND** 清空输入缓冲区但不切换单词

