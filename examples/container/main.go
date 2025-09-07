package main

import (
	"fmt"
	"log"

	"github.com/ComingCL/go-inject"
)

// 定义服务接口
type UserService interface {
	GetUser(id int) *User
	CreateUser(name, email string) *User
}

type EmailService interface {
	SendEmail(to, subject, body string) error
}

// 定义数据模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 实现用户服务
type userServiceImpl struct {
	EmailService EmailService `inject:""`
}

func (s *userServiceImpl) GetUser(id int) *User {
	return &User{
		ID:    id,
		Name:  "John Doe",
		Email: "john@example.com",
	}
}

func (s *userServiceImpl) CreateUser(name, email string) *User {
	user := &User{
		ID:    1,
		Name:  name,
		Email: email,
	}

	// 发送欢迎邮件
	if s.EmailService != nil {
		s.EmailService.SendEmail(email, "Welcome!", "Welcome to our service!")
	}

	return user
}

// 实现邮件服务
type emailServiceImpl struct{}

func (s *emailServiceImpl) SendEmail(to, subject, body string) error {
	fmt.Printf("📧 Sending email to %s\n", to)
	fmt.Printf("   Subject: %s\n", subject)
	fmt.Printf("   Body: %s\n", body)
	return nil
}

// 应用程序结构
type App struct {
	UserService  UserService  `inject:""`
	EmailService EmailService `inject:""`
}

func (app *App) Run() {
	fmt.Println("🚀 Starting Container Example Application...")

	// 获取用户
	user := app.UserService.GetUser(1)
	fmt.Printf("📋 Retrieved user: %+v\n", user)

	// 创建新用户
	newUser := app.UserService.CreateUser("Alice Smith", "alice@example.com")
	fmt.Printf("✨ Created new user: %+v\n", newUser)
}

func main() {
	fmt.Println("=== Go-Inject Container Example ===")

	// 创建 IoC 容器
	container := inject.NewContainer()

	// 创建服务实例
	userService := &userServiceImpl{}
	emailService := &emailServiceImpl{}
	app := &App{}

	// 向容器中提供服务
	if err := container.Provides(userService, emailService, app); err != nil {
		log.Fatalf("❌ Failed to provide services: %v", err)
	}

	// 也可以使用命名方式提供服务
	// container.ProvideWithName("userService", userService)
	// container.ProvideWithName("emailService", emailService)

	// 填充依赖关系
	if err := container.Populate(); err != nil {
		log.Fatalf("❌ Failed to populate dependencies: %v", err)
	}

	fmt.Println("✅ Container initialized successfully!")

	// 运行应用程序
	app.Run()

	fmt.Println("\n=== Container Example Completed ===")
}
