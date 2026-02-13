## MODIFIED Requirements

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
