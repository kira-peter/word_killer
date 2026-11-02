# Design: Online Battle Mode

## Context
- 程序是基于 Bubble Tea 的单进程 TUI，`cmd/word-killer/main.go` 中的 `model` 用若干 bool 字段管理界面状态，每 33ms 触发一次 `tickMsg`
- 单词来自 `game.Game` 的多难度词库（`generateWordsFromMultiPools`）
- 目标场景是局域网或可以直连的网络环境，2–4 名玩家

## Goals / Non-Goals
- Goals：
  - 创建/加入房间，大厅里由房主开始对局
  - 单词掉落对战：判定公平，不会出现两个人同时拿到同一个单词
  - 协议可扩展，以后能加入更多对战玩法
- Non-Goals：
  - NAT 穿透、公网房间列表、匹配系统
  - 断线重连、观战
  - 加密与身份认证（4 位房间号只是防误连，不是安全措施）

## Decisions

### 1. 拓扑：房主进程同时运行 Server 和 Client
房主进程在后台 goroutine 中启动 `online.Server`，房主自己的 UI 也用 `online.Client` 通过 `127.0.0.1` 连上去。所有玩家的客户端代码路径完全一致，判定只在 Server 端进行。
- 备选方案：房主直接在进程内调用 Server 方法。这样会出现两套代码路径，房主还会有零延迟优势，所以不采用。

### 2. 传输：TCP + 换行分隔 JSON
每条消息是一行 JSON：`{"type":"...", ...}`，用 `bufio.Scanner` 读取、`json.Encoder` 写入，只依赖标准库。
- 备选方案：WebSocket 或 gRPC，需要引入新依赖，对这个场景来说过重。

### 3. 房间号：固定端口 + 4 位随机码
Server 监听 `online_port`（默认 7788）。创建房间时生成 `0000`–`9999` 之间的随机 4 位码。客户端连接后发送的第一条必须是 `join` 消息，带上房间号和昵称；房间号不对时回复 `join_rejected` 并断开连接。每个进程同一时间只承载一个房间。

### 4. 消息定义（`pkg/online/protocol.go`）
| 方向 | type | 字段 | 说明 |
|---|---|---|---|
| C→S | `join` | room, name | 加入请求 |
| S→C | `join_accepted` | player_id, room | 加入成功 |
| S→C | `join_rejected` | reason (`bad_room` / `room_full` / `in_game`) | 拒绝并断开 |
| S→C | `lobby` | players[{id,name,is_host}], mode | 玩家列表或玩法有变化时广播 |
| C→S | `select_mode` | mode | 仅房主可用 |
| C→S | `rename` | name | 大厅中修改自己的昵称 |
| C→S | `start` | - | 仅房主可用，且人数 ≥ 2 |
| S→C | `game_start` | mode, total_words, countdown_ms | 开局倒计时（3 秒） |
| S→C | `word_spawn` | index, word, fall_ms | 开始掉落第 index 个单词 |
| C→S | `claim` | index, word | 提交答案 |
| S→C | `word_claimed` | index, player_id, points | 单词被某个玩家拿到 |
| S→C | `word_expired` | index | 单词掉到底，作废 |
| S→C | `game_over` | ranking[{id,name,score,words}], winners[] | 对局结算 |
| C→S | `back_to_lobby` | - | 仅房主可用，所有人回到大厅 |
| S→C | `host_closed` | - | 房主退出，房间关闭 |

`player_id` 由 Server 按加入顺序分配，房主固定为 1。

### 5. 判定的权威与并发
- Server 用一个 goroutine 串行处理所有事件：每个连接的读 goroutine 把消息投递到同一个 channel，计时器到期也投递到这个 channel，因此不需要加锁
- `claim` 只有在 `index == 当前单词`、`word` 与当前单词完全相同、且该单词尚未被拿走时才有效，最先到达 Server 的有效 claim 获胜
- 客户端按回车时先在本地比对：不一致就清空输入、不发送任何消息；一致才发送 `claim`
- 单词被拿走或作废后间隔 1 秒，再发送下一个 `word_spawn`

### 6. 掉落位置的计算
Server 只下发 `fall_ms`。客户端以收到 `word_spawn` 的本地时间为起点，计算 `progress = elapsed / fall_ms`，再映射到掉落区的行号来渲染。Server 用自己的计时器（`fall_ms`）判断是否作废，并以 Server 的判断为准。局域网延迟可以忽略。

### 7. 与 Bubble Tea 的集成
`online.Client` 的读 goroutine 把解析好的消息写入带缓冲的 channel。`main.go` 在每次 `tickMsg`（33ms）时非阻塞地取完 channel 里的消息，交给 `online.Battle`（客户端状态）去更新，与现有的 tick 驱动方式保持一致。

### 8. 代码组织
- `pkg/online/protocol.go`：消息结构体、编解码
- `pkg/online/server.go`：`Server`（监听、房间、大厅、对局调度）
- `pkg/online/client.go`：`Client`（连接、发送、接收 channel）
- `pkg/online/battle.go`：客户端侧的对局状态（当前单词、掉落进度、输入、比分、结果）
- `pkg/ui/online.go`：联机菜单、加入表单、大厅、对战、结算界面
- Server 需要单词列表时，通过注入的 `func(n int) []string` 获取，由 main 用 `game.Game` 的多难度词库实现，这样 `pkg/online` 不依赖 `pkg/game`

### 9. 断线处理
- 大厅中有玩家断开：移出列表并广播 `lobby`
- 对局中有玩家断开：保留该玩家已有的分数；剩余在线玩家少于 2 人时立即发送 `game_over`
- 房主退出（Esc 离开或关闭程序）：Server 向所有人广播 `host_closed` 后关闭，客户端提示“Host closed the room”并回到联机菜单
- 客户端连接异常断开：显示“Connection lost”并回到联机菜单

## Risks / Trade-offs
- 端口被占用或被防火墙拦截 → 创建房间失败时给出明确提示，端口可以通过配置修改；README 说明需要放行端口
- 客户端之间的时钟差会导致掉落动画略有偏差 → 只影响显示，判定以 Server 为准
- 房间号只是 4 位数字，可以被枚举 → 不属于本次的目标，文档中说明仅适合在可信网络中使用

## Open Questions
- 无
