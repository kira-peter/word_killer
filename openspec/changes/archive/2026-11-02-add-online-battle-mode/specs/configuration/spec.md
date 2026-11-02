## ADDED Requirements

### Requirement: Online Mode Configuration
系统 SHALL 支持通过配置文件设置联机相关参数，并提供默认值。

#### Scenario: Default online settings
- **WHEN** 配置文件中没有联机相关字段
- **THEN** `online_port` 默认为 7788
- **AND** `online_drop_duration` 默认为 8（秒）
- **AND** `online_player_name` 默认为操作系统的用户名，取不到时为 “Player”

#### Scenario: Custom online settings
- **WHEN** 配置文件中设置了 `online_port`、`online_drop_duration`、`online_player_name`
- **THEN** 系统应使用配置的值
- **AND** 昵称超过 12 个字符时截断
