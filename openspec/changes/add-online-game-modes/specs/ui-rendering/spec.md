## MODIFIED Requirements

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
- **THEN** 内容区分为左右两半：左半部分竖排显示菜单（玩法、该玩法的各项设置的名称和当前值、Start Game），房主选中的一项高亮并带 “>” 标记
- **AND** 右半部分每行一名玩家，每名玩家显示在一个圆角边框的卡片中，名字用该玩家的颜色，并标记 host 和 you；本人的卡片边框使用自己的颜色
- **AND** 顶部标题显示房主 IP:端口和人数（例如 Players 3/4）
- **AND** 房主看到的底部提示为 “[↑↓] Select │ [Enter] Open │ [K] Kick │ [Tab] Rename │ [ESC] Close”
- **AND** 非房主看到 “Waiting for host...” 和 “[Tab] Rename │ [ESC] Leave”

#### Scenario: Setting wheel rendering
- **WHEN** 房主选中某一项并按回车
- **THEN** 界面保持大厅布局不变，看起来像在原视图上操作：该项的名称留在原来的行和列，其他菜单项隐藏（位置留空），右半部分的玩家卡片照常显示
- **AND** 滚轮出现在该项取值的位置，当前值与名称同一行、与原取值同一列，高亮加粗并用 ▶ ◀ 标出
- **AND** 上下各显示两个相邻档位，离当前值越远颜色越淡；靠近两端时空出的位置留白
- **AND** 底部提示为 “[↑↓] Scroll │ [Enter] Confirm │ [ESC] Cancel”

#### Scenario: Kick picker rendering
- **WHEN** 房主打开踢人选择框
- **THEN** 标题变为 “Kick which player?”，右半部分的玩家卡片中，当前选中的玩家卡片边框变红并在名字前显示 ✕，房主自己的卡片变灰
- **AND** 底部提示为 “[↑↓] Select │ [Enter] Kick │ [ESC] Cancel”

#### Scenario: Word drop battle rendering
- **WHEN** 处于单词掉落对战中
- **THEN** 标题显示当前单词进度（例如 Word 3/10）和剩余时间，内容区顶部一行显示所有玩家的分数
- **AND** 掉落区按掉落进度在对应的行显示当前单词，并将与本人输入匹配的前缀高亮
- **AND** 掉落区下方再显示一次当前单词作为对齐基准，其下每名玩家一行（最多 4 行），输入与单词逐字对齐（规则见 Live input rows）
- **AND** 单词被拿走或作废时，在掉落区短暂显示结果提示

#### Scenario: Sentence race rendering
- **WHEN** 处于整句竞速中
- **THEN** 标题显示当前句子进度（例如 Sentence 2/5）和剩余时间，内容区顶部一行显示所有玩家的分数
- **AND** 句子按单词折成最多 2 行；每一行句子下面紧跟每名玩家对应这一行的输入，与句子逐字对齐（规则见 Live input rows），不显示进度条或百分比
- **AND** 名字超过 5 个字符时截断为 “前 3 个字符 + * + 后 2 个字符”，例如 AVeryLongNam 显示为 AVe*am
- **AND** 本人的光标显示在下一个字符所在的那一行
- **AND** 一句被拿走时，显示是谁拿到以及得分

#### Scenario: Word pool rendering
- **WHEN** 处于单词池对战中
- **THEN** 标题显示剩余时间和是否允许掠夺
- **AND** 上方以 6 行 × 6 列的网格显示盘面：空闲单词为白色，被占有的单词显示为占有者的颜色，与本人输入匹配的单词前缀高亮
- **AND** 下方每名玩家一行（最多 4 行），先显示该玩家正在输入的目标单词（盘面上与其输入开头相同部分最长的单词），再显示 “→” 和输入，按目标单词逐字比对（规则见 Live input rows）；还没输入时只显示 “...”
- **AND** 本模式中，其他玩家打对的字母用该玩家自己的颜色，而不是灰色

#### Scenario: Live input rows
- **WHEN** 对战中显示每名玩家的实时输入
- **THEN** 每行显示该玩家的昵称（本人显示为 You，昵称用该玩家的颜色）和当前输入
- **AND** 输入与目标逐字对齐，第 N 个字符显示在目标第 N 个字符的正下方
- **AND** 逐字比对：与上方目标字符不同的字符单独显示为红色，之后打对的字符照常显示；打错的空格显示为红色的 “·”
- **AND** 其他玩家打对的字符为灰色，本人打对的字符为绿色，本人带光标
- **AND** 本人的一行固定在最下面，其他玩家按进入房间的顺序排在上方；人数不足 4 人时空行留在顶部
- **AND** 其他玩家还没输入时显示 “...”

#### Scenario: Player colors
- **WHEN** 为玩家分配颜色
- **THEN** 4 种玩家颜色依次为蓝、黄、紫、青，不使用红色和绿色，避免与对错标记混淆

#### Scenario: Battle result rendering
- **WHEN** 显示对战结算
- **THEN** 界面应按排名列出玩家的昵称、分数、拿到的单词或句子数，并标记胜者
- **AND** 房主的提示为 “[Enter] Back to Lobby │ [ESC] Close Room”，非房主的提示为 “Waiting for host... │ [ESC] Leave”
