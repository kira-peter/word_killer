# online-room Specification

## Purpose
Defines online rooms: creating one, discovering and joining LAN rooms from a list, the lobby where the host picks a mode and starts, and handling disconnects.

## Requirements

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

### Requirement: Kick Player
系统 SHALL 允许房主在大厅中把其他玩家踢出房间。

#### Scenario: Open kick picker
- **WHEN** 房主在大厅按 K
- **THEN** 系统应打开踢人选择框，列出除房主以外的玩家，上下键选择，Esc 取消
- **AND** 房间里只有房主一人时，提示 “No one to kick”，不打开选择框

#### Scenario: Host kicks a player
- **WHEN** 房主在踢人选择框中选中一名玩家并按回车
- **THEN** 服务端应向该玩家发送 kicked 并断开他的连接
- **AND** 将其移出玩家列表，并向其余玩家广播新的玩家列表

#### Scenario: Kicked player returns to menu
- **WHEN** 玩家收到 kicked
- **THEN** 客户端应回到联机菜单，显示 “You were kicked”，并重新开始查找房间
- **AND** 之后可以从列表中再次加入该房间

#### Scenario: Invalid kick ignored
- **WHEN** 非房主发送 kick、房主试图踢自己，或者不在大厅阶段时发送 kick
- **THEN** 服务端应忽略该请求

### Requirement: Live Typing Sync
系统 SHALL 在对战中把每名玩家的当前输入实时同步给其他玩家。

#### Scenario: Relay typing
- **WHEN** 对战进行中，玩家的输入发生变化（输入、退格或清空）
- **THEN** 客户端应在下一帧（约 33ms）发送 typing（当前输入），同一帧内多次变化只发送最后一次，输入未变化时不发送
- **AND** 服务端应把它转发给其他玩家（不回发给本人），附上该玩家的 ID

#### Scenario: Relay limits
- **WHEN** 服务端收到 typing
- **THEN** 只在对局进行中转发，去掉控制字符，并截断到 200 个字符

#### Scenario: Reset on a new target
- **WHEN** 当前单词或句子被拿走、作废，或者出现新的单词或句子
- **THEN** 客户端应清空所有玩家显示的输入

### Requirement: Lobby Game Settings
系统 SHALL 让房主在大厅菜单中选择玩法或设置项，并在滚轮页面中从预设档位里选值，不需要打字输入；确认后所有玩家看到新设置。

#### Scenario: Lobby menu navigation
- **WHEN** 房主在大厅按上下键
- **THEN** 选中项在玩法、该玩法的各项设置和 Start Game 之间移动

#### Scenario: Open the wheel
- **WHEN** 房主选中玩法或某个设置项并按回车
- **THEN** 系统应在原大厅视图中就地调整：隐藏其他设置项，该项名称位置不变，滚轮出现在它的取值处

#### Scenario: Scroll the wheel
- **WHEN** 房主在滚轮页面按上下键
- **THEN** 当前值在预设档位之间滚动，到两端后不再变化
- **AND** 滚动只在本地预览，不广播

#### Scenario: Confirm or cancel
- **WHEN** 房主在滚轮页面按回车
- **THEN** 系统应发送新值，回到大厅菜单，服务端广播给所有玩家
- **AND** 按 Esc 时放弃本次修改，回到大厅菜单，设置保持不变

#### Scenario: Choose game mode
- **WHEN** 房主在玩法滚轮中确认了新的玩法
- **THEN** 玩法在 Word Drop、Sentence Race、Word Pool 之间切换
- **AND** 设置重置为新玩法的默认值，并广播给所有玩家

#### Scenario: Available settings
- **WHEN** 房主打开设置项的滚轮
- **THEN** Word Drop 和 Sentence Race 可以调整个数和整局时间上限
- **AND** Word Pool 可以调整游戏时长、每秒出词个数和是否允许掠夺

#### Scenario: Reject invalid settings
- **WHEN** 服务端收到不在预设档位内的设置值，或者来自非房主的设置请求
- **THEN** 服务端应忽略该请求，设置保持不变

#### Scenario: Settings carried into the game
- **WHEN** 房主开始游戏
- **THEN** game_start 应携带当前玩法和设置，所有玩家按这些设置进行对局

### Requirement: Room Lobby
系统 SHALL 提供房间大厅，显示玩家列表和当前设置，并由房主选择玩法、调整设置、踢人和开始对局。

#### Scenario: Lobby player list
- **WHEN** 有玩家加入或离开房间
- **THEN** 服务端应向所有玩家广播最新的玩家列表
- **AND** 大厅界面显示每个玩家的昵称和颜色，并标记房主

#### Scenario: Host selects mode
- **WHEN** 房主在大厅中切换玩法
- **THEN** 服务端应向所有玩家广播当前选中的玩法和设置
- **AND** 可选玩法为 Word Drop、Sentence Race、Word Pool

#### Scenario: Host starts game
- **WHEN** 房主选中 Start Game 并按回车，且房间内至少有 2 名玩家
- **THEN** 服务端应向所有玩家广播 game_start，包含 3 秒倒计时、玩法和设置
- **AND** 所有玩家进入对应玩法的对战界面

#### Scenario: Not enough players
- **WHEN** 房主选中 Start Game 并按回车，但房间内只有 1 名玩家
- **THEN** 系统不应开始对局，并提示 “Need at least 2 players”

#### Scenario: Non-host cannot control
- **WHEN** 非房主玩家在大厅按方向键、K 或回车
- **THEN** 系统应忽略这些操作，界面显示 “Waiting for host...”
- **AND** 非房主看到与房主相同的菜单和当前设置值，但没有选中标记

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
