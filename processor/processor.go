package processor

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mebaranov/disguildie/database"

	"github.com/mebaranov/disguildie/processor/helpers"
	"github.com/mebaranov/disguildie/utility"
)

type Processor struct {
	provider database.DataProvider
	funcs    map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate)
}

func New(prov database.DataProvider) helpers.MessageProcessor {
	admin := helpers.NewAdminProcessor(prov)
	proc := &Processor{provider: prov}

	proc.funcs = map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate){
		"help":  proc.help,
		"h":     proc.help,
		"admin": admin.ProcessMessage,
		"a":     admin.ProcessMessage,
	}
	return proc
}

func (proc *Processor) ProcessMessage(s *discordgo.Session, m *string, mc *discordgo.MessageCreate) {
	cmd, obj := utility.NextCommand(m)

	f, ok := proc.funcs[cmd]
	if !ok {
		rv := fmt.Sprintf("Unknown command \"%v\". Send \"!g help\" or \"!g h\" for help", mc.Message.Content)
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	f(s, &obj, mc)
}

func (proc *Processor) help(s *discordgo.Session, _ *string, mc *discordgo.MessageCreate) {
	rv := "Here is a list of commands you are allowed to use:\n"

	p, err := utility.GetPermissions(s, mc, proc.provider)
	if err != nil {
		rv += "Could not get your permissions. Error:\n" + err.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	if p > 0 {
		rv += "\t-- \"!g admin\" (\"!g a\") - administrative actions"
	}

	go utility.SendMonitored(s, &mc.ChannelID, &rv)
}
