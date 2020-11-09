package message

import (
	"fmt"

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
	superUser         *string
	money             *database.Money
	author            *database.User
}

func New(s *discordgo.Session, mc *discordgo.MessageCreate, prov database.DataProvider, superUser *string) Message {

	return &DiscordGoMessage{
		session:   s,
		orig:      mc.Message,
		prov:      prov,
		curMsg:    mc.Message.Content,
		superUser: superUser,
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

func (dgm *DiscordGoMessage) Author() (*database.User, error) {
	if dgm.author == nil {
		a, err := dgm.prov.GetUserD(dgm.orig.Author.ID)
		if err != nil {
			return nil, err
		}
		dgm.author = a
	}

	return dgm.author, nil
}

func (dgm *DiscordGoMessage) AuthorPermissions() (int, error) {
	if dgm.authorPermissions == nil {
		rv, err := dgm.getPermissions()
		if err != nil {
			return rv, err
		}
		dgm.authorPermissions = &rv
	}

	return *dgm.authorPermissions, nil
}

func (dgm *DiscordGoMessage) Money() (*database.Money, error) {
	if dgm.money == nil {
		m, err := dgm.prov.GetMoney(dgm.orig.GuildID)
		if err != nil {
			return nil, err
		}
		dgm.money = m
	}

	return dgm.money, nil
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

func (dgm *DiscordGoMessage) SendMessage(s string, strs ...interface{}) {
	msg := fmt.Sprintf(s, strs...)
	go utility.SendMonitored(dgm.session, &dgm.orig.ChannelID, &msg)
}

func (dgm *DiscordGoMessage) UserRoles(id string) ([]string, error) {
	m, err := dgm.session.GuildMember(dgm.orig.GuildID, id)
	if err != nil {
		return nil, err
	}

	return m.Roles, nil
}

func (dgm *DiscordGoMessage) getPermissions() (int, error) {
	gld, err := dgm.session.Guild(dgm.orig.GuildID)
	if err != nil {
		return 0, err
	}

	uid := dgm.orig.Author.ID
	if uid == gld.OwnerID || (dgm.superUser != nil && uid == *(dgm.superUser)) {
		return database.FullPermissions, nil
	}

	m, err := dgm.Money()
	if err != nil {
		return 0, err
	}

	if m.UserId == uid {
		return database.FullPermissions, nil
	}

	u, err := dgm.Author()
	if err != nil {
		return 0, err
	}
	if gp, ok := u.Guilds[dgm.orig.GuildID]; ok {
		return gp.Permissions, nil
	}

	return 0, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in this guild"}
}
