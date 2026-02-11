## 1. 核心游戏逻辑实现 (pkg/game/game.go)

- [ ] 1.1 添加三个新的 GameMode 常量（ModeCountdown, ModeSpeedRun, ModeRhythmMaster）
- [ ] 1.2 扩展 Game 结构体，添加8个模式专属字段
- [ ] 1.3 实现 StartCountdownMode 函数（60秒倒计时初始化）
- [ ] 1.4 实现 StartSpeedRunMode 函数（25词固定初始化）
- [ ] 1.5 实现 StartRhythmMasterMode 函数（节奏模式初始化）
- [ ] 1.6 修改 AddChar 函数，添加倒计时和节奏模式的时间检查
- [ ] 1.7 修改 TryEliminate 函数，添加三个模式的完成后逻辑（动态生成、难度递增、游戏结束判定）
- [ ] 1.8 添加 CheckTimeouts 方法（100ms tick 调用，检查倒计时和节奏模式超时）

## 2. UI 渲染实现 (pkg/ui/styles.go)

- [ ] 2.1 实现 RenderCountdownGame 函数（倒计时器、时间警告、单词区域、输入区域）
- [ ] 2.2 实现 RenderSpeedRunGame 函数（毫秒计时器、进度显示、最佳记录显示）
- [ ] 2.3 实现 RenderRhythmMasterGame 函数（连击显示、等级显示、时间限制提示）
- [ ] 2.4 实现 renderRhythmWordArea 辅助函数（带进度条的单词渲染）
- [ ] 2.5 修改 renderModeSelectionContent 函数，将模式选项从3个扩展到5个

## 3. 主程序集成 (cmd/word-killer/main.go)

- [ ] 3.1 扩展 model 结构体，添加 speedRunBestTime 字段并修改 selectedMode 注释
- [ ] 3.2 修改 tickMsg 处理器，添加 CheckTimeouts 调用
- [ ] 3.3 扩展模式选择逻辑，支持5个模式的上下导航和启动
- [ ] 3.4 更新暂停菜单的"重新开始"逻辑，保持选中的模式
- [ ] 3.5 更新结果菜单的"重新开始"逻辑，保持选中的模式
- [ ] 3.6 扩展 View 函数，添加三个新模式的渲染分支
- [ ] 3.7 实现 loadSpeedRunBestTime 和 saveSpeedRunBestTime 函数（JSON 持久化）
- [ ] 3.8 在结果页面添加极速模式记录保存逻辑

## 4. 测试验证

- [ ] 4.1 测试倒计时模式：60秒倒计时、动态单词生成、时间警告、时间到自动结束
- [ ] 4.2 测试极速模式：25词固定、毫秒计时、完成自动结束、最佳记录保存和加载
- [ ] 4.3 测试节奏大师：单词计时、难度递增（每10词-0.1秒）、超时失败、进度条显示
- [ ] 4.4 测试模式切换：在5个模式间切换、暂停/恢复/重启功能
- [ ] 4.5 测试边界情况：快速按键、极端输入、暂停后的时间计算正确性
- [ ] 4.6 测试UI渲染：所有模式界面正确、颜色变化、进度条、毫秒显示格式

## 5. 文档更新

- [ ] 5.1 更新用户文档，说明三个新模式的玩法和规则
- [ ] 5.2 添加代码注释，解释模式特定的逻辑和时间检查机制
