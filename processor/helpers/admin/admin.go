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

	ap.Funcs = map[string]func(message.Message) (string, error){
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

func (ap *AdminProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of administrative commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting permissions", err
	}

	if perm == 0 {
		return rv, nil
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

	return rv, nil
}
