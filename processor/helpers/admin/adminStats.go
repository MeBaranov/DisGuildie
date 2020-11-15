package admin

import (
	"errors"
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type AdminStatsProcessor struct {
	helpers.BaseMessageProcessor
}

func NewAdminStatsProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &AdminStatsProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message) (string, error){
		"h":      ap.help,
		"help":   ap.help,
		"a":      ap.add,
		"add":    ap.add,
		"r":      ap.remove,
		"remove": ap.remove,
		"reset":  ap.reset,
	}
	return ap
}

func (ap *AdminStatsProcessor) add(m message.Message) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author premissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		return "", errors.New("You don't have permissions to do guild-wide structure modifications.")
	}

	n, t, d := m.CurSegment(), m.CurSegment(), m.CurSegment()
	if n == "" || t == "" {
		return "", errors.New("Invalid command format")
	}

	tval, err := database.StringToType(t)
	if err != nil {
		return "parsing type", err
	}

	g, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	if _, ok := g.Stats[n]; ok {
		return "", errors.New(fmt.Sprintf("Stat with name %v already exists in the system", n))
	}

	stat := database.Stat{
		ID:          n,
		Type:        tval,
		Description: d,
	}
	if _, err := ap.Prov.AddGuildStat(g.GuildId, &stat); err != nil {
		return "adding stat", err
	}

	return fmt.Sprintf("Stat %v with type %v was added.", n, t), nil
}

func (ap *AdminStatsProcessor) remove(m message.Message) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		return "", errors.New("You don't have permissions to do guild-wide structure modifications.")
	}

	n := m.CurSegment()
	if n == "" {
		return "", errors.New("Invalid command format")
	}

	g, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	if _, ok := g.Stats[n]; !ok {
		return "", errors.New(fmt.Sprintf("Stat %v does not exist in the guild", n))
	}

	if _, err := ap.Prov.RemoveGuildStat(g.GuildId, n); err != nil {
		return "removing stat", err
	}

	return fmt.Sprintf("Stat %v was removed.", n), nil
}

func (ap *AdminStatsProcessor) reset(m message.Message) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		return "", errors.New("You don't have permissions to do guild-wide structure modifications.")
	}

	g, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	if _, err := ap.Prov.RemoveAllGuildStats(g.GuildId); err != nil {
		return "resetting stats", err
	}

	return "All stats were reset in the guild.", nil
}

func (ap *AdminStatsProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of stats management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting permissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		return rv, nil
	}

	rv += "Stats are identified by name. Stat type can be either \"int\" for numbers or \"str\" for everything else"
	rv += "\t -- \"!g admin stats add <statName> <statType> <description>\" (\"!g a s a <statName> <statType> <description>\") - Add a stat with description\n"
	rv += "\t -- \"!g admin stats add <statName> <statType>\" (\"!g a s a <statName> <statType>\") - Add a stat without description\n"
	rv += "\t -- \"!g admin stats remove <statName>\" (\"!g a s r <statName>\") - Remove a stat (notice that it will not be removed from existing characters data)\n"
	rv += "\t -- \"!g admin stats reset\" (\"!g a s reset\") - Remove all stats that were set\n"

	return rv, nil
}
