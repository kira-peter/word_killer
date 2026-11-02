## ADDED Requirements

### Requirement: LAN Room Discovery
系统 SHALL 自动发现同一局域网内的房间，玩家不需要输入任何地址。

#### Scenario: Browse rooms
- **WHEN** 玩家停留在联机菜单
- **THEN** 系统应每秒向局域网广播一次 discover 查询（发往受限广播地址、每块网卡的定向广播地址和 127.0.0.1 的 UDP `online_port`）
- **AND** 离开联机菜单后停止发送查询

#### Scenario: Host answers discovery
- **WHEN** 房主收到协议版本一致的 discover 查询
- **THEN** 房主应向查询来源单播回复 room_info，包含 room_id、房主昵称、TCP 端口、当前人数、最大人数、是否在对局中
- **AND** 房主的 IP 由玩家从回复包的来源地址获取

#### Scenario: Room appears in list
- **WHEN** 玩家收到协议版本一致的 room_info
- **THEN** 该房间应出现在联机菜单的房间列表中，或者更新已有的条目
- **AND** 同一个 room_id 只显示一条，优先显示非回环地址

#### Scenario: Room disappears
- **WHEN** 某个房间超过 3 秒没有回复
- **THEN** 系统应将其从列表中移除

#### Scenario: Ignore foreign packets
- **WHEN** 收到的 UDP 包不是合法消息、协议版本不一致，或者超过 1 KB
- **THEN** 系统应丢弃该包，列表不受影响

#### Scenario: Several browsers on one machine
- **WHEN** 同一台电脑上运行多个客户端，并且都停留在联机菜单
- **THEN** 每个客户端都应能正常发现房间，彼此没有端口冲突

### Requirement: Room Joining from List
系统 SHALL 只允许玩家从局域网房间列表中选择房间加入，不提供手动输入地址或房间号的方式。

#### Scenario: Join a discovered room
- **WHEN** 玩家在房间列表中选中一个等待中且未满的房间并按回车
- **THEN** 系统应连接该房间的 IP 和 TCP 端口，发送带协议版本号的 join 消息
- **AND** 收到 join_accepted 后进入房间大厅
- **AND** 连接期间显示 “Connecting...”，并忽略按键

#### Scenario: Unjoinable room selected
- **WHEN** 玩家选中一个对局中的房间并按回车
- **THEN** 系统不应发起连接，并提示 “Game in progress”
- **AND** 选中已满的房间时提示 “Room is full”

#### Scenario: Version mismatch
- **WHEN** 玩家的协议版本与房主不一致
- **THEN** 服务端应回复 join_rejected（reason=version_mismatch）并断开连接
- **AND** 客户端显示 “Version mismatch”

#### Scenario: Room full
- **WHEN** 房间内已有 4 名玩家时又有人加入
- **THEN** 服务端应回复 join_rejected（reason=room_full）
- **AND** 客户端显示 “Room is full”

#### Scenario: Room in game
- **WHEN** 房间正在对局时有人加入
- **THEN** 服务端应回复 join_rejected（reason=in_game）
- **AND** 客户端显示 “Game in progress”

#### Scenario: Connection failure
- **WHEN** 无法在超时时间（3 秒）内连接到房主
- **THEN** 客户端应停留在联机菜单，显示 “Cannot connect to host”

## MODIFIED Requirements

### Requirement: Online Menu Entry
系统 SHALL 在欢迎界面提供 Online 入口，进入后显示联机菜单，包含创建房间和自动发现的房间列表。

#### Scenario: Enter online menu
- **WHEN** 玩家在欢迎界面选中 Online 并按回车
- **THEN** 系统应显示联机菜单，包含 Create Room 和局域网房间列表
- **AND** 开始在局域网中查找房间

#### Scenario: Selection stays on the same room
- **WHEN** 房间列表刷新，有房间出现或消失
- **THEN** 当前选中的房间应保持选中
- **AND** 选中的房间消失后，选中项回到 Create Room

#### Scenario: Leave online menu
- **WHEN** 玩家在联机菜单按 Esc
- **THEN** 系统应停止查找房间并回到欢迎界面

### Requirement: Room Creation
系统 SHALL 允许玩家在本机创建房间，监听配置的端口，并响应局域网内的房间查询。

#### Scenario: Create room successfully
- **WHEN** 玩家在联机菜单选择 Create Room
- **THEN** 系统应在 `online_port` 端口开启 TCP 监听，并在同号 UDP 端口上响应 discover 查询
- **AND** 让房主以 player_id 1 加入自己的房间
- **AND** 进入房间大厅，显示本机局域网 IP 和端口

#### Scenario: Port unavailable
- **WHEN** 玩家选择 Create Room，但 TCP 或 UDP 端口已被占用、无法监听
- **THEN** 系统应停留在联机菜单，并显示包含端口号的错误提示

## REMOVED Requirements

### Requirement: Room Joining
**Reason**: 房间号和手动输入 IP 都已取消，加入房间的唯一方式是从局域网房间列表选择，原需求中的加入表单和 bad_room 拒绝不再适用。
**Migration**: 由新增的 “Room Joining from List” 需求替代；房间满员、对局中、连接失败的场景保留在新需求中，并新增协议版本不一致的场景。
