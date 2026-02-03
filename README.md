# Word Killer - 打字游戏

> 低调但绝不简单的命令行打字游戏

## 介绍

Word Killer 是一个基于命令行的打字游戏，通过消除屏幕上的单词来提升你的打字速度和准确度。

## 功能特性

- ✨ **经典游戏模式**: 消除屏幕中的所有单词
- 🎯 **实时匹配**: 输入时即时高亮匹配的单词
- ⏸️ **暂停功能**: 支持游戏暂停和恢复
- 📊 **详细统计**: 完整的数据统计（速度、准确率等）
- 🎨 **彩色界面**: 使用 ANSI 颜色的精美命令行界面
- ⚙️ **可配置**: 支持自定义词库和游戏设置

## 快速开始

### 编译

```bash
go build -o word-killer.exe ./cmd/word-killer
```

### 运行

```bash
./word-killer.exe
```

## 游戏玩法

1. **开始游戏**: 在欢迎界面按 `Enter` 开始
2. **输入匹配**: 输入字母匹配屏幕上的单词（自动高亮）
3. **消除单词**: 完整输入单词后按 `Enter` 消除
4. **暂停**: 按 `ESC` 进入暂停菜单
5. **退出**: 在暂停菜单中按 `ESC` 退出游戏

### 按键说明

| 按键 | 功能 |
|------|------|
| `a-z` | 输入字母 |
| `Enter` | 消除匹配的单词 |
| `Backspace` | 删除最后一个字符 |
| `ESC` | 进入暂停菜单（游戏中）/ 退出游戏（暂停菜单中） |
| `↑` / `↓` | 暂停菜单中导航 |

## 配置文件

配置文件 `config.json` 格式：

```json
{
  "word_dict_path": "data/words.txt",
  "word_count": 20
}
```

### 配置项说明

- `word_dict_path`: 单词词库文件路径
- `word_count`: 每场游戏的单词数量（0 表示使用全部词库）

## 词库文件

词库文件格式为文本文件，每行一个单词，仅包含字母（a-z）。

示例 `data/words.txt`:

```
hello
world
game
code
type
speed
```

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
│   └── word-killer/    # 主程序
├── pkg/
│   ├── config/         # 配置管理
│   ├── game/           # 游戏核心逻辑
│   ├── input/          # 输入处理
│   ├── stats/          # 统计系统
│   └── ui/             # UI 渲染
├── data/
│   └── words.txt       # 单词词库
├── config.json         # 配置文件
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
