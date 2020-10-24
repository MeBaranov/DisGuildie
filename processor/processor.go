package processor

import (
	"github.com/bwmarrin/discordgo"

	"github.com/mebaranov/disguildie/processor/helpers"
	"github.com/mebaranov/disguildie/utility"
)

type MessageProcessor interface {
	ProcessMessage(s *discordgo.Session, c *string, m *string, mc *discordgo.MessageCreate)
}

var availableCommands = map[string]MessageProcessor{
	"!help": &helpers.HelpProcessor{},
}

type Processor struct{}

func (proc *Processor) ProcessMessage(s *discordgo.Session, c *string, m *string, mc *discordgo.MessageCreate) {
	cmd, obj := utility.NextCommand(m)
	p, ok := availableCommands[cmd]
	if !ok {
		return
	}

	p.ProcessMessage(s, c, &obj, mc)
}
