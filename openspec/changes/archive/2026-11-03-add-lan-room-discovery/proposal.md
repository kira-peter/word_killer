# Change: Add LAN Room Discovery

## Why
现在加入房间要手动输入房主 IP 和 4 位房间号，需要口头传递，操作麻烦。玩家希望进入联机菜单后，就能直接看到局域网里的房间，选中即可加入。

## What Changes
- **局域网房间发现**：
  - 玩家停留在联机菜单时，每秒向局域网广播一次 UDP 查询
  - 房主在 UDP `online_port` 上监听查询，并单播回复房间信息：房主昵称、TCP 端口、人数、状态
  - 超过 3 秒没有回复的房间从列表中移除
- **联机菜单改为房间列表**：
  - 菜单只有 Create Room，以及自动发现的房间
  - 每个房间显示房主昵称、人数、状态、IP
  - 等待中且未满的房间按回车直接加入；对局中或已满的房间灰色显示，不能加入
- **BREAKING：去掉房间号和手动加入**：
  - 创建房间不再生成房间号
  - 删除 Join Room 表单（IP 和房间号输入），加入房间的唯一方式是从局域网列表选择
  - `join` 消息去掉 `room` 字段，新增 `version`（协议版本号）；版本不一致时拒绝加入，提示 “Version mismatch”
  - 删除 `bad_room` 拒绝原因
- **防火墙**：房主需要放行 `online_port` 的 TCP 和 UDP；首次使用时 Windows 可能弹出网络访问提示

## Impact
- **受影响规范**：
  - `online-room` - 修改菜单和创建两个需求；删除房间号加入需求；新增局域网房间发现和从列表加入两个需求
  - `ui-rendering` - 联机界面渲染改为房间列表、删除加入表单、大厅不再显示房间号
- **受影响代码**：
  - `pkg/online/discovery.go`（新增）- 查询广播、回复、房间列表
  - `pkg/online/protocol.go` - 新增发现消息和协议版本号，去掉房间号相关字段
  - `pkg/online/server.go` - 去掉房间号校验，增加版本校验，在 UDP 上回复查询
  - `pkg/online/client.go`、`pkg/online/battle.go` - 去掉房间号
  - `pkg/ui/online.go` - 房间列表、大厅；删除加入表单
  - `cmd/word-killer/online.go` - 浏览房间的生命周期、列表选择；删除加入表单的状态和按键处理
  - `README.md`、`CONFIG.md` - 说明自动发现和需要放行的端口
- **兼容性**：
  - 2.0.0 的客户端无法加入新版房间（会提示 Join rejected），需要所有玩家使用同一版本
  - 只能和同一局域网（同一网段、广播可达）的玩家联机；跨网段、VPN、开启客户端隔离的网络无法加入
