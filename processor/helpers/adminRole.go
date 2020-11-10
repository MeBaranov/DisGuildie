package helpers

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
)

type AdminRoleProcessor struct {
	BaseMessageProcessor
}

func NewAdminRoleProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminRoleProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message){
		"h":    ap.help,
		"help": ap.help,
	}
	return ap
}

func (ap *AdminRoleProcessor) add(m message.Message) {

	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		m.SendMessage("I'm sorry, but you don't have permissions to run guild-wide structure management operations")
		return
	}
}

func (ap *AdminRoleProcessor) help(m message.Message) {
	rv := "Here's a list of guild management commands you're allowed to use:\n"

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

	rv += "\t -- \"!g admin role add <role> <permission>\" (\"!g a r a <role> <permission>\") - Add permission to a role\n"
	rv += "\t -- \"!g admin role remove <role> <permission>\" (\"!g a r remove <role> <permission>\") - Remove permission from a role\n"
	rv += "\t -- \"!g admin role reset <role>\" (\"!g a r reset <role>\") - Remove permission from a role\n"
	rv += "\n"
	rv += "In the explanation above <role> can be role name or role mention\n"
	rv += "<permission> is one of the following:\n"
	rv += "-- \"SubEditUser\" (\"su\") - lets role members edit users and characters in their subguild (and all guilds under it)"
	rv += "-- \"SubEditGuild\" (\"sg\") - lets role members edit structure of their subguild (and all guilds under it)"
	rv += "-- \"OneUpEditUser\" (\"uu\") - lets role members edit users and characters of a subguild above their (and all under)"
	rv += "-- \"OneUpEditGuild\" (\"ug\") - lets role members edit structure of a subguild above their (and all under)"
	rv += "-- \"GuildEditUser\" (\"gu\") - lets role members edit users and characters of the entire guild"
	rv += "-- \"GuildEditGuild\" (\"gg\") - lets role members edit structure of the entire guild"
	rv += "Notice that last two permissions grant group-wide operations access. Like this one."

	m.SendMessage(rv)
}
