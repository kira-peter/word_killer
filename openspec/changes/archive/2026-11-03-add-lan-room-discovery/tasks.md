# Implementation Tasks

## 1. 协议
- [x] 1.1 `pkg/online/protocol.go` 新增 `ProtocolVersion`、`discover` / `room_info` 消息类型和对应字段（version、room_id、port、max_players、in_game）
- [x] 1.2 去掉 `Room` 字段和 `ReasonBadRoom`，新增 `ReasonVersionMismatch`

## 2. 服务端
- [x] 2.1 去掉房间号的生成和校验；join 时校验协议版本
- [x] 2.2 生成随机 room_id；在 run 中维护由互斥锁保护的房间快照（房主昵称、人数、是否在对局中）
- [x] 2.3 在 UDP `online_port` 上监听 discover 并单播回复 room_info；房主加入前不回复
- [x] 2.4 TCP 或 UDP 监听失败时，NewServer 返回错误；Close 时同时关闭 UDP

## 3. 房间发现
- [x] 3.1 创建 `pkg/online/discovery.go`：`Browser` 从临时端口每秒发送 discover（受限广播、各网卡定向广播、127.0.0.1）
- [x] 3.2 解析 room_info：校验版本和包大小、按 room_id 去重并优先保留非回环地址、超过 3 秒移除
- [x] 3.3 提供 `Rooms()` 快照（排序稳定）和 `Close()`

## 4. 客户端
- [x] 4.1 `Connect(addr, name)` 去掉房间号参数，join 消息带上协议版本号；`RejectedError` 支持 version_mismatch
- [x] 4.2 `Battle` 去掉 Room 字段

## 5. UI
- [x] 5.1 `pkg/ui/online.go` 联机菜单改为 Create Room 加房间列表，包含灰色的不可加入房间、“Searching for rooms...”、连接中和错误提示
- [x] 5.2 删除加入表单的渲染
- [x] 5.3 大厅去掉房间号，标题和内容改为显示房主地址

## 6. 主程序集成
- [x] 6.1 进入联机菜单时启动 Browser，离开时关闭
- [x] 6.2 菜单选中项按 room_id 跟踪，列表刷新时保持选中；房间消失后回到 Create Room
- [x] 6.3 选中可加入的房间后异步连接，连接中忽略按键；不可加入时给出提示
- [x] 6.4 删除加入表单的状态和按键处理，去掉 Back 菜单项

## 7. 测试
- [x] 7.1 服务端测试改为无房间号；新增版本不一致被拒绝的测试
- [x] 7.2 发现测试：本机房主可被发现，人数和对局状态会更新，关房后 3 秒内消失
- [x] 7.3 发现测试：非法包、版本不一致的包、超大包被忽略；同一台电脑上两个 Browser 能同时发现房间
- [x] 7.4 UI 流程测试：从列表加入并打完一局；对局中的房间不可加入；选中项随列表刷新保持不变

## 8. 文档与验证
- [x] 8.1 README、CONFIG.md 说明自动发现、只支持局域网，以及需要放行 TCP 和 UDP 端口
- [x] 8.2 `go vet ./...`、`go test ./...` 通过
- [x] 8.3 手动验证：本机开两个终端，从列表直接加入并打完一局
