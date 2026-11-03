## ADDED Requirements

### Requirement: Online Interface Rendering
系统 SHALL 为联机流程提供与现有风格一致的界面（80 字符宽、顶部标题、中间内容、底部操作提示），房间只能从局域网列表加入。

#### Scenario: Welcome menu with Online option
- **WHEN** 显示欢迎界面
- **THEN** 菜单选项应为 Start、Online、About

#### Scenario: Online menu rendering
- **WHEN** 显示联机菜单
- **THEN** 界面上方显示 Create Room 选项
- **AND** 下方的 “Rooms on your network” 区域列出发现的房间，每行显示房主昵称、人数（例如 2/4）、状态（Waiting / Full / In game）和 IP
- **AND** 对局中或已满的房间用灰色显示
- **AND** 没有房间时显示 “Searching for rooms...”
- **AND** 有错误或正在连接时，在列表下方显示提示
- **AND** 底部提示为 “[↑↓] Select │ [Enter] Confirm │ [ESC] Back”

#### Scenario: Lobby rendering
- **WHEN** 显示房间大厅
- **THEN** 界面应显示房主 IP:端口、玩家列表（最多 4 人，房主带标记）、当前玩法，不显示房间号
- **AND** 房主看到的底部提示为 “[↑↓] Mode │ [Tab] Rename │ [Enter] Start │ [ESC] Close Room”
- **AND** 非房主看到 “Waiting for host...” 和 “[Tab] Rename │ [ESC] Leave”

#### Scenario: Word drop battle rendering
- **WHEN** 处于单词掉落对战中
- **THEN** 顶部比分栏显示所有玩家的昵称和分数，以及当前单词进度（例如 Word 3/10）
- **AND** 中间掉落区按掉落进度在对应的行显示当前单词，并将与输入匹配的前缀高亮
- **AND** 底部显示本玩家的输入框
- **AND** 单词被拿走或作废时，在掉落区短暂显示结果提示

#### Scenario: Battle result rendering
- **WHEN** 显示对战结算
- **THEN** 界面应按排名列出玩家的昵称、分数、拿到的单词数，并标记胜者
- **AND** 房主的提示为 “[Enter] Back to Lobby │ [ESC] Close Room”，非房主的提示为 “Waiting for host... │ [ESC] Leave”

## REMOVED Requirements

### Requirement: Online Screens Rendering
**Reason**: 加入表单（IP 和房间号输入）已删除，联机菜单改为局域网房间列表，大厅不再显示房间号。
**Migration**: 由新增的 “Online Interface Rendering” 需求替代；欢迎菜单、大厅、对战、结算的场景保留，联机菜单改为房间列表，加入表单的场景删除。
