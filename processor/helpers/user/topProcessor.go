package user

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type TopProcessor struct {
	helpers.BaseMessageProcessor
}

func NewTopProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &ListProcessor{}
	ap.Prov = prov
	return ap
}

func (ap *TopProcessor) ProcessMessage(m message.Message) (string, error) {
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

func (ap *TopProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of characters listing commands you're allowed to use:\n"
	rv += "\t -- \"!g top\" (\"!g t\") - Get guild top by default stat\n"
	rv += "\t -- \"!g top <stat>\" (\"!g t <stat>\") - Get guild top by stat name\n"

	return rv, nil
}
