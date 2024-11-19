package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    int    `json:"type"`
	From    string `json:"from"`
	Content string `json:"content"`
}

var (
	name = flag.String("name", "匿名用户", "设置你的聊天名称")
)

func main() {
	flag.Parse()

	url := fmt.Sprintf("ws://localhost:8000/chat/%s", *name)
	log.Printf("正在连接到 %s\n", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer conn.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if !strings.Contains(err.Error(), "closed network connection") {
					log.Println("读取消息错误:", err)
				}
				return
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err == nil {
				switch msg.Type {
				case 0: // 普通消息
					fmt.Printf("%s: %s\n", msg.From, msg.Content)
				case 1: // 系统消息
					fmt.Printf("\033[33m[系统] %s\033[0m\n", msg.Content)
				}
			} else {
				fmt.Printf("%s\n", message)
			}
		}
	}()

	fmt.Printf("以 %s 的身份连接成功！\n", *name)
	fmt.Println("开始聊天（输入 'quit' 退出）:")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			fmt.Println("\n正在关闭连接...")
			err := conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				log.Println("写入关闭消息错误:", err)
			}
			return
		default:
			if scanner.Scan() {
				text := scanner.Text()
				if strings.ToLower(text) == "quit" {
					return
				}

				if err := conn.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
					log.Println("发送消息错误:", err)
					return
				}
			}
		}
	}
}
