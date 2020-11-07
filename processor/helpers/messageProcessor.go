package helpers

import "github.com/mebaranov/disguildie/message"

type MessageProcessor interface {
	ProcessMessage(m message.Message)
}
