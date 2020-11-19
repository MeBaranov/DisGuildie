package helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/utility"
)

type MessageProcessor interface {
	ProcessMessage(m message.Message) (string, error)
}

type BaseMessageProcessor struct {
	Prov  database.DataProvider
	Funcs map[string]func(message.Message) (string, error)
}

func (ap *BaseMessageProcessor) ProcessMessage(m message.Message) (string, error) {
	cmd := m.PeekSegment()
	if utility.IsUserMention(cmd) {
		cmd = ""
	} else {
		cmd = m.CurSegment()
	}
	cmd = strings.ToLower(cmd)

	f, ok := ap.Funcs[cmd]
	if !ok {
		return "", errors.New(fmt.Sprintf("Unknown command \"%v\". Use \"!g help\" (\"!g h\") for help", m.FullMessage()))
	}

	return f(m)
}

func (ap *BaseMessageProcessor) UserOrAuthorByMention(ment string, m message.Message) (*database.User, error) {
	if ment != "" {
		uid, err := utility.ParseUserMention(ment)
		if err != nil {
			return nil, err
		}

		u, err := ap.Prov.GetUserD(uid)
		if err != nil {
			return nil, err
		}
		return u, nil
	} else {
		u, err := m.Author()

		if err != nil {
			return nil, err
		}
		return u, nil
	}

}
