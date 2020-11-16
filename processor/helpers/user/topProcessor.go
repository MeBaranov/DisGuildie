package user

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type TopProcessor struct {
	helpers.BaseMessageProcessor
}

func NewTopProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &TopProcessor{}
	ap.Prov = prov
	return ap
}

func (ap *TopProcessor) ProcessMessage(m message.Message) (string, error) {
	t, s, l := m.CurSegment(), m.CurSegment(), m.CurSegment()
	asc := false
	if t == "asc" || t == "a" {
		asc = true
	} else if l != "" {
		return "", errors.New("Invalid command format")
	} else {
		l = s
		s = t
	}

	if l == "" && (s == "h" || s == "help") && !m.MoreSegments() {
		return ap.help(m)
	}

	var err error
	limit := -1
	if l == "" {
		limit, err = strconv.Atoi(s)
		if err != nil {
			limit = -1
		} else {
			s = ""
		}
	} else {
		limit, err = strconv.Atoi(l)
		if err != nil {
			return "", errors.New("Invalid command format")
		}
	}

	gld, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}
	if s == "" {
		s = gld.DefaultStat
	}
	if s == "" {
		return "", errors.New("This guild doesn't have any stats yet")
	}

	stat, ok := gld.Stats[s]
	if !ok {
		return "", errors.New("Stat with name " + s + " is not defined in guild")
	}

	chars, err := ap.Prov.GetCharactersOutdated(m.GuildId(), gld.StatVersion)
	if err != nil {
		return "getting outdated characters", nil
	}

	for _, c := range chars {
		_, err = ap.Prov.SetCharacterStatVersion(c.GuildId, c.UserId, c.Name, gld.Stats, gld.StatVersion)
		if err != nil {
			return "setting character stat version", err
		}
	}

	chars, err = ap.Prov.GetCharactersSorted(m.GuildId(), stat.ID, stat.Type, asc, limit)
	if err != nil {
		return "getting sorted characters", err
	}

	if limit <= 0 {
		limit = len(chars)
	}
	rv := fmt.Sprintf("Top %v characters by %v.", limit, stat.ID)
	if asc {
		rv += " (Lowest first)\n"
	} else {
		rv += " (Highest first)\n"
	}

	for i, c := range chars {
		rv += fmt.Sprintf("\t%v : %v : %v", i, c.Name, c.Body[stat.ID])
	}

	return rv, nil
}

func (ap *TopProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of guild tops commands you're allowed to use:\n"
	rv += "\t -- \"!g top\" (\"!g t\") - Get guild top characters by default stat (descending)\n"
	rv += "\t -- \"!g top <count>\" (\"!g t <count>\") - Get guild top <count> characters by default stat (descending)\n"
	rv += "\t -- \"!g top <stat>\" (\"!g t <stat>\") - Get guild top characters by stat name (descending)\n"
	rv += "\t -- \"!g top <stat> <count>\" (\"!g t <stat> <count>\") - Get guild top <count> characters by stat name (descending)\n"
	rv += "\nTo get top in ascending order - pass the same commands with \"!g top asc\" (\"!g t a\") prefix. For example:\n"
	rv += "\t -- \"!g top asc <stat> <count>\" (\"!g t a <stat> <count>\") - Get guild top <count> characters by stat name in ascendong order\n"

	return rv, nil
}
