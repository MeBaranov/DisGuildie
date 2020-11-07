package message

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/utility"
)

type DiscordGoMessage struct {
	mentions          []string
	authorPermissions *int
	members           map[string]string
	session           *discordgo.Session
	orig              *discordgo.Message
	prov              database.DataProvider
	curMsg            string
}

func New(s *discordgo.Session, mc *discordgo.MessageCreate, prov database.DataProvider) Message {

	return &DiscordGoMessage{
		session: s,
		orig:    mc.Message,
		prov:    prov,
		curMsg:  mc.Message.Content,
	}
}

func (dgm *DiscordGoMessage) GuildId() string {
	return dgm.orig.GuildID
}

func (dgm *DiscordGoMessage) ChannelId() string {
	return dgm.orig.ChannelID
}

func (dgm *DiscordGoMessage) FullMessage() string {
	return dgm.orig.Content
}

func (dgm *DiscordGoMessage) Author() string {
	return dgm.orig.Author.ID
}

func (dgm *DiscordGoMessage) AuthorPermissions() (int, error) {
	if dgm.authorPermissions == nil {
		rv, err := utility.GetPermissions(dgm.session, dgm.orig, dgm.prov)
		if err != nil {
			return rv, err
		}
		dgm.authorPermissions = &rv
	}

	return *dgm.authorPermissions, nil
}

func (dgm *DiscordGoMessage) Mentions() []string {
	if dgm.mentions == nil {
		dgm.mentions = make([]string, len(dgm.orig.Mentions), len(dgm.orig.Mentions))
		for i, u := range dgm.orig.Mentions {
			dgm.mentions[i] = u.ID
		}
	}

	return dgm.mentions
}

func (dgm *DiscordGoMessage) GuildMembers() (map[string]string, error) {
	if dgm.members == nil {
		guildies := make(map[string]string)

		cur := ""
		for count := 1000; count == 100; {
			gld, err := dgm.session.GuildMembers(dgm.orig.GuildID, cur, 1000)
			if err != nil {
				return nil, err
			}

			count = len(gld)
			for _, m := range gld {
				guildies[m.User.ID] = m.Nick
			}
		}

		dgm.members = guildies
	}

	return dgm.members, nil
}

func (dgm *DiscordGoMessage) CurSegment() string {
	var rv string
	rv, dgm.curMsg = utility.NextCommand(&dgm.curMsg)

	return rv
}

func (dgm *DiscordGoMessage) MoreSegments() bool {
	return dgm.curMsg == ""
}

func (dgm *DiscordGoMessage) SendMessage(s *string) {
	utility.SendMonitored(dgm.session, &dgm.orig.ChannelID, s)
}
