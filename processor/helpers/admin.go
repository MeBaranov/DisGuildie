package helpers

import (
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
)

type AdminProcessor struct {
	prov  database.DataProvider
	funcs map[string]func(message.Message)
}

func NewAdminProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminProcessor{prov: prov}
	apu := NewAdminUserProcessor(prov)

	ap.funcs = map[string]func(message.Message){
		"h":    ap.help,
		"help": ap.help,
		"u":    apu.ProcessMessage,
		"user": apu.ProcessMessage,
	}
	return ap
}

func (ap *AdminProcessor) ProcessMessage(m message.Message) {
	cmd := m.CurSegment()

	f, ok := ap.funcs[cmd]
	if !ok {
		rv := fmt.Sprintf("Unknown admin command \"%v\". Send \"!g admin help\" or \"!g a h\" for help", m.FullMessage())
		go m.SendMessage(&rv)
		return
	}

	f(m)
}

func (ap *AdminProcessor) help(m message.Message) {
	rv := "Here's a list of administrative commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		go m.SendMessage(&rv)
		return
	}

	if perm == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		go m.SendMessage(&rv)
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

	go m.SendMessage(&rv)
}
