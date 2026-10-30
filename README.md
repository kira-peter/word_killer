# Word Killer - 打字游戏

> 低调但绝不简单的命令行打字游戏

## 介绍

Word Killer 是一个基于命令行的打字游戏，通过消除屏幕上的单词来提升你的打字速度和准确度。

## 功能特性

- ✨ **多种游戏模式**:
  - 🎯 **经典模式**: 消除固定数量的单词
  - 📝 **句子模式**: 完整输入句子
  - ⏱️ **倒计时模式**: 60秒限时挑战
  - 🏃 **极速模式**: 25词速度竞赛，追踪最佳记录
  - 🎵 **节奏大师**: 逐词限时，难度递增
- 🎯 **实时匹配**: 输入时即时高亮匹配的单词
- ⏸️ **暂停功能**: 支持游戏暂停和恢复
- 📊 **详细统计**: 完整的数据统计（速度、准确率等）
- 🎨 **彩色界面**: 使用 ANSI 颜色的精美命令行界面
- ⚙️ **可配置**: 支持自定义词库和游戏设置

## 快速开始

### 编译

```bash
make build        # 输出到 build/（包含 word-killer.exe、config.json、data/）
# 或直接使用 go build
go build -o word-killer.exe ./cmd/word-killer
```

### 运行

```bash
make run          # 编译并在 build/ 目录内运行
# 或
./word-killer.exe
```

执行 `make help` 可以查看全部目标（test、clean、build-all 等）。

## 游戏玩法

### 开始游戏

1. 运行游戏后，进入欢迎界面
2. 选择 **开始游戏**，进入模式选择
3. 使用 `↑` `↓` 键选择游戏模式：
   - **经典模式**: 消除固定数量单词（默认50个）
   - **句子模式**: 完整输入英文句子
   - **倒计时模式**: 60秒限时，消除尽可能多的单词
   - **极速模式**: 固定25个单词，追求最快完成时间
   - **节奏大师**: 每个单词限时完成，难度逐步递增
4. 按 `Enter` 开始游戏

### 游戏操作

1. **输入匹配**: 输入字母匹配屏幕上的单词（自动高亮）
2. **消除单词**: 完整输入单词后按 `Enter` 消除
3. **暂停**: 按 `ESC` 进入暂停菜单
4. **退出**: 在暂停菜单或结果页选择相应选项

### 按键说明

| 按键 | 功能 |
|------|------|
| `a-z` | 输入字母 |
| `Enter` | 消除匹配的单词 |
| `Backspace` | 删除最后一个字符 |
| `ESC` | 暂停游戏 / 返回上级菜单 |
| `↑` / `↓` / `k` / `j` | 菜单中导航 |

### 游戏模式详解

#### 🎯 经典模式
- 消除固定数量的单词（默认50个）
- 无时间限制，专注于准确性
- 适合练习打字基础

#### 📝 句子模式
- 完整输入英文句子
- 支持空格和标点符号
- 适合练习真实文本输入

#### ⏱️ 倒计时模式
- 60秒限时挑战
- 消除尽可能多的单词
- 剩余单词少于10个时自动补充
- 时间少于10秒时红色警告

#### 🏃 极速模式
- 固定25个单词
- 毫秒级精度计时
- 自动保存最佳记录
- 追求速度极限

#### 🎵 节奏大师
- 每个单词有时间限制（初始2秒）
- 每完成10个单词，时间限制减少0.1秒
- 最小时间限制0.5秒
- 任何单词超时即游戏结束
- 挑战反应速度和节奏感

## 配置文件

配置文件位置：`config.json`

### 快速配置示例

```json
{
  "word_count": 50,
  "short_ratio": 30,
  "medium_ratio": 50,
  "long_ratio": 20,
  "countdown_duration": 60,
  "speedrun_word_count": 25,
  "rhythm_initial_time_limit": 2.0,
  "rhythm_min_time_limit": 0.5,
  "rhythm_difficulty_step": 0.1,
  "rhythm_words_per_level": 10
}
```

### 主要配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `word_count` | 经典模式单词数量 | 50 |
| `short_ratio` | 短单词比例权重 | 30 |
| `medium_ratio` | 中等单词比例权重 | 50 |
| `long_ratio` | 长单词比例权重 | 20 |
| `countdown_duration` | 倒计时模式时长（秒） | 60 |
| `speedrun_word_count` | 极速模式单词数 | 25 |
| `rhythm_initial_time_limit` | 节奏大师初始时限（秒） | 2.0 |
| `rhythm_min_time_limit` | 节奏大师最小时限（秒） | 0.5 |
| `rhythm_difficulty_step` | 节奏大师每级减少时间（秒） | 0.1 |
| `rhythm_words_per_level` | 节奏大师每级所需单词数 | 10 |

> 📖 **详细配置说明**: 请查看 [CONFIG.md](CONFIG.md) 获取完整的配置文档和示例

### 难度预设

**简单模式** (适合新手):
```json
{
  "short_ratio": 50, "medium_ratio": 40, "long_ratio": 10,
  "countdown_duration": 90,
  "speedrun_word_count": 15,
  "rhythm_initial_time_limit": 3.0,
  "rhythm_min_time_limit": 1.0
}
```

**困难模式** (适合高手):
```json
{
  "short_ratio": 10, "medium_ratio": 40, "long_ratio": 50,
  "countdown_duration": 30,
  "speedrun_word_count": 40,
  "rhythm_initial_time_limit": 1.5,
  "rhythm_min_time_limit": 0.3
}
```

## 词库文件

项目包含三个难度级别的词库：

- `data/google-10000-short.txt`: 短单词（3-5个字母）
- `data/google-10000-medium.txt`: 中等单词（6-8个字母）
- `data/google-10000-long.txt`: 长单词（9+个字母）
- `data/sentences.txt`: 句子库（用于句子模式）

词库文件格式为文本文件，每行一个单词/句子，仅包含字母（a-z）。

## 统计指标

游戏结束后会显示以下统计数据：

- **总敲击数**: 所有按键次数
- **有效敲击数**: 匹配到单词的按键次数
- **正确字符数**: 正确输入的字符数
- **完成单词数**: 成功消除的单词数
- **总字母数**: 所有已消除单词的总字母数
- **总耗时**: 游戏时长（排除暂停时间）
- **字母速度**: 字母数/秒
- **单词速度**: 单词数/秒
- **准确率**: 正确字符数 / 总敲击数 × 100%

## 项目结构

```
word-killer/
├── cmd/
│   └── word-killer/       # 主程序
├── pkg/
│   ├── config/            # 配置管理
│   ├── game/              # 游戏核心逻辑
│   ├── stats/             # 统计系统
│   └── ui/                # UI 渲染
├── data/
│   ├── google-10000-short.txt   # 短单词词库
│   ├── google-10000-medium.txt  # 中等单词词库
│   ├── google-10000-long.txt    # 长单词词库
│   └── sentences.txt            # 句子词库
├── config.json            # 配置文件
├── CONFIG.md              # 配置文件详细说明
├── go.mod
└── README.md
```

## 系统要求

- Go 1.16 或更高版本
- 支持 ANSI 转义序列的终端
  - Windows: Windows Terminal, PowerShell, Git Bash
  - macOS: Terminal, iTerm2
  - Linux: 大部分终端

## 开发

### 依赖

- `github.com/charmbracelet/bubbletea`: TUI 框架
- `github.com/charmbracelet/lipgloss`: 终端样式库

### 运行测试

```bash
go test ./...
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

---

**享受打字的乐趣！** 🎮⌨️
