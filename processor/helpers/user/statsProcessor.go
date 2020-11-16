package user

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
	"github.com/mebaranov/disguildie/utility"
)

type StatsProcessor struct {
	helpers.BaseMessageProcessor
}

func NewStatsProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &StatsProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message) (string, error){
		"help": ap.help,
		"h":    ap.help,
		"list": ap.list,
		"l":    ap.list,
	}
	return ap
}

func (ap *StatsProcessor) ProcessMessage(m message.Message) (string, error) {
	v1, v2, v3, v4 := m.CurSegment(), m.CurSegment(), m.CurSegment(), m.CurSegment()
	if v1 != "" && v2 == "" && v3 == "" && v4 == "" {
		f, ok := ap.Funcs[v1]
		if ok {
			return f(m)
		}
	}

	if v3 == "" && v4 == "" {
		if v1 != "" && utility.IsUserMention(v1) {
			return ap.getStat(m, v1, v2)
		}
		if v2 == "" {
			return ap.getStat(m, "", v1)
		}

		return ap.setStat(m, "", "", v1, v2)
	}

	if v4 != "" {
		return ap.setStat(m, v1, v2, v3, v4)
	}

	if utility.IsUserMention(v1) {
		return ap.setStat(m, v1, "", v2, v3)
	}

	return ap.setStat(m, "", v1, v2, v3)
}

func (ap *StatsProcessor) getStat(m message.Message, ment string, char string) (string, error) {
	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	c, err := ap.Prov.GetCharacter(m.GuildId(), u.Id, char)
	if err != nil {
		return "getting character", err
	}

	gld, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	if c.StatVersion < gld.StatVersion {
		c, err = ap.Prov.SetCharacterStatVersion(c.GuildId, c.UserId, c.Name, gld.Stats, gld.StatVersion)
		if err != nil {
			return "setting stat version", err
		}
	}

	rvs := make([]string, 0, len(c.Body))
	for n, v := range c.Body {
		rvs = append(rvs, fmt.Sprintf("\t%v:%v\n", n, v))
	}

	sort.Strings(rvs)
	rv := fmt.Sprintf("Stats are:\n\tmain:%v\n\tname:%v\n", c.Main, c.Name)
	for _, s := range rvs {
		rv += s + "\n"
	}
	return rv, nil
}

func (ap *StatsProcessor) setStat(m message.Message, ment string, char string, stat string, value string) (string, error) {
	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to change this user")
	}

	c, err := ap.Prov.GetCharacter(m.GuildId(), u.Id, char)
	if err != nil {
		return "getting character", err
	}

	gld, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	s, ok := gld.Stats[stat]
	if !ok {
		return "", errors.New(fmt.Sprintf("Stat %v is not defined in your guild", stat))
	}

	var val interface{}
	switch s.Type {
	case database.Number:
		val, err = strconv.Atoi(value)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Expected numeric value. Got %v", value))
		}
	case database.Str:
		val = value
	}

	_, err = ap.Prov.SetCharacterStat(c.GuildId, c.UserId, c.Name, stat, val)
	if err != nil {
		return "setting character stat", err
	}

	if c.StatVersion < gld.StatVersion {
		c, err = ap.Prov.SetCharacterStatVersion(c.GuildId, c.UserId, c.Name, gld.Stats, gld.StatVersion)
		if err != nil {
			return "setting stat version", err
		}
	}

	return fmt.Sprintf("Stat %v set to %v for character %v", stat, value, c.Name), nil
}

func (ap *StatsProcessor) list(m message.Message) (string, error) {
	gld, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	rvs := make([]string, 0, len(gld.Stats))
	for _, v := range gld.Stats {
		t := database.TypeToString(v.Type)
		id := v.ID
		if v.ID == gld.DefaultStat {
			id = "(*) " + id
		}
		rvs = append(rvs, fmt.Sprintf("\t%v[%v]:%v\n", id, t, v.Description))
	}

	sort.Strings(rvs)
	rv := "Guild stats are:\n"
	for _, s := range rvs {
		rv += s + "\n"
	}
	rv += "* - default stat for sorting"
	return rv, nil
}

func (ap *StatsProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of charact stats commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	rv += "\t -- \"!g stat list\" (\"!g s l\") - List guild stats\n"

	rv += "\t -- \"!g stat\" (\"!g s\") - Get stats for your main character\n"
	rv += "\t -- \"!g stat <char name>\" (\"!g s <name>\") - Get stats for your character\n"
	rv += "\t -- \"!g stat <mention user>\" (\"!g s <mention>\") - Get stats for users main character\n"
	rv += "\t -- \"!g stat <mention user> <char name>\" (\"!g s <mention> <name>\") - Get stats for users character\n"

	rv += "\t -- \"!g stat <stat name> <stat value>\" (\"!g s <stat name> <stat value>\") - Set stat for your main character\n"
	rv += "\t -- \"!g stat <char name> <stat name> <stat value>\" (\"!g s <char name> <stat name> <stat value>\") - Set stat for your character\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g stat <mention user> <stat name> <stat value>\" (\"!g s <mention user> <stat name> <stat value>\") - Set stat for other users main character\n"
		rv += "\t -- \"!g stat <mention user> <char name> <stat name> <stat value>\" (\"!g s <mention user> <char name> <stat name> <stat value>\") - Set stat for other users character\n"
	}

	return rv, nil
}
