package game

import (
	"math/rand"
	"time"
)

// getFishASCII 根据大小获取小鱼ASCII（内部函数避免循环导入）
func getFishASCII(size int) string {
	switch size {
	case 1:
		return "><>"
	case 2:
		return "><(°>"
	case 3:
		return "><((°>"
	default:
		return "><>"
	}
}

// Fish 代表一条带单词的小鱼
type Fish struct {
	Word        string    // 要输入的单词
	X           float64   // 水平位置 (0.0-1.0 归一化)
	Y           int       // 垂直行号 (0-9)
	Speed       float64   // 游动速度 (0.005-0.015)
	Direction   int       // 方向：1右，-1左
	Size        int       // 大小 (1-3) 基于单词长度
	Completed   bool      // 是否完成输入
	CompletedAt time.Time // 完成时间（用于发光动画）
	Glowing     bool      // 是否处于发光状态
}

// UnderwaterState 海底模式状态
type UnderwaterState struct {
	Fishes           []Fish
	CountdownStart   time.Time      // 倒计时开始时间
	BackgroundFrame  int            // 背景动画帧
	SeaweedPositions []int          // 海藻X位置 (5列)
	BubbleStreams    []BubbleStream // 气泡流
}

// BubbleStream 上升的气泡流
type BubbleStream struct {
	X      int     // X位置
	Y      float64 // Y位置（浮点数实现平滑移动）
	Speed  float64 // 上升速度 (~0.1)
	Active bool    // 是否激活
}

// GenerateFishes 生成指定数量的小鱼（优化分布避免重叠）
func (g *Game) GenerateFishes(count int) []Fish {
	fishes := make([]Fish, 0, count)
	words := g.GetAvailableWords()

	// 创建占用网格来避免重叠（72列×16行，留出顶部和底部各2行）
	const fishRows = 16
	occupied := make([][]bool, fishRows)
	for i := range occupied {
		occupied[i] = make([]bool, 72)
	}

	// 确保小鱼在不同高度均匀分布
	rowCounts := make([]int, fishRows) // 统计每行的小鱼数量

	for i := 0; i < count; i++ {
		if len(words) == 0 {
			break
		}

		// 随机选择一个单词
		word := words[rand.Intn(len(words))]

		// 根据单词长度确定小鱼大小
		var size int
		wordLen := len(word)
		if wordLen <= 5 {
			size = 1 // 小
		} else if wordLen <= 10 {
			size = 2 // 中
		} else {
			size = 3 // 大
		}

		// 计算小鱼+单词所需的总宽度（>word< 格式）
		totalWidth := len(word) + 2 + 6 // 单词+2个括号+6格缓冲

		// 尝试找到一个不重叠的位置（最多尝试100次）
		var fish Fish
		placed := false
		for attempt := 0; attempt < 100; attempt++ {
			// 优先选择小鱼较少的行（均匀分布）
			var y int
			if attempt < 50 {
				// 前50次尝试：选择最少小鱼的行
				minCount := rowCounts[0]
				y = 0
				for row := 1; row < fishRows; row++ {
					if rowCounts[row] < minCount {
						minCount = rowCounts[row]
						y = row
					}
				}
				// 在该行附近随机偏移±1行
				if rand.Float64() < 0.3 && y > 0 {
					y--
				} else if rand.Float64() < 0.3 && y < fishRows-1 {
					y++
				}
			} else {
				// 后50次尝试：完全随机
				y = rand.Intn(fishRows)
			}

			// Y坐标从2开始（顶部2行留给波浪）
			actualY := y + 2

			x := rand.Float64()
			xPos := int(x * 72)

			// 检查是否与已有小鱼重叠
			if canPlaceFish(occupied, xPos, y, totalWidth) {
				// 标记占用区域
				markOccupied(occupied, xPos, y, totalWidth)
				rowCounts[y]++

				fish = Fish{
					Word:      word,
					X:         x,
					Y:         actualY,
					Speed:     0.003 + rand.Float64()*0.007, // 减慢速度：0.003-0.010
					Direction: []int{-1, 1}[rand.Intn(2)],
					Size:      size,
					Completed: false,
					Glowing:   false,
				}
				placed = true
				break
			}
		}

		// 如果找到了合适位置就添加
		if placed {
			fishes = append(fishes, fish)
		}
	}

	return fishes
}

