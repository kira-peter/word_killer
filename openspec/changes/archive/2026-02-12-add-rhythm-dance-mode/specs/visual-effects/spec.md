## ADDED Requirements

### Requirement: Rhythm Bar Rendering
系统 SHALL 渲染节奏条、指针和黄金分割点标记。

#### Scenario: Render rhythm bar
- **WHEN** 渲染节奏舞蹈模式界面
- **THEN** 系统应绘制长度为 40 字符的横向节奏条
- **AND** 节奏条应有明显的边界符号（如 ├─┤）
- **AND** 显示在屏幕中间右侧区域

#### Scenario: Render golden ratio marker
- **WHEN** 渲染节奏条
- **THEN** 系统应在黄金分割点位置（40 * 0.618 ≈ 第 25 个字符）绘制标记（如 ▼）
- **AND** 使用亮黄色高亮该标记

#### Scenario: Render pointer position
- **WHEN** 渲染节奏条
- **THEN** 系统应根据当前指针位置（0.0-1.0）计算字符位置
- **AND** 在对应位置绘制指针符号（如 ◆）
- **AND** 指针颜色应根据距离黄金点的远近渐变（近=亮黄，远=暗灰）

#### Scenario: Color gradient based on distance
- **WHEN** 计算指针颜色
- **THEN** 距离黄金点 <= 0.05 时使用亮黄色
- **AND** 距离 > 0.05 且 <= 0.15 时使用黄色
- **AND** 距离 > 0.15 且 <= 0.30 时使用暗黄色
- **AND** 距离 > 0.30 时使用灰色

### Requirement: Judgment Visual Effects
系统 SHALL 根据判定等级渲染对应的视觉特效。

#### Scenario: Perfect effect
- **WHEN** 判定为 Perfect
- **THEN** 系统应在节奏条周围渲染黄色边框闪烁（3 次）
- **AND** 在节奏条上方显示 "PERFECT!!" 文字
- **AND** 文字应上浮并逐渐淡出（1 秒内完成）

#### Scenario: Nice effect
- **WHEN** 判定为 Nice
- **THEN** 系统应在节奏条周围渲染蓝色边框闪烁（2 次）
- **AND** 在节奏条上方显示 "Nice!" 文字
- **AND** 文字应持续显示 0.5 秒后消失

#### Scenario: OK effect
- **WHEN** 判定为 OK
- **THEN** 系统应在节奏条周围渲染白色边框闪烁（1 次）
- **AND** 在节奏条上方显示 "OK" 文字
- **AND** 文字应持续显示 0.5 秒后消失

#### Scenario: Miss effect
- **WHEN** 判定为 Miss
- **THEN** 系统应在节奏条周围渲染红色边框（不闪烁）
- **AND** 在节奏条上方显示 "Miss..." 文字
- **AND** 文字应持续显示 0.5 秒后消失

### Requirement: Word Display in Rhythm Mode
系统 SHALL 在节奏舞蹈模式下渲染单词和输入。

#### Scenario: Render current word
- **WHEN** 渲染节奏舞蹈模式界面
- **THEN** 系统应在屏幕中间左侧显示当前目标单词
- **AND** 未输入的字母显示为普通白色
- **AND** 已正确输入的字母显示为绿色
- **AND** 已错误输入的字母显示为红色

#### Scenario: Render input buffer
- **WHEN** 玩家输入字母
- **THEN** 系统应在单词下方显示当前输入
- **AND** 使用下划线分隔每个字母（如 w_o_r_d）
- **AND** 输入匹配时字母变为绿色，不匹配时为红色

### Requirement: Dance Character Rendering
系统 SHALL 渲染舞蹈小人的 ASCII 动画。

#### Scenario: Render dancer in upper area
- **WHEN** 渲染节奏舞蹈模式界面
- **THEN** 系统应在屏幕上部中央渲染舞蹈小人
- **AND** 小人占用约 3-5 行高度
- **AND** 根据当前动画状态显示对应帧

#### Scenario: Render Idle frames
- **WHEN** 动画状态为 Idle
- **THEN** 系统应在两个 Idle 帧之间切换
- **AND** 帧 1: 双手垂下姿态
- **AND** 帧 2: 双手稍抬起姿态

#### Scenario: Render judgment animation frames
- **WHEN** 动画状态为 Perfect/Nice/OK/Miss
- **THEN** 系统应按顺序播放对应的动画帧
- **AND** Perfect: 跳跃姿态（手臂上举）
- **AND** Nice: 摆臂姿态（单手举起）
- **AND** OK: 点头姿态（身体前倾）
- **AND** Miss: 摔倒姿态（身体倾斜）

### Requirement: Rhythm Mode Statistics Display
系统 SHALL 显示节奏舞蹈模式的实时统计信息。

#### Scenario: Display time and score
- **WHEN** 渲染节奏舞蹈模式界面
- **THEN** 系统应在屏幕底部显示剩余时间
- **AND** 显示当前总分
- **AND** 剩余时间 < 10 秒时以红色或闪烁显示

#### Scenario: Display judgment counts
- **WHEN** 渲染节奏舞蹈模式界面
- **THEN** 系统应显示 Perfect、Nice、OK、Miss 各档次的计数
- **AND** 使用不同颜色区分（Perfect=黄色，Nice=蓝色，OK=白色，Miss=红色）

#### Scenario: Display final results
- **WHEN** 游戏结束
- **THEN** 系统应显示最终统计:
  - 总分
  - 完成单词数
  - Perfect/Nice/OK/Miss 各档计数
  - 准确率（正确字母数 / 总字母数）
  - 平均每词用时
