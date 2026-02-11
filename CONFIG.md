# Word Killer 配置文件说明

配置文件位置：`config.json`

## 配置项详解

### 基础设置

#### `word_count`
- **类型**: 整数
- **默认值**: 50
- **说明**: 经典模式下的单词数量
- **建议值**: 20-100
- **示例**: `"word_count": 50`

---

### 词库路径

#### `short_dict_path`
- **类型**: 字符串
- **默认值**: `"data/google-10000-short.txt"`
- **说明**: 短单词词库文件路径（3-5个字母）
- **示例**: `"short_dict_path": "data/google-10000-short.txt"`

#### `medium_dict_path`
- **类型**: 字符串
- **默认值**: `"data/google-10000-medium.txt"`
- **说明**: 中等长度单词词库文件路径（6-8个字母）
- **示例**: `"medium_dict_path": "data/google-10000-medium.txt"`

#### `long_dict_path`
- **类型**: 字符串
- **默认值**: `"data/google-10000-long.txt"`
- **说明**: 长单词词库文件路径（9+个字母）
- **示例**: `"long_dict_path": "data/google-10000-long.txt"`

#### `sentence_dict_path`
- **类型**: 字符串
- **默认值**: `"data/sentences.txt"`
- **说明**: 句子模式词库文件路径
- **示例**: `"sentence_dict_path": "data/sentences.txt"`

---

### 难度配比

#### `short_ratio`
- **类型**: 数字
- **默认值**: 30
- **说明**: 短单词出现比例权重
- **计算**: 实际比例 = short_ratio / (short_ratio + medium_ratio + long_ratio)
- **建议**: 想增加短单词，可以提高此值；想增加难度，可以降低此值
- **示例**: `"short_ratio": 30` (30%)

#### `medium_ratio`
- **类型**: 数字
- **默认值**: 50
- **说明**: 中等长度单词出现比例权重
- **计算**: 实际比例 = medium_ratio / (short_ratio + medium_ratio + long_ratio)
- **示例**: `"medium_ratio": 50` (50%)

#### `long_ratio`
- **类型**: 数字
- **默认值**: 20
- **说明**: 长单词出现比例权重
- **计算**: 实际比例 = long_ratio / (short_ratio + medium_ratio + long_ratio)
- **建议**: 想增加挑战性，可以提高此值
- **示例**: `"long_ratio": 20` (20%)

**难度配比示例：**
- **简单模式**: `50:40:10` (更多短单词)
- **标准模式**: `30:50:20` (默认配置)
- **困难模式**: `10:40:50` (更多长单词)

---

### 倒计时模式设置

#### `countdown_duration`
- **类型**: 整数
- **单位**: 秒
- **默认值**: 60
- **说明**: 倒计时模式的总时长
- **建议值**:
  - 休闲：90-120秒
  - 标准：60秒
  - 挑战：30-45秒
- **示例**: `"countdown_duration": 60`

---

### 极速模式设置

#### `speedrun_word_count`
- **类型**: 整数
- **默认值**: 25
- **说明**: 极速模式需要完成的固定单词数量
- **建议值**:
  - 快速游戏：10-15个
  - 标准：25个
  - 马拉松：40-50个
- **示例**: `"speedrun_word_count": 25`
- **注意**: 完成时间会保存为最佳记录（保存在 `speedrun_record.json`）

---

### 节奏大师模式设置

#### `rhythm_initial_time_limit`
- **类型**: 浮点数
- **单位**: 秒
- **默认值**: 2.0
- **说明**: 节奏大师模式初始的每词时间限制
- **建议值**:
  - 新手：3.0-4.0秒
  - 标准：2.0秒
  - 专家：1.5秒
- **示例**: `"rhythm_initial_time_limit": 2.0`

