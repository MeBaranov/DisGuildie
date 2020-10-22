package processor

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/mebaranov/disguildie/processor/helpers"
	"github.com/mebaranov/disguildie/utility"
)

type MessageProcessor interface {
	ProcessMessage(s *discordgo.Session, c *string, m *string)
}

var availableCommands = map[string]MessageProcessor{
	"!help": &helpers.HelpProcessor{},
}

type Processor struct{}

func (proc *Processor) ProcessMessage(s *discordgo.Session, c *string, m *string) {
	if !strings.HasPrefix(*m, "!") {
		return
	}

	cmd, obj := utility.NextCommand(m)
	s.ChannelMessageSend(*c, "CMD:"+cmd+"\n OBJ:"+obj)
	p, ok := availableCommands[cmd]
	if !ok {
		return
	}

	p.ProcessMessage(s, c, &obj)
}
