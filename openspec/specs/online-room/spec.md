# online-room Specification

## Purpose
定义联机房间的创建、加入、大厅和断线处理：玩家通过房主 IP 和 4 位房间号进入同一房间，由房主选择玩法并开始对局。

## Requirements

### Requirement: Online Menu Entry
系统 SHALL 在欢迎界面提供 Online 入口，进入后可以选择创建房间或加入房间。

#### Scenario: Enter online menu
- **WHEN** 玩家在欢迎界面选中 Online 并按回车
- **THEN** 系统应显示联机菜单，包含 Create Room、Join Room、Back 三个选项

#### Scenario: Leave online menu
- **WHEN** 玩家在联机菜单按 Esc 或选择 Back
- **THEN** 系统应回到欢迎界面

### Requirement: Room Creation
系统 SHALL 允许玩家在本机创建房间，监听配置的端口，并生成 4 位房间号。

#### Scenario: Create room successfully
- **WHEN** 玩家在联机菜单选择 Create Room
- **THEN** 系统应在 `online_port` 端口开启 TCP 监听
- **AND** 生成一个 0000–9999 之间的随机 4 位房间号
- **AND** 让房主以 player_id 1 加入自己的房间
- **AND** 进入房间大厅，显示房间号和本机局域网 IP

#### Scenario: Port unavailable
- **WHEN** 玩家选择 Create Room，但端口已被占用或无法监听
- **THEN** 系统应停留在联机菜单，并显示包含端口号的错误提示

### Requirement: Room Joining
系统 SHALL 允许玩家通过输入房主 IP 和房间号加入房间。

#### Scenario: Join form input
- **WHEN** 玩家在联机菜单选择 Join Room
- **THEN** 系统应显示加入表单，包含 IP 和房间号两个输入框
- **AND** Tab 或上下键在两个输入框之间切换，Backspace 删除字符
- **AND** 房间号输入框只接受数字，最多 4 位

#### Scenario: Join successfully
- **WHEN** 玩家填写 IP 和房间号后按回车
- **AND** 房间号正确、房间未满、且不在对局中
- **THEN** 系统应连接 `IP:online_port`，发送 join 消息
- **AND** 收到 join_accepted 后进入房间大厅

#### Scenario: Wrong room number
- **WHEN** 玩家提交的房间号与房主的不一致
- **THEN** 服务端应回复 join_rejected（reason=bad_room）并断开连接
- **AND** 客户端停留在加入表单，显示 “Room not found”

#### Scenario: Room full
- **WHEN** 房间内已有 4 名玩家时又有人加入
- **THEN** 服务端应回复 join_rejected（reason=room_full）
- **AND** 客户端显示 “Room is full”

#### Scenario: Room in game
- **WHEN** 房间正在对局时有人加入
- **THEN** 服务端应回复 join_rejected（reason=in_game）
- **AND** 客户端显示 “Game in progress”

#### Scenario: Connection failure
- **WHEN** 无法在超时时间（3 秒）内连接到 `IP:online_port`
- **THEN** 客户端应停留在加入表单，显示 “Cannot connect to host”

### Requirement: Room Lobby
系统 SHALL 提供房间大厅，显示玩家列表，并由房主选择玩法、开始对局。

#### Scenario: Lobby player list
- **WHEN** 有玩家加入或离开房间
- **THEN** 服务端应向所有玩家广播最新的玩家列表
- **AND** 大厅界面显示每个玩家的昵称，并标记房主

#### Scenario: Host selects mode
- **WHEN** 房主在大厅用上下键切换玩法
- **THEN** 服务端应向所有玩家广播当前选中的玩法
- **AND** 第一版只有 Word Drop 一种玩法

#### Scenario: Host starts game
- **WHEN** 房主按回车开始，且房间内至少有 2 名玩家
- **THEN** 服务端应向所有玩家广播 game_start，包含 3 秒倒计时
- **AND** 所有玩家进入对战界面

#### Scenario: Not enough players
- **WHEN** 房主按回车开始，但房间内只有 1 名玩家
- **THEN** 系统不应开始对局，并提示 “Need at least 2 players”

#### Scenario: Non-host cannot control
- **WHEN** 非房主玩家在大厅按上下键或回车
- **THEN** 系统应忽略这些操作，界面显示 “Waiting for host...”

#### Scenario: Rename in lobby
- **WHEN** 任意玩家在大厅按 Tab，编辑昵称后按回车
- **THEN** 服务端应更新该玩家的昵称（最多 12 个字符）并向所有玩家广播玩家列表
- **AND** 编辑时按 Esc 取消修改，并留在大厅；内容为空时按回车保留原昵称

#### Scenario: Rename interrupted by game start
- **WHEN** 玩家正在编辑昵称时房主开始了对局
- **THEN** 系统应放弃未保存的修改，并进入对战界面

### Requirement: Disconnection Handling
系统 SHALL 妥善处理玩家断线和房主退出。

#### Scenario: Player leaves lobby
- **WHEN** 非房主玩家在大厅按 Esc 或断开连接
- **THEN** 服务端应将其移出玩家列表并广播
- **AND** 该玩家回到联机菜单

#### Scenario: Host closes room
- **WHEN** 房主在大厅或对局中按 Esc 退出
- **THEN** 服务端应向所有玩家广播 host_closed 并关闭监听
- **AND** 其他玩家显示 “Host closed the room”，然后回到联机菜单

#### Scenario: Connection lost
- **WHEN** 客户端与服务端的连接意外断开
- **THEN** 客户端应显示 “Connection lost”，然后回到联机菜单
