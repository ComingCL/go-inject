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

type Cache interface {
	Get(key string) string
	Set(key string, value string)
}

// 实现具体的服务
type ConsoleLogger struct {
	Prefix string
}

func (l *ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", l.Prefix, message)
}

type MySQLDatabase struct {
	ConnectionString string
}

func (db *MySQLDatabase) Query(sql string) string {
	return fmt.Sprintf("MySQL(%s): %s", db.ConnectionString, sql)
}

type PostgreSQLDatabase struct {
	ConnectionString string
}

func (db *PostgreSQLDatabase) Query(sql string) string {
	return fmt.Sprintf("PostgreSQL(%s): %s", db.ConnectionString, sql)
}

type RedisCache struct {
	Host string
}

func (c *RedisCache) Get(key string) string {
	return fmt.Sprintf("Redis(%s): cached_%s", c.Host, key)
}

func (c *RedisCache) Set(key string, value string) {
	fmt.Printf("Redis(%s): Set %s = %s\n", c.Host, key, value)
}

type MemoryCache struct{}

func (c *MemoryCache) Get(key string) string {
	return fmt.Sprintf("Memory: cached_%s", key)
}

func (c *MemoryCache) Set(key string, value string) {
	fmt.Printf("Memory: Set %s = %s\n", key, value)
}

// 使用命名注入的服务
type OrderService struct {
	// 使用命名注入来区分不同的数据库
	ReadDB  Database `inject:"read-db"`
	WriteDB Database `inject:"write-db"`

	// 使用命名注入来区分不同的缓存
	L1Cache Cache `inject:"l1-cache"`
	L2Cache Cache `inject:"l2-cache"`

	Logger Logger `inject:"order-logger"`
}

func (s *OrderService) CreateOrder(orderID string) {
	s.Logger.Log(fmt.Sprintf("Creating order: %s", orderID))

	// 写入主数据库
	result := s.WriteDB.Query(fmt.Sprintf("INSERT INTO orders (id) VALUES ('%s')", orderID))
	s.Logger.Log(result)

	// 缓存到 L1 和 L2
	s.L1Cache.Set("order_"+orderID, "created")
	s.L2Cache.Set("order_"+orderID, "created")
}

func (s *OrderService) GetOrder(orderID string) {
	s.Logger.Log(fmt.Sprintf("Getting order: %s", orderID))

	// 先尝试 L1 缓存
	cached := s.L1Cache.Get("order_" + orderID)
	if cached != "" {
		s.Logger.Log("Found in L1 cache: " + cached)
		return
	}

	// 再尝试 L2 缓存
	cached = s.L2Cache.Get("order_" + orderID)
	if cached != "" {
		s.Logger.Log("Found in L2 cache: " + cached)
		// 回填到 L1 缓存
		s.L1Cache.Set("order_"+orderID, cached)
		return
	}

	// 最后从读数据库查询
	result := s.ReadDB.Query(fmt.Sprintf("SELECT * FROM orders WHERE id = '%s'", orderID))
	s.Logger.Log("From read DB: " + result)
}

// 用户服务使用不同的依赖配置
type UserService struct {
	Database Database `inject:"write-db"` // 用户服务只需要写数据库
	Cache    Cache    `inject:"l1-cache"` // 只使用 L1 缓存
	Logger   Logger   `inject:"user-logger"`
}

func (s *UserService) CreateUser(userID string) {
	s.Logger.Log(fmt.Sprintf("Creating user: %s", userID))
	result := s.Database.Query(fmt.Sprintf("INSERT INTO users (id) VALUES ('%s')", userID))
	s.Logger.Log(result)
	s.Cache.Set("user_"+userID, "created")
}

func main() {
	fmt.Println("高级依赖注入示例 - 命名注入和多实例管理")
	fmt.Println()

	// 创建依赖注入图
	var g inject.Graph

	// 注册命名的依赖
	err := g.Provide(
		// 数据库实例
		&inject.Object{
			Value: &MySQLDatabase{ConnectionString: "mysql://localhost:3306/orders"},
			Name:  "write-db",
		},
		&inject.Object{
			Value: &PostgreSQLDatabase{ConnectionString: "postgres://localhost:5432/orders_read"},
			Name:  "read-db",
		},

		// 缓存实例
		&inject.Object{
			Value: &MemoryCache{},
			Name:  "l1-cache",
		},
		&inject.Object{
			Value: &RedisCache{Host: "localhost:6379"},
			Name:  "l2-cache",
		},

		// 日志实例
		&inject.Object{
			Value: &ConsoleLogger{Prefix: "ORDER"},
			Name:  "order-logger",
		},
		&inject.Object{
			Value: &ConsoleLogger{Prefix: "USER"},
			Name:  "user-logger",
		},

		// 业务服务
		&inject.Object{Value: &OrderService{}},
		&inject.Object{Value: &UserService{}},
	)
	if err != nil {
		log.Fatal("Failed to provide dependencies:", err)
	}

	// 执行依赖注入
	if err := g.Populate(); err != nil {
		log.Fatal("Failed to populate dependencies:", err)
	}

	// 获取注入后的服务
	var orderService *OrderService
	var userService *UserService

	for _, obj := range g.Objects() {
		switch v := obj.Value.(type) {
		case *OrderService:
			orderService = v
		case *UserService:
			userService = v
		}
	}

	// 演示服务使用
	fmt.Println("=== 订单服务演示 ===")
	orderService.CreateOrder("ORDER-001")
	orderService.GetOrder("ORDER-001")
	orderService.GetOrder("ORDER-002") // 不存在的订单

	fmt.Println("\n=== 用户服务演示 ===")
	userService.CreateUser("USER-001")

	fmt.Println("\n=== 依赖注入图信息 ===")
	fmt.Printf("总共注册了 %d 个对象\n", len(g.Objects()))
	for _, obj := range g.Objects() {
		if obj.Name != "" {
			fmt.Printf("- %s (名称: %s)\n", obj.String(), obj.Name)
		} else {
			fmt.Printf("- %s\n", obj.String())
		}
	}
}
