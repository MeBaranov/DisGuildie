package user

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type ListProcessor struct {
	helpers.BaseMessageProcessor
}

func NewListProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &ListProcessor{}
	ap.Prov = prov
	return ap
}

func (ap *ListProcessor) ProcessMessage(m message.Message) (string, error) {
	ment := m.CurSegment()

	if (ment == "h" || ment == "help") && !m.MoreSegments() {
		return ap.help(m)
	}

	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	chars, err := ap.Prov.GetCharacters(m.GuildId(), u.Id)
	if err != nil {
		return "getting characters", err
	}

	rv := "List of characters:\n"
	for _, c := range chars {
		rv += "\t"
		if c.Main {
			rv += "[Main] "
		}
		rv += c.Name + "\n"
	}

	return rv, nil
}

func (ap *ListProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of characters listing commands you're allowed to use:\n"
	rv += "\t -- \"!g list\" (\"!g c\") - List your characters\n"
	rv += "\t -- \"!g list <mention user>\" (\"!g c <mention>\") - List users characters\n"

	return rv, nil
}
