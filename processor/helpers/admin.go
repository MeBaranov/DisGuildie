package helpers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mebaranov/disguildie/database"

	"github.com/mebaranov/disguildie/utility"
)

type AdminProcessor struct {
	prov  database.DataProvider
	funcs map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate)
}

func NewAdminProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminProcessor{prov: prov}
	apu := NewAdminUserProcessor(prov)

	ap.funcs = map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate){
		"h":    ap.help,
		"help": ap.help,
		"u":    apu.ProcessMessage,
		"user": apu.ProcessMessage,
	}
	return ap
}

func (ap *AdminProcessor) ProcessMessage(s *discordgo.Session, m *string, mc *discordgo.MessageCreate) {
	cmd, obj := utility.NextCommand(m)

	f, ok := ap.funcs[cmd]
	if !ok {
		rv := fmt.Sprintf("Unknown admin command \"%v\". Send \"!g admin help\" or \"!g a h\" for help", mc.Message.Content)
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	f(s, &obj, mc)
}

func (ap *AdminProcessor) help(s *discordgo.Session, _ *string, mc *discordgo.MessageCreate) {
	rv := "Here's a list of administrative commands you're allowed to use:\n"

	perm, err := utility.GetPermissions(s, mc, ap.prov)
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	if perm == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
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

	go utility.SendMonitored(s, &mc.ChannelID, &rv)
}
