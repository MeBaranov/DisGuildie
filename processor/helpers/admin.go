package helpers

import (
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
	ap.funcs = map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate){
		"h":    ap.adminHelp,
		"help": ap.adminHelp,
	}
	return ap
}

func (ap *AdminProcessor) ProcessMessage(s *discordgo.Session, m *string, mc *discordgo.MessageCreate) {
	cmd, obj := utility.NextCommand(m)
	_ = obj
	switch cmd {
	case "user":
		// change user/s
	case "":
	}
}

func (ap *AdminProcessor) adminHelp(s *discordgo.Session, _ *string, mc *discordgo.MessageCreate) {

}
