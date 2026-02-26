// @Author daixk 2025/11/6 14:52:00
package dtoken_gf

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRenewPool_Resize(t *testing.T) {
	pool, err := NewRenewPoolManagerWithConfig(&RenewPoolConfig{
		MinSize:       2,
		MaxSize:       200,
		ScaleUpRate:   0.6,
		ScaleDownRate: 0.3,
		CheckInterval: 1 * time.Second,
		Expiry:        5 * time.Second,
		PreAlloc:      false,
		NonBlocking:   true,
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("=== 🚀 RenewPool 动态扩缩容演示（速率可控版） ===")

	go func() {
		for {
			// ===============================
			// 高负载阶段（任务提交很快）
			// ===============================
			fmt.Println("\n>>> 🔥 高负载阶段（提交频率极高）")
			for i := 0; i < 10000; i++ {
				_ = pool.Submit(func() {
					time.Sleep(300 * time.Millisecond) // 模拟重任务
				})
				// 高负载时快速提交（每 0.5ms 一次）
				time.Sleep(500 * time.Microsecond)
			}
			time.Sleep(10 * time.Second) // 稍等观察扩容稳定态

			// ===============================
			// 中负载阶段（任务量适中）
			// ===============================
			fmt.Println("\n>>> ⚙️ 中负载阶段（提交频率适中）")
			for i := 0; i < 3000; i++ {
				_ = pool.Submit(func() {
					time.Sleep(200 * time.Millisecond)
				})
				time.Sleep(2 * time.Millisecond)
			}
			time.Sleep(8 * time.Second)

			// ===============================
			// 低负载阶段（任务少 + 提交慢）
			// ===============================
			fmt.Println("\n>>> 🧊 低负载阶段（任务稀疏，容易触发缩容）")
			for i := 0; i < 200; i++ {
				_ = pool.Submit(func() {
					time.Sleep(100 * time.Millisecond)
				})
				// 每 20ms 提交一次 → 任务极少
				time.Sleep(20 * time.Millisecond)
			}
			// 等一会，让池空闲触发缩容
			time.Sleep(15 * time.Second)
		}
	}()

	// ===============================
	// 状态监控打印 + 彩色进度条
	// ===============================
	for {
		r, c, usage := pool.Stats()

		// 绘制条形图
		barLen := int(usage * 30)
		if barLen > 30 {
			barLen = 30
		}
		bar := fmt.Sprintf("[%s%s]", strings.Repeat("█", barLen), strings.Repeat(" ", 30-barLen))

		// 彩色区分负载
		color := ""
		reset := "\033[0m"
		switch {
		case usage >= 0.9:
			color = "\033[31m" // 红色 高负载
		case usage >= 0.6:
			color = "\033[33m" // 黄色 中负载
		default:
			color = "\033[32m" // 绿色 低负载
		}

		fmt.Printf("%s[池状态] Running=%-4d | Capacity=%-4d | Usage=%5.1f%% %s%s\n",
			color, r, c, usage*100, bar, reset)

		time.Sleep(1 * time.Second)
	}
}
