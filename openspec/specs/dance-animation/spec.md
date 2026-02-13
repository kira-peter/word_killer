# dance-animation Specification

## Purpose
TBD - created by archiving change add-rhythm-dance-mode. Update Purpose after archive.
## Requirements
### Requirement: Dance Character Animation States
系统 SHALL 根据判定结果显示不同的 ASCII 舞蹈动画。

#### Scenario: Idle animation loop
- **WHEN** 没有判定触发时
- **THEN** 系统应循环播放 Idle 动画（2 帧）
- **AND** 每帧持续约 10 个游戏帧（约 0.33 秒）

#### Scenario: Perfect animation trigger
- **WHEN** 判定结果为 Perfect
- **THEN** 系统应播放 Perfect 动画（跳跃姿态，3 帧）
- **AND** 动画播放完成后（约 1 秒）返回 Idle 状态

#### Scenario: Nice animation trigger
- **WHEN** 判定结果为 Nice
- **THEN** 系统应播放 Nice 动画（摆臂姿态，2 帧）
- **AND** 动画播放完成后（约 1 秒）返回 Idle 状态

#### Scenario: OK animation trigger
- **WHEN** 判定结果为 OK
- **THEN** 系统应播放 OK 动画（点头姿态，2 帧）
- **AND** 动画播放完成后（约 1 秒）返回 Idle 状态

#### Scenario: Miss animation trigger
- **WHEN** 判定结果为 Miss
- **THEN** 系统应播放 Miss 动画（摔倒姿态，2 帧）
- **AND** 动画播放完成后（约 1 秒）返回 Idle 状态

### Requirement: Animation Frame Management
系统 SHALL 管理动画帧的切换和计时。

#### Scenario: Track animation state
- **WHEN** 触发特定判定动画
- **THEN** 系统应记录当前动画类型（Perfect/Nice/OK/Miss）
- **AND** 重置动画帧计数器
- **AND** 开始播放对应动画序列

#### Scenario: Return to idle after animation
- **WHEN** 特效动画播放完成
- **THEN** 系统应自动切换回 Idle 状态
- **AND** 恢复 Idle 循环动画

#### Scenario: Animation interruption
- **WHEN** 正在播放某个判定动画时
- **AND** 新的判定触发
- **THEN** 系统应立即中断当前动画
- **AND** 开始播放新判定的动画

