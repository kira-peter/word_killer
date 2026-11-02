# Change: Add Online Battle Mode

## Why
目前所有玩法都是单人练习，缺少和别人比拼的乐趣。新增联机玩法：一名玩家创建房间，其他玩家输入房主 IP 和房间号加入，在同一局里抢着输入单词，让打字练习变成实时对战。

第一版先提供一种对战玩法“单词掉落（Word Drop）”：屏幕上方依次掉落 10 个单词，谁先完整无误地输入并按回车，谁就拿到这个单词，得分等于单词的字母数，最后总分高者获胜。

## What Changes
- **欢迎界面新增 Online 入口**：进入后可以选择 Create Room 或 Join Room
- **房间系统**：
  - 房主在固定端口（默认 7788，可配置）开启 TCP 监听，并随机生成 4 位房间号
  - 加入者输入房主 IP 和房间号；房间号错误、房间已满或正在对局时会被拒绝
  - 每个房间 2–4 人，房主自己也参与对局
  - 房间大厅显示房间号、房主本机 IP、玩家列表；由房主选择玩法并开始游戏
- **单词掉落对战（Word Drop）**：
  - 共 10 个单词，一次只掉一个，掉落时长可配置（默认 8 秒）
  - 第一个提交正确单词的玩家获得这个单词，得分 = 字母数
  - 单词掉到底还没人拿到就作废，也计入这 10 个
  - 按回车时如果输错，就清空输入，不扣分
  - 10 个单词结束后显示排名，总分最高者获胜，同分则并列
- **网络协议**：房主进程权威裁决，使用换行分隔的 JSON 消息；只依赖 Go 标准库 `net`
- **新增配置项**：`online_port`、`online_player_name`、`online_drop_duration`

这是全新功能，不涉及破坏性变更，现有单机模式不受影响。

## Impact
- **新增能力规范**：
  - `online-room` - 房间创建、加入、大厅、断线处理
  - `online-word-drop` - 单词掉落对战规则、判定与结算
- **受影响规范**：
  - `ui-rendering` - 欢迎菜单的 Online 入口，以及联机菜单、加入表单、大厅、对战、结算界面
  - `configuration` - 新增联机相关配置项
- **受影响代码**：
  - `pkg/online/`（新增）- `protocol.go`、`server.go`、`client.go`、`battle.go`
  - `pkg/ui/online.go`（新增）- 联机相关界面渲染
  - `pkg/ui/styles.go` - 欢迎菜单增加 Online 选项
  - `pkg/config/config.go`、`config.json`、`config.example.json` - 新增配置项
  - `cmd/word-killer/main.go` - 联机界面状态机、按键处理、网络消息分发
- **依赖**：不新增第三方依赖
