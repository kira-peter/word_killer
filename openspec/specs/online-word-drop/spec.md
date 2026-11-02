# online-word-drop Specification

## Purpose
定义单词掉落对战的规则：10 个单词依次掉落，由服务端裁决归属，最先提交正确单词的玩家得分，最后按总分排名。

## Requirements

### Requirement: Word Drop Round Setup
系统 SHALL 在单词掉落对战开始时准备 10 个单词，并在倒计时结束后依次掉落。

#### Scenario: Prepare words
- **WHEN** 房主开始 Word Drop 对局
- **THEN** 服务端应从多难度词库中取出 10 个互不相同的单词
- **AND** 所有玩家的分数清零

#### Scenario: Countdown before first word
- **WHEN** 玩家收到 game_start
- **THEN** 对战界面应显示 3、2、1 倒计时
- **AND** 倒计时结束后，服务端下发第 1 个单词

### Requirement: Single Word Falling
系统 SHALL 一次只掉落一个单词，并在所有玩家的屏幕上同步显示。

#### Scenario: Word spawn
- **WHEN** 服务端下发 word_spawn（index、word、fall_ms）
- **THEN** 所有玩家的界面应在掉落区顶部显示该单词
- **AND** 清空每个玩家的输入框

#### Scenario: Word falling progress
- **WHEN** 单词处于掉落中
- **THEN** 客户端应根据 `已过时间 / fall_ms` 计算单词所在的行
- **AND** fall_ms 默认取 `online_drop_duration`（8 秒）

#### Scenario: Next word interval
- **WHEN** 当前单词被拿走或作废
- **THEN** 服务端应在 1 秒后下发下一个单词，直到 10 个单词全部结束

### Requirement: Word Claim Judgment
系统 SHALL 由服务端裁决单词归属，最先提交正确单词的玩家获得该单词。

#### Scenario: Correct claim
- **WHEN** 玩家输入的内容与当前单词完全一致并按回车
- **THEN** 客户端应发送 claim（index、word）
- **AND** 如果该单词还没有被拿走，服务端应把它判给这名玩家，加上与单词字母数相等的分数
- **AND** 服务端向所有玩家广播 word_claimed（index、player_id、points）

#### Scenario: Wrong input on Enter
- **WHEN** 玩家按回车时，输入与当前单词不一致
- **THEN** 客户端应清空输入，不发送任何消息，也不扣分

#### Scenario: Simultaneous claims
- **WHEN** 两名玩家先后对同一个单词发送了正确的 claim
- **THEN** 服务端只认最先到达的那个 claim
- **AND** 之后到达的 claim 被忽略

#### Scenario: Stale claim
- **WHEN** 服务端收到的 claim 中 index 不是当前单词，或者当前单词已经结束
- **THEN** 服务端应忽略这个 claim

#### Scenario: Claim feedback
- **WHEN** 玩家收到 word_claimed
- **THEN** 界面应显示是谁拿到了这个单词以及得分（例如 “Alice +5”），并更新比分栏

### Requirement: Word Expiration
系统 SHALL 让掉到底还没人拿到的单词作废。

#### Scenario: Word reaches bottom
- **WHEN** 单词下发后过了 fall_ms 仍然没有人拿到
- **THEN** 服务端应广播 word_expired（index）
- **AND** 该单词不计入任何玩家的分数，但计入 10 个单词的总数
- **AND** 界面显示 “Missed!” 提示

### Requirement: Battle Result
系统 SHALL 在 10 个单词结束后结算并显示排名。

#### Scenario: Game over ranking
- **WHEN** 10 个单词都已被拿走或作废
- **THEN** 服务端应广播 game_over，包含按分数从高到低排列的排名（昵称、分数、拿到的单词数）
- **AND** 分数最高的玩家为胜者；同分时并列获胜

#### Scenario: Early end by disconnection
- **WHEN** 对局中在线玩家少于 2 人
- **THEN** 服务端应立即广播 game_over，按当前分数结算

#### Scenario: Back to lobby
- **WHEN** 结算界面上房主按回车
- **THEN** 服务端应让所有玩家回到房间大厅，可以再来一局
- **AND** 非房主玩家按 Esc 离开房间，回到联机菜单
