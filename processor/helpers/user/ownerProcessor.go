package user

import (
	"errors"
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type OwnerProcessor struct {
	helpers.BaseMessageProcessor
}

func NewOwnerProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &OwnerProcessor{}
	ap.Prov = prov
	return ap
}

func (ap *OwnerProcessor) ProcessMessage(m message.Message) (string, error) {
	c := m.CurSegment()
	if c == "h" || c == "help" {
		return ap.help(m)
	}

	if c == "" {
		return "", errors.New("Invalid command format. Try \"!g o h\"")
	}

	chars, err := ap.Prov.GetCharactersByName(m.GuildId(), c)
	if err != nil {
		return "getting characters by name", err
	}

	if len(chars) == 0 {
		return "", errors.New("Characters with name " + c + " are not present in the guild")
	}

	rv := "Members of the guild who have characters named " + c + ":"
	for _, char := range chars {
		rv += fmt.Sprintf(" <@!%v>,", char.UserId)
	}

	return rv, nil
}

func (ap *OwnerProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of owners commands you're allowed to use:\n"
	rv += "\t -- \"!g owner <char name>\" (\"!g o <name>\") - Get possible owners of a character with specified name\n"
	rv += ";)"

	return rv, nil
}