// canPlaceFish 检查是否可以在指定位置放置小鱼
func canPlaceFish(occupied [][]bool, x, y, width int) bool {
	// 只检查当前行，因为单词在一行内
	if y < 0 || y >= len(occupied) {
		return false
	}

	// 检查宽度范围
	for dx := 0; dx < width; dx++ {
		checkX := x + dx
		if checkX < 0 || checkX >= len(occupied[0]) {
			return false
		}
		if occupied[y][checkX] {
			return false
		}
	}
	return true
}

// markOccupied 标记占用区域
func markOccupied(occupied [][]bool, x, y, width int) {
	// 只标记当前行
	if y < 0 || y >= len(occupied) {
		return
	}

	for dx := 0; dx < width; dx++ {
		checkX := x + dx
		if checkX >= 0 && checkX < len(occupied[0]) {
			occupied[y][checkX] = true
		}
	}
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// UpdateFishPositions 更新所有小鱼的位置
func (g *Game) UpdateFishPositions() {
	if g.UnderwaterState == nil {
		return
	}

	// 收集需要移除的小鱼索引（发光动画结束的）
	toRemove := []int{}

	for i := range g.UnderwaterState.Fishes {
		fish := &g.UnderwaterState.Fishes[i]

		// 检查发光动画是否结束（800ms后）
		if fish.Completed && fish.Glowing {
			if time.Since(fish.CompletedAt).Milliseconds() > 800 {
				fish.Glowing = false // 结束发光状态
				toRemove = append(toRemove, i)
			}
			continue // 发光中不移动
		}

		// 已完成但不发光的小鱼不再渲染，也不需要移动
		if fish.Completed {
			continue
		}

		// 水平移动
		fish.X += fish.Speed * float64(fish.Direction)

		// 边界环绕（修复：确保在0.0-1.0范围内）
		for fish.X > 1.0 {
			fish.X -= 1.0
		}
		for fish.X < 0.0 {
			fish.X += 1.0
		}

		// 轻微垂直振荡（正弦波）- 去掉，保持固定Y坐标
		// oscillation := math.Sin(fish.X*2*math.Pi) * 0.3
		// newY := int(float64(fish.Y) + oscillation)
		// if newY >= 0 && newY < 8 { // 限制在0-7行
		// 	fish.Y = newY
		// }
	}

	// 从后往前删除已完成的小鱼，避免索引错乱
	for i := len(toRemove) - 1; i >= 0; i-- {
		idx := toRemove[i]
		g.UnderwaterState.Fishes = append(
			g.UnderwaterState.Fishes[:idx],
			g.UnderwaterState.Fishes[idx+1:]...,
		)
	}

	// 补充新的小鱼，保持总数为10
	currentCount := len(g.UnderwaterState.Fishes)
	if currentCount < 10 {
		newFishes := g.GenerateFishes(10 - currentCount)
		g.UnderwaterState.Fishes = append(g.UnderwaterState.Fishes, newFishes...)
	}
}

// UpdateBackgroundAnimation 更新背景动画
func (g *Game) UpdateBackgroundAnimation() {
	if g.UnderwaterState == nil {
		return
	}

	g.UnderwaterState.BackgroundFrame++

	// 更新气泡流（上升）
	for i := range g.UnderwaterState.BubbleStreams {
		bubble := &g.UnderwaterState.BubbleStreams[i]
		bubble.Y -= bubble.Speed // 向上移动

		if bubble.Y < 0 {
			bubble.Y = 20.0 // 从底部重新开始（适配20行）
		}
	}
}

// generateSeaweedPositions 生成海藻位置
func generateSeaweedPositions() []int {
	positions := make([]int, 5)
	for i := 0; i < 5; i++ {
		positions[i] = 10 + i*15 // 均匀分布
	}
	return positions
}

// generateBubbleStreams 生成气泡流
func generateBubbleStreams() []BubbleStream {
	streams := make([]BubbleStream, 8)
	for i := 0; i < 8; i++ {
		streams[i] = BubbleStream{
			X:      5 + i*9,                        // 均匀分布
			Y:      rand.Float64() * 20,            // 随机初始高度（适配20行）
			Speed:  0.08 + rand.Float64()*0.04,     // 0.08-0.12
			Active: true,
		}
	}
	return streams
}
