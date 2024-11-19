

# 基于go语言的在线聊天室

### 项目结构

```shell
chat/
├── server/
│   ├── main.go #服务端
│   └── static/ #前端
│       ├── index.html
│       ├── css/
│       │   └── style.css
│       └── js/
│           └── chat.js
└── client/
    └── main.go #客户端
```

### 使用教程

```shell
git clone https://github.com/yty666zsy/chatroom_go.git
cd chatroom_go
```

先运行服务端代码

```shell
cd server
go run main.go
然后访问http://localhost:8000
```

如图：![image-20241119170248434](E:\go\project\test\image\image-20241119170248434.png)

然后运行客户端代码

```shell
cd ..
cd client
go run main.go -name "名字"
```

如图![image-20241119170546563](E:\go\project\test\image\image-20241119170546563.png)

一个简易的基于go语言的在线聊天室就搭建完成了