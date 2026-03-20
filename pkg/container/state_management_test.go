package container

import (
	"sync"
	"testing"
)

func TestContainerStaleState(t *testing.T) {
	c := &Container{}

	// 初始状态应该是 false
	if c.IsStale() {
		t.Error("初始状态应该是 false")
	}

	// 设置为 true
	c.SetStale(true)
	if !c.IsStale() {
		t.Error("设置后应该是 true")
	}

	// 设置为 false
	c.SetStale(false)
	if c.IsStale() {
		t.Error("重置后应该是 false")
	}
}

func TestContainerMarkedForUpdateState(t *testing.T) {
	c := &Container{}

	// 初始状态应该是 false
	if c.IsMarkedForUpdate() {
		t.Error("初始状态应该是 false")
	}

	// 设置为 true
	c.SetMarkedForUpdate(true)
	if !c.IsMarkedForUpdate() {
		t.Error("设置后应该是 true")
	}

	// 设置为 false
	c.SetMarkedForUpdate(false)
	if c.IsMarkedForUpdate() {
		t.Error("重置后应该是 false")
	}
}

func TestContainerLinkedToRestartingState(t *testing.T) {
	c := &Container{}

	// 初始状态应该是 false
	if c.IsLinkedToRestarting() {
		t.Error("初始状态应该是 false")
	}

	// 设置为 true
	c.SetLinkedToRestarting(true)
	if !c.IsLinkedToRestarting() {
		t.Error("设置后应该是 true")
	}

	// 设置为 false
	c.SetLinkedToRestarting(false)
	if c.IsLinkedToRestarting() {
		t.Error("重置后应该是 false")
	}
}

func TestContainerStateIndependence(t *testing.T) {
	c := &Container{}

	// 设置 Stale
	c.SetStale(true)
	if !c.IsStale() {
		t.Error("Stale 应该是 true")
	}
	if c.IsMarkedForUpdate() {
		t.Error("MarkedForUpdate 应该是 false")
	}
	if c.IsLinkedToRestarting() {
		t.Error("LinkedToRestarting 应该是 false")
	}

	// 设置 MarkedForUpdate
	c.SetMarkedForUpdate(true)
	if !c.IsStale() {
		t.Error("Stale 应该保持 true")
	}
	if !c.IsMarkedForUpdate() {
		t.Error("MarkedForUpdate 应该是 true")
	}
	if c.IsLinkedToRestarting() {
		t.Error("LinkedToRestarting 应该是 false")
	}

	// 设置 LinkedToRestarting
	c.SetLinkedToRestarting(true)
	if !c.IsStale() {
		t.Error("Stale 应该保持 true")
	}
	if !c.IsMarkedForUpdate() {
		t.Error("MarkedForUpdate 应该保持 true")
	}
	if !c.IsLinkedToRestarting() {
		t.Error("LinkedToRestarting 应该是 true")
	}
}

func TestContainerToRestart(t *testing.T) {
	c := &Container{}

	// 初始状态应该返回 false
	if c.ToRestart() {
		t.Error("初始状态应该返回 false")
	}

	// 当 MarkedForUpdate 为 true 时应该返回 true
	c.SetMarkedForUpdate(true)
	if !c.ToRestart() {
		t.Error("MarkedForUpdate 为 true 时应该返回 true")
	}

	// 重置并测试 LinkedToRestarting
	c.SetMarkedForUpdate(false)
	c.SetLinkedToRestarting(true)
	if !c.ToRestart() {
		t.Error("LinkedToRestarting 为 true 时应该返回 true")
	}

	// 当两个字段都为 true 时应该返回 true
	c.SetMarkedForUpdate(true)
	if !c.ToRestart() {
		t.Error("两个字段都为 true 时应该返回 true")
	}

	// 当两个字段都为 false 时应该返回 false
	c.SetMarkedForUpdate(false)
	c.SetLinkedToRestarting(false)
	if c.ToRestart() {
		t.Error("两个字段都为 false 时应该返回 false")
	}
}

func TestContainerStateConcurrency(t *testing.T) {
	c := &Container{}
	var wg sync.WaitGroup

	// 并发设置 Stale
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetStale(i%2 == 0)
		}()
	}
	wg.Wait()

	// 并发设置 MarkedForUpdate
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetMarkedForUpdate(i%2 == 0)
		}()
	}
	wg.Wait()

	// 并发设置 LinkedToRestarting
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetLinkedToRestarting(i%2 == 0)
		}()
	}
	wg.Wait()

	// 并发设置所有字段
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetStale(i%2 == 0)
			c.SetMarkedForUpdate(i%3 == 0)
			c.SetLinkedToRestarting(i%5 == 0)
		}()
	}
	wg.Wait()

	// 验证不会崩溃
	_ = c.IsStale()
	_ = c.IsMarkedForUpdate()
	_ = c.IsLinkedToRestarting()
	_ = c.ToRestart()
}

func TestContainerStateTransitions(t *testing.T) {
	c := &Container{}

	// 模拟完整的更新流程
	if c.IsStale() || c.IsMarkedForUpdate() || c.ToRestart() {
		t.Error("初始状态应该全部为 false")
	}

	// 检测到新镜像
	c.SetStale(true)
	if !c.IsStale() {
		t.Error("Stale 应该是 true")
	}
	if c.IsMarkedForUpdate() {
		t.Error("MarkedForUpdate 应该是 false")
	}
	if c.ToRestart() {
		t.Error("ToRestart 应该是 false")
	}

	// 标记为需要更新
	c.SetMarkedForUpdate(true)
	if !c.IsStale() || !c.IsMarkedForUpdate() {
		t.Error("Stale 和 MarkedForUpdate 都应该是 true")
	}
	if !c.ToRestart() {
		t.Error("ToRestart 应该是 true")
	}

	// 更新完成
	c.SetMarkedForUpdate(false)
	c.SetStale(false)
	if c.IsStale() || c.IsMarkedForUpdate() || c.ToRestart() {
		t.Error("所有状态应该重置为 false")
	}
}

func TestContainerDependencyScenario(t *testing.T) {
	dependent := &Container{}
	parent := &Container{}

	// 父容器被标记为需要重启
	parent.SetMarkedForUpdate(true)
	if !parent.ToRestart() {
		t.Error("父容器应该需要重启")
	}

	// 依赖容器标记为链接到正在重启的容器
	dependent.SetLinkedToRestarting(true)
	if !dependent.ToRestart() {
		t.Error("依赖容器应该需要重启")
	}

	// 父容器重启完成
	parent.SetMarkedForUpdate(false)
	if parent.ToRestart() {
		t.Error("父容器不应该需要重启")
	}

	// 依赖容器重启完成
	dependent.SetLinkedToRestarting(false)
	if dependent.ToRestart() {
		t.Error("依赖容器不应该需要重启")
	}
}