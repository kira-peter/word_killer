## ADDED Requirements

### Requirement: Rhythm Dance Mode Initialization
系统 SHALL 支持节奏舞蹈模式的初始化和启动。

#### Scenario: Initialize rhythm dance mode
- **WHEN** 用户选择节奏舞蹈模式并指定时长（如 60 秒）
- **THEN** 系统应设置游戏模式为 ModeRhythmDance
- **AND** 从词库中随机选择第一个单词
- **AND** 初始化指针位置为 0.5，方向为向右
- **AND** 初始化指针速度为配置的初始速度（默认 0.01）
- **AND** 初始化倒计时为指定时长
- **AND** 初始化得分为 0，各判定计数为 0
- **AND** 将游戏状态设置为 Running

#### Scenario: Handle unloaded dictionaries
- **WHEN** 启动节奏舞蹈模式时词库未加载
- **THEN** 系统应返回错误信息
- **AND** 不启动游戏

### Requirement: Rhythm Dance Mode Time Management
系统 SHALL 管理节奏舞蹈模式的倒计时。

#### Scenario: Track countdown timer
- **WHEN** 游戏处于 Running 状态
- **THEN** 系统应每帧更新剩余时间
- **AND** 剩余时间应持续递减

#### Scenario: Auto-end on timeout
- **WHEN** 倒计时归零
- **THEN** 系统应自动结束游戏
- **AND** 转换状态为 Finished
- **AND** 计算并保存最终统计数据

#### Scenario: Display time warning
- **WHEN** 剩余时间少于 10 秒
- **THEN** 系统应以红色或闪烁方式显示倒计时
- **AND** 提醒玩家时间紧迫

### Requirement: Rhythm Dance Word Progression
系统 SHALL 管理节奏舞蹈模式中的单词切换。

#### Scenario: Switch to next word on space
- **WHEN** 玩家按下空格键完成当前单词判定
- **THEN** 系统应从词库中随机选择下一个单词
- **AND** 清空输入缓冲区
- **AND** 增加完成单词计数
- **AND** 根据完成数判断是否需要加速指针

#### Scenario: Pointer speed increase
- **WHEN** 玩家完成一个单词
- **THEN** 系统应增加指针摆动速度
- **AND** 速度增量为 0.001（每完成一个单词）
- **AND** 更新 RhythmDanceState 中的 PointerSpeed 字段
