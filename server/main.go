package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type User struct {
	Name string
	Conn *websocket.Conn
	Send chan []byte
}

type Message struct {
	Type    int    `json:"type"` // 0: 普通消息, 1: 系统消息
	From    string `json:"from"`
	Content string `json:"content"`
}

type ChatRoom struct {
	users     map[string]*User
	broadcast chan *Message
	mutex     sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func newChatRoom() *ChatRoom {
	return &ChatRoom{
		users:     make(map[string]*User),
		broadcast: make(chan *Message),
	}
}

func (cr *ChatRoom) run() {
	for msg := range cr.broadcast {
		data, _ := json.Marshal(msg)
		cr.mutex.RLock()
		for _, user := range cr.users {
			select {
			case user.Send <- data:
			default:
				close(user.Send)
				delete(cr.users, user.Name)
			}
		}
		cr.mutex.RUnlock()
	}
}

func (cr *ChatRoom) broadcastUserList() {
	userList := make([]string, 0)
	cr.mutex.RLock()
	for name := range cr.users {
		userList = append(userList, name)
	}
	cr.mutex.RUnlock()

	cr.broadcast <- &Message{
		Type:    1,
		From:    "System",
		Content: "当前在线用户: " + strings.Join(userList, ", "),
	}
}

func handleWebSocket(cr *ChatRoom, w http.ResponseWriter, r *http.Request, name string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket升级失败:", err)
		return
	}

	user := &User{
		Name: name,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	cr.mutex.Lock()
	cr.users[name] = user
	cr.mutex.Unlock()

	// 广播新用户加入
	cr.broadcast <- &Message{
		Type:    1,
		From:    "System",
		Content: name + " 加入了聊天室",
	}
	cr.broadcastUserList()

	// 启动一个goroutine来处理发送消息
	go func() {
		defer func() {
			cr.mutex.Lock()
			if _, ok := cr.users[name]; ok {
				delete(cr.users, name)
				close(user.Send)
			}
			cr.mutex.Unlock()
			conn.Close()

			cr.broadcast <- &Message{
				Type:    1,
				From:    "System",
				Content: name + " 离开了聊天室",
			}
			cr.broadcastUserList()
		}()

		for msg := range user.Send {
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				return
			}
		}
	}()

	// 处理接收的消息
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			return
		}

		cr.broadcast <- &Message{
			Type:    0,
			From:    name,
			Content: string(p),
		}
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	chatRoom := newChatRoom()
	go chatRoom.run()

	r.Static("/static", "./static")
	r.LoadHTMLGlob("static/*.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// WebSocket路由
	r.GET("/chat/:name", func(c *gin.Context) {
		name := c.Param("name")
		handleWebSocket(chatRoom, c.Writer, c.Request, name)
	})

	port := ":8000"
	log.Printf("聊天服务器启动在 http://localhost%s\n", port)
	r.Run(port)
}
