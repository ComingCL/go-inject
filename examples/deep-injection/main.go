package main

import (
	"fmt"
	"log"

	"github.com/ComingCL/go-inject"
)

// Logger 接口
type Logger interface {
	Log(message string)
}

// ConsoleLogger 实现 Logger 接口
type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(message string) {
	fmt.Println("[LOG]", message)
}

// EmailService 服务 - 类似 TestForDeepInject 中的 TypeB
type EmailService struct{}

func (e *EmailService) SendEmail(to, subject, body string) {
	fmt.Printf("Email sent to %s: %s\n", to, subject)
}

// NotificationService 服务 - 类似 TestForDeepInject 中的 TypeC
type NotificationService struct {
	EmailService *EmailService `inject:""`
}

func (n *NotificationService) SendNotification(message string) {
	if n.EmailService != nil {
		n.EmailService.SendEmail("user@example.com", "Notification", message)
	}
}

// UserService 服务 - 类似 TestForDeepInject 中的 TypeA
type UserService struct {
	NotificationService *NotificationService `inject:""`
}

func (u *UserService) CreateUser(name string) {
	fmt.Printf("Creating user: %s\n", name)
	if u.NotificationService != nil {
		u.NotificationService.SendNotification(fmt.Sprintf("User %s has been created", name))
	}
}

// App 应用程序 - 类似 TestForDeepInject 中的 TypeD
type App struct {
	UserService *UserService `inject:""`
}

func main() {
	fmt.Println("=== Deep Injection Example ===")

	// 创建依赖图
	var g inject.Graph

	// 注册基础服务
	emailService := &EmailService{}
	notificationService := &NotificationService{}

	// 创建包含手动实例的 App - 这里模拟 TestForDeepInject 的模式
	app := &App{
		UserService: &UserService{}, // 手动创建的实例，依赖为空
	}

	// 按照 TestForDeepInject 的顺序注册服务
	if err := g.Provide(&inject.Object{Value: emailService}); err != nil {
		log.Fatal("Failed to provide emailService:", err)
	}

	if err := g.Provide(&inject.Object{Value: notificationService}); err != nil {
		log.Fatal("Failed to provide notificationService:", err)
	}

	if err := g.Provide(&inject.Object{Value: app}); err != nil {
		log.Fatal("Failed to provide app:", err)
	}

	// 执行依赖注入
	if err := g.Populate(); err != nil {
		log.Fatal("Failed to populate dependencies:", err)
	}

	// 验证深度注入是否成功 - 按照 TestForDeepInject 的验证方式
	fmt.Println("\n=== Testing Deep Injection ===")

	if app.UserService == nil {
		log.Fatal("app.UserService is nil")
	}

	if app.UserService.NotificationService == nil {
		log.Fatal("app.UserService.NotificationService is nil - deep injection failed!")
	}

	if app.UserService.NotificationService.EmailService == nil {
		log.Fatal("app.UserService.NotificationService.EmailService is nil - deep injection failed!")
	}

	// 验证引用关系 - 类似 TestForDeepInject 中的 c.B != b 检查
	if notificationService.EmailService != emailService {
		log.Fatal("notificationService.EmailService != emailService - reference mismatch!")
	}

	if app.UserService.NotificationService.EmailService != emailService {
		log.Fatal("Deep reference mismatch - deep injection failed!")
	}

	fmt.Println("✓ Deep injection successful!")
	fmt.Println("✓ All dependencies properly injected into manually created instances")
	fmt.Println("✓ Reference integrity maintained across deep injection")

	// 测试功能
	fmt.Println("\n=== Testing Functionality ===")
	app.UserService.CreateUser("Alice")

	fmt.Println("\n=== Deep Injection Complete ===")
	fmt.Println("This example demonstrates how go-inject can automatically inject")
	fmt.Println("dependencies into manually created object instances and their nested dependencies.")
	fmt.Println("Even when objects are created manually with empty dependencies,")
	fmt.Println("the injection system can populate them recursively.")
}
