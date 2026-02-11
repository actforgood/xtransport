package broker

type Message struct {
	Body  []byte
	Props Props
}

type MessageChan chan Message
