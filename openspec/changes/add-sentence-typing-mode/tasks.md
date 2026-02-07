## 1. 数据准备
- [ ] 1.1 创建 `data/sentences.txt` 示例文件，包含至少 50 条英文句子

## 2. 配置扩展
- [ ] 2.1 在 `pkg/config/config.go` 中添加 `SentenceDictPath` 字段
- [ ] 2.2 更新 `DefaultConfig()` 设置默认句子文件路径为 `data/sentences.txt`

## 3. 游戏模式抽象
- [ ] 3.1 在 `pkg/game/game.go` 中定义 `GameMode` 枚举类型（Classic, Sentence）
- [ ] 3.2 在 `Game` 结构体中添加 `Mode` 字段
- [ ] 3.3 创建句子模式相关字段：`TargetSentence string`, `UserInput string`
- [ ] 3.4 实现句子数据加载函数 `LoadSentences(path string) ([]string, error)`
- [ ] 3.5 实现 `StartSentenceMode()` 方法：随机选择句子并初始化游戏状态

## 4. 输入处理扩展
- [ ] 4.1 修改 `AddChar()` 方法，支持句子模式下的字符添加（字母、数字、标点、空格）
- [ ] 4.2 修改 `Backspace()` 方法，支持句子模式下的字符删除
- [ ] 4.3 修改 `TryEliminate()` 方法，在句子模式下检查输入长度是否达到目标长度，然后结束游戏
- [ ] 4.4 更新统计逻辑，在句子模式下按字符匹配计算正确率

## 5. UI 渲染
- [ ] 5.1 在欢迎页面选择 "start" 后，显示模式选择菜单（Classic Mode / Sentence Mode）
- [ ] 5.2 创建 `RenderModeSelection()` 函数渲染模式选择界面
- [ ] 5.3 创建 `RenderSentenceGame()` 函数渲染句子模式游戏界面：
  - [ ] 5.3.1 显示目标句子
  - [ ] 5.3.2 显示用户输入（绿色=正确，红色=错误）
  - [ ] 5.3.3 显示统计信息（总字符数、正确字符数、错误字符数、准确率等）
- [ ] 5.4 修改 `cmd/word-killer/main.go` 的 `handleKey()` 添加模式选择逻辑

## 6. 主程序集成
- [ ] 6.1 在 `model` 结构体中添加 `modeSelectionIndex` 字段和 `modeSelected` 标志
- [ ] 6.2 在 `Update()` 中添加模式选择的状态处理逻辑
- [ ] 6.3 根据选择的模式调用 `game.Start()` 或 `game.StartSentenceMode()`
- [ ] 6.4 在 `View()` 中根据游戏模式调用相应的渲染函数

## 7. 测试
- [ ] 7.1 手动测试句子数据加载
- [ ] 7.2 手动测试模式选择流程
- [ ] 7.3 手动测试句子模式游戏流程（输入、删除、完成）
- [ ] 7.4 手动测试错误字符显示（红色）和正确字符显示（绿色）
- [ ] 7.5 手动测试统计数据准确性
- [ ] 7.6 手动测试两种模式的切换

## 8. 文档
- [ ] 8.1 更新 README（如果存在）说明新增的句子模式
- [ ] 8.2 在 `data/sentences.txt` 中添加注释说明文件格式（每行一个句子）
