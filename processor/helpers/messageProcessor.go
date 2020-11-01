package helpers

import "github.com/bwmarrin/discordgo"

type MessageProcessor interface {
	ProcessMessage(s *discordgo.Session, m *string, mc *discordgo.MessageCreate)
}
