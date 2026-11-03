## ADDED Requirements

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

## MODIFIED Requirements

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
