package chat_msg

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
)

type Message struct {
	Id       int64  `db:"id"`
	FormId   string `db:"form_id"`
	TargetId string `db:"target_id"`
	Type     string `db:"type"`  // 群聊 私聊 广播
	Media    string `db:"media"` // 文字 图片 音频
	Content  string `db:"content"`
	Pic      string `db:"pic"`
	Url      string `db:"url"`
	Desc     string `db:"desc"`
	SendTime int64  `db:"send_time"`
	IsRead   bool   `db:"is_read"`
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

var (
	clientMap   = make(map[string]*Node)
	rwLocker    sync.RWMutex
	udpSendChan = make(chan []byte, 1024)
)

func Chat(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userId := query.Get("userId")
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//context := query.Get("context")
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	go sendProc(node)
	go recvProc(node)
	SendMsg("admin", userId, []byte("欢迎进入"))
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("[ws]<<<<<", string(data))
		broadMsg(data)
	}
}

func broadMsg(data []byte) {
	udpSendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
}

// UDP 广播消息发送
func udpSendProc() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 255, 255, 255),
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	for {
		select {
		case data := <-udpSendChan:
			_, err1 := conn.Write(data)
			if err1 != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// UDP 广播消息接收
func udpRecvProc() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 3000})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	for {
		var buf [1024]byte
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatchMsg(buf[0:n])
	}
}
func dispatchMsg(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case "":
		// todo 发送消息
		SendMsg(msg.FormId, msg.TargetId, data)
	}
}

func SendMsg(userId, targetId string, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[targetId]
	rwLocker.RUnlock()
	if ok {
		node.DataQueue <- msg
	}
}
