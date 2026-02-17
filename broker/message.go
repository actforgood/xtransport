package broker

// Message represents a message to be sent or received from the broker.
type Message struct {
	// Body contains the actual content of the message.
	Body []byte
	// Props contains the properties of the message, such as headers, content type, etc.
	Props Props
}

// MessageChan represents a channel for receiving messages.
type MessageChan chan Message
