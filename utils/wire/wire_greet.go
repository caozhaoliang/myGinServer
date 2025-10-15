package wire

import "fmt"

type Message string

func NewMessage() Message {
	return Message("测试message")
}

type Greeter struct {
	message Message
}

func NewGreeter(message Message) Greeter {
	return Greeter{
		message: message,
	}
}
func (g *Greeter) Greet() string {
	return string(g.message)
}

// Event 事件，依赖 Greeter
type Event struct {
	Greeter Greeter
}

// NewEvent 创建 Event 的 Provider 函数
func NewEvent(g Greeter) Event {
	return Event{Greeter: g}
}

// Start 事件开始
func (e Event) Start() {
	msg := e.Greeter.Greet()
	fmt.Println(msg)
}
