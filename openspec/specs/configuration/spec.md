# configuration Specification

## Purpose
Defines the configuration file fields and their defaults: difficulty ratios, dictionary paths, per-mode parameters and online settings. Missing or invalid values fall back to built-in defaults.

## Requirements

### Requirement: Difficulty Ratio Configuration
配置系统 SHALL 支持用户设定 short、medium、long 单词的随机比例。

#### Scenario: Load default difficulty ratios
- **WHEN** 用户未在 config.json 中配置难度比例
- **THEN** 系统应使用默认比例值 (short: 30, medium: 50, long: 20)
- **AND** 这些默认值应在 DefaultConfig 函数中定义

#### Scenario: Load custom difficulty ratios
- **WHEN** 用户在 config.json 中配置 {"short_ratio": 10, "medium_ratio": 20, "long_ratio": 70}
- **THEN** 系统应成功加载这些自定义比例
- **AND** 在游戏初始化时使用这些值

#### Scenario: Validate ratio values are positive
- **WHEN** 配置文件被加载
- **THEN** 系统应验证每个比例值 >= 0
- **AND** 至少一个比例值 > 0
- **AND** 如果验证失败,返回清晰的错误信息

### Requirement: Word Dictionary Path Configuration
配置系统 SHALL 支持指定三个难度级别的单词文件路径。

#### Scenario: Use default dictionary paths
- **WHEN** 用户未配置词库路径
- **THEN** 系统应使用默认路径:
  - short: `data/google-10000-short.txt`
  - medium: `data/google-10000-medium.txt`
  - long: `data/google-10000-long.txt`

#### Scenario: Load custom dictionary paths
- **WHEN** 用户在 config.json 中配置自定义路径 {"short_dict_path": "custom/short.txt", ...}
- **THEN** 系统应从指定路径加载词库文件
- **AND** 如果文件不存在,返回明确的错误信息

#### Scenario: Backward compatibility with legacy config
- **WHEN** 用户只配置了旧版 `word_dict_path` 字段,未配置难度路径
- **THEN** 系统应将该路径用作 medium 难度词库
- **AND** short 和 long 词库使用默认路径
- **AND** 难度比例默认为 medium: 100, short: 0, long: 0 (保持原有行为)

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
