package message

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
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

func (dgm *DiscordGoMessage) AuthorId() string {
	return dgm.orig.Author.ID
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

func (dgm *DiscordGoMessage) GuildMembersWithRole(r string) (map[string]string, error) {
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
				found := false
				for _, rid := range m.Roles {
					if rid == r {
						found = true
						break
					}
				}

				if found {
					guildies[m.User.ID] = m.Nick
				}
			}
		}

		dgm.members = guildies
	}

	return dgm.members, nil
}

func (dgm *DiscordGoMessage) CurSegment() string {
	var rv string
	for rv == "" && dgm.curMsg != "" {
		rv, dgm.curMsg = utility.NextCommand(&dgm.curMsg)
	}

	return rv
}

func (dgm *DiscordGoMessage) PeekSegment() string {
	var rv string
	tmp := dgm.curMsg
	for rv == "" && dgm.curMsg != "" {
		rv, tmp = utility.NextCommand(&tmp)
	}

	return rv
}

func (dgm *DiscordGoMessage) LeftOverSegments() string {
	return dgm.curMsg
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

func (dgm *DiscordGoMessage) GetRoleId(name string) (string, error) {
	rs, err := dgm.session.GuildRoles(dgm.orig.GuildID)
	if err != nil {
		return "", err
	}

	for _, r := range rs {
		if r.Name == name {
			return r.ID, nil
		}
	}

	return "nil", errors.New(fmt.Sprintf("Role with name %v was not found", name))
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

func (dgm *DiscordGoMessage) CheckGuildModificationPermissions(gid uuid.UUID) (bool, error) {
	var err error

	perm, err := dgm.AuthorPermissions()
	if err != nil {
		return false, err
	}

	if perm&database.EditGuildCharsPerm != 0 {
		return true, nil
	}

	auth, err := dgm.Author()
	if err != nil {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	gper, ok := auth.Guilds[dgm.GuildId()]
	if !ok {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	ok, err = utility.ValidateGuildAccess(dgm.prov, gper, gid)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (dgm *DiscordGoMessage) CheckUserModificationPermissions(uid string) (bool, error) {
	if uid == dgm.AuthorId() {
		return true, nil
	}

	var err error
	trgUser, err := dgm.prov.GetUserD(uid)
	if err != nil {
		return false, errors.New("User you're trying to modify doesn't seem to be a part of this guild")
	}

	trgPerm, ok := trgUser.Guilds[dgm.GuildId()]
	if !ok {
		return false, errors.New("User you're trying to modify doesn't seem to be a part of this guild")
	}

	perm, err := dgm.AuthorPermissions()
	if err != nil {
		return false, err
	}

	if perm&database.EditGuildCharsPerm != 0 {
		return true, nil
	}

	auth, err := dgm.Author()
	if err != nil {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	gper, ok := auth.Guilds[dgm.GuildId()]
	if !ok {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	ok, err = utility.ValidateUserAccess(dgm.prov, gper, trgPerm.GuildId)
	if err != nil {
		return false, err
	}

	return ok, nil
}
