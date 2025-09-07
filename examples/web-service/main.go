package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/ComingCL/go-inject"
)

// 定义服务接口
type Logger interface {
	Log(message string)
}

type UserRepository interface {
	GetUser(id int) (*User, error)
	CreateUser(user *User) error
}

// 数据模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 实现具体的服务
type ConsoleLogger struct{}

func (l *ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", "INFO", message)
}

type InMemoryUserRepository struct {
	users  map[int]*User
	nextID int
	Logger Logger `inject:""`
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepository) GetUser(id int) (*User, error) {
	r.Logger.Log(fmt.Sprintf("Getting user with ID: %d", id))
	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}
	return user, nil
}

func (r *InMemoryUserRepository) CreateUser(user *User) error {
	r.Logger.Log(fmt.Sprintf("Creating user: %s", user.Name))
	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = user
	return nil
}

// 业务服务
type UserService struct {
	Repository UserRepository `inject:""`
	Logger     Logger         `inject:""`
}

func (s *UserService) GetUser(id int) (*User, error) {
	s.Logger.Log(fmt.Sprintf("UserService: Getting user %d", id))
	return s.Repository.GetUser(id)
}

func (s *UserService) CreateUser(name, email string) (*User, error) {
	s.Logger.Log(fmt.Sprintf("UserService: Creating user %s", name))
	user := &User{
		Name:  name,
		Email: email,
	}
	err := s.Repository.CreateUser(user)
	return user, err
}

// HTTP 处理器
type UserHandler struct {
	Service *UserService `inject:""`
	Logger  Logger       `inject:""`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	h.Logger.Log("Handling GET /users/{id}")

	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.Service.GetUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	h.Logger.Log("Handling POST /users")

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.Service.CreateUser(req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Web 服务器
type WebServer struct {
	Handler *UserHandler `inject:""`
	Logger  Logger       `inject:""`
}

func (s *WebServer) Start(port string) {
	s.Logger.Log(fmt.Sprintf("Starting web server on port %s", port))

	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.Handler.GetUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.Handler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.Logger.Log("Server started successfully")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	fmt.Println("Web 服务依赖注入示例")
	fmt.Println("启动 HTTP 服务器，演示在 Web 应用中使用依赖注入")
	fmt.Println()

	// 创建应用结构体
	var app struct {
		Logger     Logger         `inject:""`
		Repository UserRepository `inject:""`
		Service    *UserService   `inject:""`
		Handler    *UserHandler   `inject:""`
		Server     *WebServer     `inject:""`
	}

	// 进行依赖注入
	err := inject.Populate(
		&ConsoleLogger{},
		NewInMemoryUserRepository(),
		&UserService{},
		&UserHandler{},
		&WebServer{},
		&app,
	)
	if err != nil {
		log.Fatal("Failed to populate dependencies:", err)
	}

	// 创建一些测试数据
	app.Service.CreateUser("Alice", "alice@example.com")
	app.Service.CreateUser("Bob", "bob@example.com")

	fmt.Println("测试 API:")
	fmt.Println("GET  http://localhost:8080/users/1")
	fmt.Println("GET  http://localhost:8080/users/2")
	fmt.Println("POST http://localhost:8080/users")
	fmt.Println("     Body: {\"name\":\"Charlie\",\"email\":\"charlie@example.com\"}")
	fmt.Println()

	// 启动服务器
	app.Server.Start("8080")
}
