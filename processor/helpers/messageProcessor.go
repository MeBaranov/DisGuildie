package helpers

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
)

type MessageProcessor interface {
	ProcessMessage(m message.Message)
}

type BaseMessageProcessor struct {
	Prov  database.DataProvider
	Funcs map[string]func(message.Message)
}

func (ap *BaseMessageProcessor) ProcessMessage(m message.Message) {
	cmd := m.CurSegment()

	f, ok := ap.Funcs[cmd]
	if !ok {
		m.SendMessage("Unknown command \"%v\". Use \"!g help\" or \"!g h\" for help", m.FullMessage())
		return
	}

	f(m)
}
