package admin

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type AdminGuildProcessor struct {
	helpers.BaseMessageProcessor
}

func NewAdminGuildProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &AdminGuildProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message){
		"h":      ap.help,
		"help":   ap.help,
		"a":      ap.add,
		"add":    ap.add,
		"r":      ap.rename,
		"rename": ap.rename,
		"m":      ap.move,
		"move":   ap.move,
		"remove": ap.remove,
	}
	return ap
}

func (ap *AdminGuildProcessor) add(m message.Message) {
	gldName := m.CurSegment()
	parent := m.CurSegment()

	if parent == "" || gldName == "" {
		m.SendMessage("Malformed command. Expected sub-guild name and parent name. Received: '%v'", m.FullMessage())
		return
	}

	var pguild *database.Guild
	var err error
	if parent == "main" {
		pguild, err = ap.Prov.GetGuildD(m.GuildId())
	} else {
		pguild, err = ap.Prov.GetGuildN(m.GuildId(), parent)
	}
	if err != nil {
		m.SendMessage("Error getting parrent guild: %v", err.Error())
		return
	}

	g := &database.Guild{
		Name:     gldName,
		ParentId: pguild.GuildId,
	}
	if _, err = ap.Prov.AddGuild(g); err != nil {
		m.SendMessage("Error creating subguild: %v", err.Error())
		return
	}

	m.SendMessage("Subguild %v registered under %v.", gldName, parent)
}

func (ap *AdminGuildProcessor) rename(m message.Message) {
	oldName := m.CurSegment()
	newName := m.CurSegment()
	if oldName == "" || newName == "" {
		m.SendMessage("Malformed command. Expected sub-guild old name and new name. Received: '%v'", m.FullMessage())
		return
	}

	g, err := ap.Prov.GetGuildN(m.GuildId(), oldName)
	if err != nil {
		m.SendMessage("Error: %v", err.Error())
		return
	}

	if _, err = ap.Prov.RenameGuild(g.GuildId, newName); err != nil {
		m.SendMessage("Error: %v", err.Error())
		return
	}

	m.SendMessage("Guild renamed to %v", newName)
}

func (ap *AdminGuildProcessor) move(m message.Message) {
	name := m.CurSegment()
	parent := m.CurSegment()
	if name == "" || parent == "" {
		m.SendMessage("Malformed command. Expected sub-guild name and parent name. Received: '%v'", m.FullMessage())
		return
	}

	g, err := ap.Prov.GetGuildN(m.GuildId(), name)
	if err != nil {
		m.SendMessage("Error getting guild: %v", err.Error())
		return
	}

	var pguild *database.Guild
	if parent == "main" {
		pguild, err = ap.Prov.GetGuildD(m.GuildId())
	} else {
		pguild, err = ap.Prov.GetGuildN(m.GuildId(), parent)
	}
	if err != nil {
		m.SendMessage("Error getting parrent guild: %v", err.Error())
		return
	}

	if _, err = ap.Prov.MoveGuild(g.GuildId, pguild.GuildId); err != nil {
		m.SendMessage("Error: %v", err.Error())
		return
	}

	m.SendMessage("Guild %v moved under parent %v", name, parent)
}

func (ap *AdminGuildProcessor) remove(m message.Message) {
	name := m.CurSegment()
	if name == "" {
		m.SendMessage("Malformed command. Expected sub-guild name. Received: '%v'", m.FullMessage())
		return
	}

	g, err := ap.Prov.GetGuildN(m.GuildId(), name)
	if err != nil {
		m.SendMessage("Error getting guild: %v", err.Error())
		return
	}

	subs, err := ap.Prov.GetSubGuilds(g.GuildId)
	if err != nil {
		m.SendMessage("Error getting subguilds: %v", err.Error())
		return
	}

	users, err := ap.Prov.GetUsersInGuild(m.GuildId())
	if err != nil {
		m.SendMessage("Error getting users: %v", err.Error())
		return
	}

	for _, s := range users {
		if perm, ok := s.Guilds[m.GuildId()]; ok {
			if _, ok = subs[perm.GuildId]; ok {
				_, err = ap.Prov.SetUserSubGuild(s.DiscordId, &database.GuildPermission{TopGuild: m.GuildId(), GuildId: g.ParentId})
				if err != nil {
					m.SendMessage("Error moving users from subguilds: %v", err.Error())
					return
				}
			}
		}
	}

	if _, err = ap.Prov.RemoveGuild(g.GuildId); err != nil {
		m.SendMessage("Error removing guild: %v", err.Error())
		return
	}

	m.SendMessage("Guild %v removed", name)
}

func (ap *AdminGuildProcessor) help(m message.Message) {
	rv := "Here's a list of guild management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		m.SendMessage(rv)
		return
	}

	if perm&database.StructurePermissions == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		m.SendMessage(rv)
		return
	}

	rv += "\t -- \"!g admin guild add <child guild name> main\" (\"!g a g a <name>\") - Add sub-guild to the main level\n"
	rv += "\t -- \"!g admin guild add <child guild name> <parent guild name>\" (\"!g a g a <child> <parent>\") - Add sub-guild to a parent sub-guild\n"
	rv += "\t -- \"!g admin guild rename <old sub-guild name> <new name>\" (\"!g a g r <old> <new>\") - Rename sub-guild\n"
	rv += "\t -- \"!g admin guild move <child guild name> main\" (\"!g a g m <name> main\") - Move subguild to a the main level\n"
	rv += "\t -- \"!g admin guild move <child guild name> <new parent guild>\" (\"!g a g m <name> <new parent>\") - Move subguild to a new parent\n"
	rv += "\t -- \"!g admin guild remove <child guild name>\" (\"!g a g remove <name>\") - Remove sub-guild\n"
	rv += "Be aware that your ability to modify structure depends on the guild you're assigned to.\n"

	m.SendMessage(rv)
}
