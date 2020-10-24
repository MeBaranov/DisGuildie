package helpers

import (
	"github.com/bwmarrin/discordgo"

	"github.com/mebaranov/disguildie/utility"
)

type AdminProcessor struct{}

func (ap *AdminProcessor) ProcessMessage(s *discordgo.Session, c *string, m *string, mc *discordgo.MessageCreate) {
	cmd, obj := utility.NextCommand(m)
	_ = cmd
	_ = obj
	switch cmd {
	case "user":
		// change user/s
	case "":
	}
}
