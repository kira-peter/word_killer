# Implementation Tasks

## 1. 配置
- [x] 1.1 `pkg/config/config.go` 新增 `OnlinePort`、`OnlineDropDuration`、`OnlinePlayerName` 字段及默认值
- [x] 1.2 加载配置时，零值字段回落到默认值，并处理昵称截断
- [x] 1.3 更新 `config.json`、`config.example.json`、`CONFIG.md`
- [x] 1.4 在 `pkg/config/config_test.go` 中为默认值和截断补充测试

## 2. 协议
- [x] 2.1 创建 `pkg/online/protocol.go`，定义消息类型常量和消息结构体
- [x] 2.2 实现换行分隔 JSON 的编码和解码辅助函数
- [x] 2.3 为编解码的往返补充单元测试

## 3. 服务端
- [x] 3.1 创建 `pkg/online/server.go`：监听端口、生成房间号、接受连接
- [x] 3.2 每个连接一个读 goroutine，消息汇入同一个事件 channel，由主循环串行处理
- [x] 3.3 实现 join 校验（bad_room / room_full / in_game）、分配 player_id、广播 lobby
- [x] 3.4 实现 select_mode、start（仅房主、人数 ≥ 2）、back_to_lobby
- [x] 3.5 实现 Word Drop 调度：取 10 个词、3 秒倒计时、依次 word_spawn、fall_ms 计时作废、间隔 1 秒
- [x] 3.6 实现 claim 判定：index 匹配、单词一致、先到先得，计分并广播
- [x] 3.7 实现 game_over 结算（排序、并列胜者）以及人数不足时提前结束
- [x] 3.8 实现断线移除，以及 Close 时广播 host_closed
- [x] 3.9 单词来源通过 `func(n int) []string` 注入

## 4. 客户端
- [x] 4.1 创建 `pkg/online/client.go`：带 3 秒超时连接、发送 join、读 goroutine 写入消息 channel
- [x] 4.2 提供 `Poll()`，非阻塞地取出所有待处理消息；连接断开时产生一个断开事件
- [x] 4.3 创建 `pkg/online/battle.go`：客户端对局状态（玩家、比分、当前单词、开始时间、输入、提示、结果）
- [x] 4.4 在 Battle 中实现本地回车比对：一致才发送 claim，否则清空输入

## 5. 服务端与客户端测试
- [x] 5.1 在 `pkg/online/server_test.go` 中用真实回环连接测试：加入成功、房间号错误、房间已满
- [x] 5.2 测试一局完整对局：两个客户端抢词、计分、作废、game_over 排名
- [x] 5.3 测试同一个单词被两人先后 claim 时只有一人得分
- [x] 5.4 测试对局中有人断线后提前结算

## 6. UI
- [x] 6.1 `pkg/ui/styles.go` 欢迎菜单改为 Start / Online / About
- [x] 6.2 创建 `pkg/ui/online.go`：联机菜单、加入表单（含错误提示）、大厅
- [x] 6.3 实现对战界面：比分栏、掉落区（按进度定位单词、前缀高亮）、输入框、结果提示、开局倒计时
- [x] 6.4 实现结算界面：排名、胜者标记，房主和非房主显示不同提示

## 7. 主程序集成
- [x] 7.1 `cmd/word-killer/main.go` 新增联机界面状态（menu / join / lobby / battle / result）
- [x] 7.2 欢迎界面支持 3 个选项，Online 进入联机菜单，About 的索引顺延
- [x] 7.3 Create Room：启动 Server，房主 Client 连接 127.0.0.1；失败时显示错误
- [x] 7.4 Join Room：表单输入处理、连接并处理拒绝原因
- [x] 7.5 在 tick 中轮询 Client 消息并更新 Battle 和界面状态；处理 host_closed 和断线
- [x] 7.6 各联机界面的按键处理，Esc 退出时清理 Client 和 Server
- [x] 7.7 在 main 中把 `game.Game` 的多难度词库包装成单词提供函数
- [x] 7.8 大厅支持按 Tab 修改昵称：新增 rename 消息，服务端校验后广播

## 8. 文档与验证
- [x] 8.1 README 新增联机玩法说明（创建、加入、端口放行）
- [x] 8.2 `go vet ./...`、`go test ./...` 通过
- [x] 8.3 手动验证：在同一台机器上开两个终端（房主 + 127.0.0.1 加入），完整打完一局
