package admin

import (
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
	ap.Funcs = map[string]func(message.Message){
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

func (ap *AdminStatsProcessor) add(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		m.SendMessage("Sorry, you don't have permissions to do guild-wide structure modifications.")
		return
	}

	n, t, d := m.CurSegment(), m.CurSegment(), m.CurSegment()
	if n == "" || t == "" {
		m.SendMessage("Malformed command. Expected stat name and type")
		return
	}

	tval, err := database.StringToType(t)
	if err != nil {
		m.SendMessage(err.Error())
	}

	g, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		m.SendMessage("Error getting guild. Something must be terribly wrong. Try again")
		return
	}

	if _, ok := g.Stats[n]; ok {
		m.SendMessage("Stat with name %v already exists in the system", n)
		return
	}

	stat := database.Stat{
		ID:          n,
		Type:        tval,
		Description: d,
	}
	if _, err := ap.Prov.AddGuildStat(g.GuildId, &stat); err != nil {
		m.SendMessage("Could not add stat: %v", err.Error())
		return
	}

	m.SendMessage("Stat %v with type %v was added.", n, t)
}

func (ap *AdminStatsProcessor) remove(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		m.SendMessage("Sorry, you don't have permissions to do guild-wide structure modifications.")
		return
	}

	n := m.CurSegment()
	if n == "" {
		m.SendMessage("Malformed command. Expected stat name")
		return
	}

	g, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		m.SendMessage("Error getting guild. Something must be terribly wrong. Try again")
		return
	}

	if _, ok := g.Stats[n]; !ok {
		m.SendMessage("Stat does %v does not exist in the guild", n)
		return
	}

	if _, err := ap.Prov.RemoveGuildStat(g.GuildId, n); err != nil {
		m.SendMessage("Could not remove stat: %v", err.Error())
		return
	}

	m.SendMessage("Stat %v was removed.", n)
}

func (ap *AdminStatsProcessor) reset(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		m.SendMessage("Sorry, you don't have permissions to do guild-wide structure modifications.")
		return
	}

	g, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		m.SendMessage("Error getting guild. Something must be terribly wrong. Try again")
		return
	}

	if _, err := ap.Prov.RemoveAllGuildStats(g.GuildId); err != nil {
		m.SendMessage("Could not reset guild stats: %v", err.Error())
		return
	}

	m.SendMessage("All stats were reset in the guild.")
}

func (ap *AdminStatsProcessor) help(m message.Message) {
	rv := "Here's a list of stats management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		m.SendMessage(rv)
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		m.SendMessage(rv)
		return
	}

	rv += "Stats are identified by name. Stat type can be either \"int\" for numbers or \"str\" for everything else"
	rv += "\t -- \"!g admin stats add <statName> <statType> <description>\" (\"!g a s a <statName> <statType> <description>\") - Add a stat with description\n"
	rv += "\t -- \"!g admin stats add <statName> <statType>\" (\"!g a s a <statName> <statType>\") - Add a stat without description\n"
	rv += "\t -- \"!g admin stats remove <statName>\" (\"!g a s r <statName>\") - Remove a stat (notice that it will not be removed from existing characters data)\n"
	rv += "\t -- \"!g admin stats reset\" (\"!g a s reset\") - Remove all stats that were set\n"

	m.SendMessage(rv)
}
