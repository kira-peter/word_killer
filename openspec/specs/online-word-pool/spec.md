# online-word-pool Specification

## Purpose
Defines the Word Pool battle: a 36-slot board keeps receiving new words, players type a word to own it (and optionally steal it from others), and the score is the total letters of the words held when time runs out.

## Requirements

### Requirement: Word Pool Board
系统 SHALL 提供单词池玩法：盘面上持续出现新单词，玩家输入单词把它占为己有。

#### Scenario: Board layout
- **WHEN** Word Pool 对局开始
- **THEN** 盘面应有 36 个格子，开始时全部为空
- **AND** 3 秒倒计时结束后开始出词，并开始计算游戏时长（30–120 秒，默认 60 秒）

#### Scenario: Spawn words
- **WHEN** 对局进行中
- **THEN** 服务端应按设置每秒出 1–3 个新单词（默认 1 个），放进随机的空格子，并广播 pool_spawn（slot、word）
- **AND** 盘面上的单词互不重复

#### Scenario: Board full
- **WHEN** 出新词时没有空格子
- **THEN** 新单词应替换出现最早、且还没有人占有的单词
- **AND** 如果所有单词都已被占有，本次不出新词

### Requirement: Word Pool Claim
系统 SHALL 由服务端裁决单词的占有，并按房主的设置决定能否掠夺。

#### Scenario: Claim a free word
- **WHEN** 玩家输入与盘面上某个空闲单词一致并按回车
- **THEN** 客户端应发送 claim（slot、word）
- **AND** 服务端应把该单词判给这名玩家，并广播 pool_claimed（slot、player_id、from=0）
- **AND** 所有玩家的界面上，该单词变成这名玩家的颜色

#### Scenario: Steal a word
- **WHEN** 房主开启了掠夺，玩家输入了别人已占有的单词并按回车
- **THEN** 服务端应把该单词改判给这名玩家，并广播 pool_claimed，from 为原占有者

#### Scenario: Stealing disabled
- **WHEN** 房主关闭了掠夺，玩家提交了别人已占有的单词
- **THEN** 服务端应忽略该请求，单词归属不变

#### Scenario: Own word or no match
- **WHEN** 玩家按回车时，输入的是自己已占有的单词，或者盘面上没有可以占有的匹配单词
- **THEN** 客户端应清空输入，不发送任何消息

#### Scenario: Simultaneous pool claims
- **WHEN** 两名玩家几乎同时提交同一个空闲单词
- **THEN** 服务端按到达顺序处理：第一个占有它；关闭掠夺时第二个被忽略，开启掠夺时第二个把它抢走

### Requirement: Word Pool Scoring
系统 SHALL 在游戏时长结束时按每人手中的单词计分。

#### Scenario: Final score
- **WHEN** 游戏时长到
- **THEN** 每名玩家的分数应为其占有的所有单词的字母数之和，单词数为占有的单词个数
- **AND** 服务端广播 game_over，按分数排名

#### Scenario: Live score
- **WHEN** 有单词被占有或被掠夺
- **THEN** 所有玩家界面上的比分栏应实时显示每人当前手中单词的字母数之和
