## ADDED Requirements

### Requirement: Screen Initialization
系统 SHALL 正确初始化和清理命令行屏幕环境。

#### Scenario: Terminal setup on game start
- **WHEN** 游戏启动时
- **THEN** 系统清空终端屏幕
- **AND** 隐藏光标
- **AND** 切换到原始模式（raw mode）

#### Scenario: Terminal cleanup on game exit
- **WHEN** 游戏退出时
- **THEN** 系统显示光标
- **AND** 恢复终端正常模式
- **AND** 清空或保留最终输出（可配置）

### Requirement: Word List Rendering
系统 SHALL 在屏幕上清晰渲染当前所有单词。

#### Scenario: Render active words
- **WHEN** 渲染游戏画面时
- **AND** 存在未消除的单词
- **THEN** 系统显示所有单词列表
- **AND** 每个单词占一行或分列显示

#### Scenario: Update word display on elimination
- **WHEN** 某个单词被消除后
- **THEN** 系统从显示列表移除该单词
- **AND** 重新渲染剩余单词

### Requirement: Input Highlighting
系统 SHALL 高亮显示当前匹配的单词，提供即时视觉反馈。

#### Scenario: Highlight matched word
- **WHEN** 用户输入匹配某个单词前缀
- **THEN** 系统高亮显示该单词（如使用不同颜色）
- **AND** 高亮匹配的字符部分

#### Scenario: Multiple word highlighting
- **WHEN** 用户输入匹配多个单词
- **THEN** 系统高亮所有匹配的单词

#### Scenario: Remove highlight on mismatch
- **WHEN** 用户输入不再匹配任何单词
- **THEN** 系统移除所有高亮
- **AND** 恢复单词正常显示

#### Scenario: Display current input
- **WHEN** 渲染游戏画面时
- **THEN** 系统在固定位置显示当前输入缓冲区内容
- **AND** 使用明显的视觉样式（如下划线或边框）

### Requirement: Statistics Display
系统 SHALL 实时显示游戏统计信息。

#### Scenario: Display real-time stats during game
- **WHEN** 游戏进行中
- **THEN** 系统在屏幕固定区域显示：
  - 剩余单词数量
  - 已完成单词数量
  - 当前用时
  - 实时速度（字母/秒）

#### Scenario: Display final results
- **WHEN** 游戏完成时
- **THEN** 系统显示完整统计报告：
  - 总敲击数
  - 有效敲击数
  - 正确字符数
  - 完成单词数
  - 总字母数
  - 总耗时
  - 平均字母速度（字母/秒）
  - 平均单词速度（单词/秒）
  - 准确率

### Requirement: Frame Rate Control
系统 SHALL 控制屏幕刷新频率，保证流畅体验。

#### Scenario: Limit refresh rate
- **WHEN** 游戏循环运行时
- **THEN** 系统控制渲染频率不超过 60fps
- **AND** 避免不必要的重绘

#### Scenario: Immediate render on input
- **WHEN** 用户输入导致视觉变化时
- **THEN** 系统立即触发重绘
- **AND** 显示更新后的状态

### Requirement: Cross-platform Compatibility
系统 SHALL 在主流平台上正确渲染。

#### Scenario: Render on Windows
- **WHEN** 在 Windows 终端运行
- **THEN** 系统正确显示所有元素和颜色

#### Scenario: Render on macOS/Linux
- **WHEN** 在 Unix-like 终端运行
- **THEN** 系统正确显示所有元素和颜色

#### Scenario: Fallback for unsupported terminals
- **WHEN** 检测到终端不支持彩色或高级特性时
- **THEN** 系统降级到基本文本渲染
- **AND** 保证核心功能可用

### Requirement: Pause Menu Rendering
系统 SHALL 在游戏暂停时渲染暂停菜单界面。

#### Scenario: Render pause menu
- **WHEN** 游戏状态为 Paused
- **THEN** 系统在屏幕中央或固定位置显示暂停菜单
- **AND** 显示标题"游戏已暂停"或"PAUSED"
- **AND** 列出菜单选项（继续、结束）

#### Scenario: Highlight selected menu item
- **WHEN** 渲染暂停菜单时
- **THEN** 系统高亮显示当前选中的菜单项
- **AND** 使用箭头、颜色或其他视觉提示标识选中状态

#### Scenario: Hide game content during pause
- **WHEN** 暂停菜单显示时
- **THEN** 系统隐藏或淡化游戏内容（可选）
- **AND** 暂停菜单清晰可见，不被遮挡
