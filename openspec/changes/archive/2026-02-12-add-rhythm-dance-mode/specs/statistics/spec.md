## ADDED Requirements

### Requirement: Rhythm Dance Mode Statistics Tracking
系统 SHALL 追踪节奏舞蹈模式的专属统计数据。

#### Scenario: Track judgment counts
- **WHEN** 节奏舞蹈模式下执行判定
- **THEN** 系统应根据判定结果更新对应计数:
  - Perfect 判定 -> PerfectCount++
  - Nice 判定 -> NiceCount++
  - OK 判定 -> OKCount++
  - Miss 判定 -> MissCount++

#### Scenario: Track total score
- **WHEN** 节奏舞蹈模式下执行判定
- **THEN** 系统应根据判定等级增加总分:
  - Perfect -> 总分 + 5
  - Nice -> 总分 + 3
  - OK -> 总分 + 1
  - Miss -> 总分 + 0

#### Scenario: Track completed words count
- **WHEN** 玩家成功完成一个单词（任意判定等级）
- **THEN** 系统应增加完成单词计数
- **AND** 记录该单词的判定等级

#### Scenario: Calculate accuracy rate
- **WHEN** 计算准确率
- **THEN** 准确率 = (Perfect + Nice + OK) / (Perfect + Nice + OK + Miss) × 100%
- **AND** 保留两位小数

#### Scenario: Calculate average time per word
- **WHEN** 游戏结束时计算统计
- **THEN** 平均时间 = (游戏总时长 - 剩余时间) / 完成单词数
- **AND** 以秒为单位显示

#### Scenario: Track maximum combo
- **WHEN** 玩家连续获得 Perfect 或 Nice 判定
- **THEN** 系统应记录当前连击数（Combo）
- **AND** 更新最大连击数

#### Scenario: Reset combo on OK or Miss
- **WHEN** 玩家获得 OK 或 Miss 判定
- **THEN** 系统应将当前连击数重置为 0
- **AND** 保留最大连击数记录
