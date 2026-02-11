# Change: Add Word Difficulty Control

## Why
用户需要根据自己的打字水平选择合适难度的单词进行练习。通过允许配置不同难度单词的随机比例,可以让用户自定义训练强度,提供更加个性化的练习体验。例如初学者可以增加短单词比例,高级用户可以增加长单词比例。

## What Changes
- **多难度词库支持**: 支持同时加载三个难度级别的单词文件 (short/medium/long)
- **比例配置**: 在 config.json 中添加难度比例配置项
- **智能单词生成**: 根据配置的比例从不同难度词库中随机选择单词
- **比例归一化**: 自动处理用户输入的比例值,支持任意正数形式

这是对现有游戏核心逻辑的增强,不涉及破坏性变更。

## Impact
- **受影响规范**:
  - `game-core` - 需要修改单词加载和生成逻辑
  - `configuration` - 需要添加难度比例配置项

- **受影响代码**:
  - `pkg/config/config.go` - 添加难度比例配置字段
  - `pkg/game/game.go` - 修改 LoadWordDict 和 generateWords 方法
  - `config.json` - 添加默认比例配置

- **数据依赖**:
  - 需要三个难度级别的单词文件:
    - `data/google-10000-short.txt` (已存在,2184 个单词)
    - `data/google-10000-medium.txt` (已存在,5459 个单词)
    - `data/google-10000-long.txt` (已存在,2241 个单词)
