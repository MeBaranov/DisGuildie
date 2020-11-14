package admin

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type AdminProcessor struct {
	helpers.BaseMessageProcessor
}

func NewAdminProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &AdminProcessor{}
	apu := NewAdminUserProcessor(prov)
	apg := NewAdminGuildProcessor(prov)
	apr := NewAdminRoleProcessor(prov)
	aps := NewAdminStatsProcessor(prov)

	ap.Prov = prov

	// TODO: Check permissions in guild
	ap.Funcs = map[string]func(message.Message){
		"h":     ap.help,
		"help":  ap.help,
		"u":     apu.ProcessMessage,
		"user":  apu.ProcessMessage,
		"g":     apg.ProcessMessage,
		"guild": apg.ProcessMessage,
		"r":     apr.ProcessMessage,
		"role":  apr.ProcessMessage,
		"s":     aps.ProcessMessage,
		"stats": aps.ProcessMessage,
	}
	return ap
}

func (ap *AdminProcessor) help(m message.Message) {
	rv := "Here's a list of administrative commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		m.SendMessage(rv)
		return
	}

	if perm&database.StructurePermissions > 0 {
		rv += "\t-- \"!g admin guild\" (\"!g a g\") - subguilds management\n"
	}
	if perm&database.CharsPermissions > 0 {
		rv += "\t-- \"!g admin user\" (\"!g a u\") - users management\n"
	}
	if perm&database.EditGuildStructurePerm > 0 {
		rv += "\t-- \"!g admin stat\" (\"!g a s\") - stats management\n"
		rv += "\t-- \"!g admin role\" (\"!g a r\") - roles management\n"
	}

	m.SendMessage(rv)
}
