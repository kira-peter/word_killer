## 1. Configuration
- [ ] 1.1 在 Config 结构体中添加难度比例字段 (ShortRatio, MediumRatio, LongRatio)
- [ ] 1.2 在 DefaultConfig 中设置合理的默认比例值 (如 30:50:20)
- [ ] 1.3 实现比例归一化方法,确保总和为 100%
- [ ] 1.4 添加配置验证,确保所有比例值 > 0

## 2. Game Core Logic
- [ ] 2.1 修改 Game 结构体,将单个 wordPool 改为三个词库 (shortPool, mediumPool, longPool)
- [ ] 2.2 修改 LoadWordDict 方法,支持同时加载三个难度文件
- [ ] 2.3 修改 generateWords 方法,按比例从三个词库中随机选择单词
- [ ] 2.4 确保单词不重复选择 (即使跨词库也不能重复)

## 3. Configuration File
- [ ] 3.1 更新 config.json 添加难度比例配置示例

## 4. Testing
- [ ] 4.1 测试比例归一化功能 (1:2:1 → 25:50:25)
- [ ] 4.2 测试边界情况 (某个比例为 0 时的行为)
- [ ] 4.3 测试单词生成是否符合设定比例
- [ ] 4.4 测试跨词库的单词去重功能
- [ ] 4.5 手动游戏测试,验证不同难度配置的实际体验