#### `rhythm_min_time_limit`
- **类型**: 浮点数
- **单位**: 秒
- **默认值**: 0.5
- **说明**: 节奏大师模式的最小时间限制（难度上限）
- **建议值**:
  - 休闲：1.0秒
  - 标准：0.5秒
  - 极限：0.3秒
- **示例**: `"rhythm_min_time_limit": 0.5`
- **注意**: 达到此限制后难度不再增加

#### `rhythm_difficulty_step`
- **类型**: 浮点数
- **单位**: 秒
- **默认值**: 0.1
- **说明**: 每升一级时间限制减少的秒数
- **建议值**:
  - 渐进式：0.05秒（缓慢增加难度）
  - 标准：0.1秒
  - 急速升级：0.15-0.2秒（快速增加难度）
- **示例**: `"rhythm_difficulty_step": 0.1`

#### `rhythm_words_per_level`
- **类型**: 整数
- **默认值**: 10
- **说明**: 完成多少个单词后升一级（减少时间限制）
- **建议值**:
  - 频繁升级：5个词
  - 标准：10个词
  - 缓慢升级：15-20个词
- **示例**: `"rhythm_words_per_level": 10`

---

## 完整配置示例

### 标准配置（默认）
```json
{
  "word_count": 50,
  "short_dict_path": "data/google-10000-short.txt",
  "medium_dict_path": "data/google-10000-medium.txt",
  "long_dict_path": "data/google-10000-long.txt",
  "sentence_dict_path": "data/sentences.txt",
  "short_ratio": 30,
  "medium_ratio": 50,
  "long_ratio": 20,
  "countdown_duration": 60,
  "speedrun_word_count": 25,
  "rhythm_initial_time_limit": 2.0,
  "rhythm_min_time_limit": 0.5,
  "rhythm_difficulty_step": 0.1,
  "rhythm_words_per_level": 10
}
```

### 简单模式配置
适合新手或休闲玩家：
```json
{
  "word_count": 30,
  "short_ratio": 50,
  "medium_ratio": 40,
  "long_ratio": 10,
  "countdown_duration": 90,
  "speedrun_word_count": 15,
  "rhythm_initial_time_limit": 3.0,
  "rhythm_min_time_limit": 1.0,
  "rhythm_difficulty_step": 0.05,
  "rhythm_words_per_level": 15
}
```

### 困难模式配置
适合高手玩家：
```json
{
  "word_count": 100,
  "short_ratio": 10,
  "medium_ratio": 40,
  "long_ratio": 50,
  "countdown_duration": 30,
  "speedrun_word_count": 40,
  "rhythm_initial_time_limit": 1.5,
  "rhythm_min_time_limit": 0.3,
  "rhythm_difficulty_step": 0.15,
  "rhythm_words_per_level": 5
}
```

---

## 节奏大师难度计算公式

当前时间限制 = max(初始时间限制 - 难度等级 × 难度步长, 最小时间限制)

**示例**（使用默认配置）：
- 第1-10词：2.0秒/词（等级0）
- 第11-20词：1.9秒/词（等级1，2.0 - 0.1×1）
- 第21-30词：1.8秒/词（等级2，2.0 - 0.1×2）
- ...
- 第151+词：0.5秒/词（等级15+，已达最小值）

---

## 配置修改步骤

1. 用文本编辑器打开 `config.json`
2. 修改想要调整的参数值
3. 保存文件
4. 重新启动游戏（配置会在启动时加载）

## 注意事项

- ⚠️ 配置文件必须是有效的 JSON 格式
- ⚠️ 所有路径使用正斜杠 `/` 或双反斜杠 `\\`
- ⚠️ 数字类型不要加引号，字符串类型必须加引号
- ⚠️ 最后一项后面不要加逗号
- ⚠️ 如果配置文件损坏，游戏将使用内置的默认配置

## 恢复默认配置

如果配置出现问题，可以删除 `config.json` 文件，程序会自动使用内置的默认配置。

或者复制上面的"标准配置"内容覆盖 `config.json` 文件。
