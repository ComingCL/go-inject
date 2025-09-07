package main

import (
	"fmt"
	"log"

	"github.com/ComingCL/go-inject"
)

// 定义服务接口
type Logger interface {
	Log(message string)
}

type Database interface {
	Query(sql string) string
}

// 实现具体的服务
type ConsoleLogger struct{}

func (l *ConsoleLogger) Log(message string) {
	fmt.Println("[LOG]", message)
}

type MySQLDatabase struct{}

func (db *MySQLDatabase) Query(sql string) string {
	return fmt.Sprintf("Executing: %s", sql)
}

// 业务服务，依赖于 Logger 和 Database
type UserService struct {
	Logger   Logger   `inject:""`
	Database Database `inject:""`
}

func (s *UserService) CreateUser(name string) {
	s.Logger.Log(fmt.Sprintf("Creating user: %s", name))
	result := s.Database.Query("INSERT INTO users (name) VALUES (?)")
	s.Logger.Log(result)
}

func main() {
	// 创建包含服务的结构体
	var app struct {
		Logger      Logger       `inject:""`
		Database    Database     `inject:""`
		UserService *UserService `inject:""`
	}

	// 使用便捷函数进行依赖注入
	err := inject.Populate(
		&ConsoleLogger{},
		&MySQLDatabase{},
		&UserService{},
		&app,
	)
	if err != nil {
		log.Fatal("Failed to populate dependencies:", err)
	}

	// 使用注入后的服务
	app.UserService.CreateUser("Alice")
	app.UserService.CreateUser("Bob")
}
