package inject

import "testing"

// 测试深度注入的各种场景
func TestDeepInjectAdvanced(t *testing.T) {
	// 场景1: 多层嵌套的手动注入实例
	type Level3 struct {
		Value string
	}

	type Level2 struct {
		L3 *Level3 `inject:""`
	}

	type Level1 struct {
		L2 *Level2 `inject:""`
	}

	type Root struct {
		L1 *Level1 `inject:""`
	}

	var g Graph

	// 手动创建嵌套结构
	root := &Root{
		L1: &Level1{
			L2: &Level2{
				L3: &Level3{Value: "manually created"},
			},
		},
	}

	// 提供根对象
	if err := g.Provide(&Object{Value: root}); err != nil {
		t.Fatal("failed to provide root:", err)
	}

	// 执行深度注入
	if err := g.Populate(); err != nil {
		t.Fatal("failed to populate:", err)
	}

	// 验证深度注入是否成功
	if root.L1 == nil {
		t.Fatal("root.L1 should not be nil")
	}
	if root.L1.L2 == nil {
		t.Fatal("root.L1.L2 should not be nil")
	}
	if root.L1.L2.L3 == nil {
		t.Fatal("root.L1.L2.L3 should not be nil")
	}
	if root.L1.L2.L3.Value != "manually created" {
		t.Fatal("deep injected value should be preserved")
	}
}

func TestDeepInjectWithMixedProvision(t *testing.T) {
	// 场景2: 混合手动创建和依赖注入提供的实例
	type ServiceA struct {
		Name string
	}

	type ServiceB struct {
		A *ServiceA `inject:""`
	}

	type ServiceC struct {
		B *ServiceB `inject:""`
	}

	var g Graph

	// 手动提供ServiceA
	serviceA := &ServiceA{Name: "ServiceA"}
	if err := g.Provide(&Object{Value: serviceA}); err != nil {
		t.Fatal("failed to provide serviceA:", err)
	}

	// 手动创建包含部分依赖的ServiceC
	serviceC := &ServiceC{
		B: &ServiceB{}, // B没有A的依赖
	}

	if err := g.Provide(&Object{Value: serviceC}); err != nil {
		t.Fatal("failed to provide serviceC:", err)
	}

	// 执行注入
	if err := g.Populate(); err != nil {
		t.Fatal("failed to populate:", err)
	}

	// 验证混合注入结果
	if serviceC.B == nil {
		t.Fatal("serviceC.B should not be nil")
	}
	if serviceC.B.A == nil {
		t.Fatal("serviceC.B.A should be injected")
	}
	if serviceC.B.A != serviceA {
		t.Fatal("serviceC.B.A should be the same instance as serviceA")
	}
	if serviceC.B.A.Name != "ServiceA" {
		t.Fatal("injected service should preserve its properties")
	}
}

func TestDeepInjectCircularDependency(t *testing.T) {
	// 场景3: 测试循环依赖的处理（这应该能正常工作）
	type CircularB struct {
		A interface{} `inject:""`
	}

	type CircularA struct {
		B *CircularB `inject:""`
	}

	var g Graph

	// 手动创建一个带有循环依赖的结构
	a := &CircularA{}
	b := &CircularB{A: a}
	a.B = b

	// 提供到依赖图
	if err := g.Provide(&Object{Value: a}); err != nil {
		t.Fatal("failed to provide a:", err)
	}

	if err := g.Provide(&Object{Value: b}); err != nil {
		t.Fatal("failed to provide b:", err)
	}

	// 执行注入
	if err := g.Populate(); err != nil {
		t.Fatal("failed to populate:", err)
	}

	// 验证循环依赖保持完整
	if a.B != b {
		t.Fatal("circular dependency should be preserved")
	}
	if b.A.(*CircularA) != a {
		t.Fatal("circular dependency should be preserved")
	}
}
