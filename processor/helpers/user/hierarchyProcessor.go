package user

import (
	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type HierarchyProcessor struct {
	helpers.BaseMessageProcessor
}

func NewHierarchyProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &HierarchyProcessor{}
	ap.Prov = prov
	return ap
}

func (ap *HierarchyProcessor) ProcessMessage(m message.Message) (string, error) {
	s := m.CurSegment()
	if s == "h" || s == "help" {
		return ap.help(m)
	}

	var gld *database.Guild
	var err error
	if s != "" {
		gld, err = ap.Prov.GetGuildN(m.GuildId(), s)
		if err != nil {
			return "getting subguild", err
		}
	} else {
		gld, err = ap.Prov.GetGuildD(m.GuildId())
		if err != nil {
			return "getting guild", err
		}
	}

	a := uuid.Nil
	auth, err := m.Author()
	if err == nil {
		if g, ok := auth.Guilds[m.GuildId()]; ok {
			a = g.GuildId
		}
	}

	subs, err := ap.Prov.GetSubGuilds(gld.GuildId)
	if err != nil {
		return "getting subguilds", err
	}

	rv := "Sub-guilds hierarchy:\n"
	rv += gld.Name
	if gld.GuildId == a {
		rv += "<-- You are here"
	}
	rv += "\n"
	tmp, err := ap.printSubGuild(subs, gld.GuildId, "-", a)
	if err != nil {
		return tmp, err
	}
	rv += tmp
	return rv, nil
}

func (ap *HierarchyProcessor) printSubGuild(subs map[uuid.UUID]*database.Guild, gld uuid.UUID, t string, auth uuid.UUID) (string, error) {
	if len(t) > 15 {
		return "", nil
	}

	rv := ""
	for _, s := range subs {
		if s.ParentId == gld {
			rv += t + s.Name
			if s.GuildId == auth {
				rv += " <-- You are here"
			}
			rv += "\n"

			tmp, err := ap.printSubGuild(subs, s.GuildId, t+"-", auth)
			if err != nil {
				return tmp, err
			}
			rv += tmp
		}
	}

	return rv, nil
}

func (ap *HierarchyProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of hierarchy commands you're allowed to use:\n"
	rv += "\t -- \"!g hierarchy\" (\"!g hi\") - Get sub-guilds hierarchy\n"
	rv += "\t -- \"!g hierarchy <sub-guild name>\" (\"!g hi <name>\") - Get sub-guilds hierarchy for a sub-guild\n"

	return rv, nil
}
