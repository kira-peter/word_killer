## ADDED Requirements

### Requirement: Welcome Title Animation
系统 SHALL 在欢迎界面为 "Word Killer" 标题提供持续循环的动画效果。

#### Scenario: Word 抖动特效
- **WHEN** 用户查看欢迎界面
- **THEN** "Word" 文字 SHALL 显示随机抖动效果，每个字符独立进行随机位置偏移
- **AND** 抖动 SHALL 持续循环播放

#### Scenario: Killer 红色呼吸脉冲
- **WHEN** 用户查看欢迎界面
- **THEN** "Killer" 文字 SHALL 显示红色呼吸脉冲渐变效果
- **AND** 颜色 SHALL 在暗红色到亮红色之间循环渐变
- **AND** 渐变 SHALL 持续循环播放

#### Scenario: 动画帧管理
- **WHEN** 欢迎界面显示时
- **THEN** 系统 SHALL 维护动画帧计数器
- **AND** 帧计数 SHALL 随时间持续递增
- **AND** 动画效果 SHALL 基于当前帧数计算
