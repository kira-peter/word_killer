# online-sentence-race Specification

## Purpose
Defines the Sentence Race battle: a whole sentence is shown at a time, the first player to type it exactly wins it, and the score equals the number of letters in the sentence.

## Requirements

### Requirement: Sentence Race Round
系统 SHALL 提供整句竞速玩法：一次显示一整句，最先完整打对的玩家得到这一句。

#### Scenario: Prepare sentences
- **WHEN** 房主开始 Sentence Race 对局
- **THEN** 服务端应从句子库中取出与设置个数相同（3–10 句，默认 5 句）的互不相同的句子
- **AND** 3 秒倒计时结束后，向所有玩家广播第 1 句

#### Scenario: Next sentence
- **WHEN** 当前句子被某名玩家拿到
- **THEN** 服务端应在 1 秒后广播下一句，并清空所有玩家的输入框
- **AND** 直到全部句子结束，或者整局时间上限（60–300 秒，默认 180 秒）到达

### Requirement: Sentence Typing and Claim
系统 SHALL 让玩家完整输入句子，由服务端裁决归属并按字母数计分。

#### Scenario: Type a sentence
- **WHEN** 玩家在整句玩法中输入
- **THEN** 客户端应接受所有可打印字符（包括空格和标点），Backspace 删除最后一个字符
- **AND** 逐字标出与目标句子一致和不一致的部分

#### Scenario: Correct sentence
- **WHEN** 玩家的输入与当前句子完全一致并按回车
- **THEN** 客户端应发送 claim（index、句子）
- **AND** 如果这一句还没有被拿走，服务端应把它判给这名玩家，分数等于句中的字母数（不计空格和标点）
- **AND** 服务端广播 word_claimed，界面显示是谁拿到以及得分

#### Scenario: Incorrect sentence on Enter
- **WHEN** 玩家按回车时，输入与当前句子不一致
- **THEN** 客户端不应清空输入，也不发送任何消息
- **AND** 提示 “Not yet - fix the red part”，玩家可以退格修改

#### Scenario: Simultaneous sentence claims
- **WHEN** 两名玩家先后对同一句发送了正确的 claim
- **THEN** 服务端只认最先到达的那个 claim，之后的被忽略
