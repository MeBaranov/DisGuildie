package utility

import (
	"strings"
	"time"

	"github.com/mebaranov/disguildie/database"

	"github.com/bwmarrin/discordgo"
)

const charLimit = 2000
const delay = 50 * time.Millisecond
const longDelay = 1 * time.Second
const limit = 5

var SuperUserID = ""

func NextCommand(in *string) (cmd string, obj string) {
	pos := strings.IndexByte(*in, ' ')
	if pos < 0 {
		cmd = *in
		obj = ""
		return
	}

	cmd = (*in)[:pos]
	obj = (*in)[pos+1:]
	return
}

func sendMonitored(s *discordgo.Session, c *string, msg *string) {
	if len(*msg) < charLimit {
		s.ChannelMessageSend(*c, *msg)
		return
	}
	split := strings.Split(*msg, "\n")
	l, count, cur := 0, 0, ""
	for _, str := range split {
		if l+len(str) >= charLimit {
			s.ChannelMessageSend(*c, cur)
			count += 1
			if count >= limit {
				count = 0
				time.Sleep(longDelay)
			}
			time.Sleep(delay)

			cur = str
			l = len(str)
		} else {
			cur += "\n" + str
			l += 1 + len(str)
		}
	}

	if cur != "" {
		s.ChannelMessageSend(*c, cur)
	}
}

func SendMonitored(s *discordgo.Session, c *string, msg *string) {
	go sendMonitored(s, c, msg)
}

func GetPermissions(s *discordgo.Session, mc *discordgo.MessageCreate, prov database.DataProvider) (int, error) {
	gld, err := s.Guild(mc.GuildID)
	if err != nil {
		return 0, err
	}

	uid := mc.Message.Author.ID
	if uid == gld.OwnerID || uid == SuperUserID {
		return database.FullPermissions, nil
	}

	m, err := prov.GetMoney(mc.GuildID)
	if err != nil {
		return 0, err
	}

	if m.UserId == uid {
		return database.FullPermissions, nil
	}

	u, err := prov.GetUserD(uid)
	if err != nil {
		return 0, err
	}
	if gp, ok := u.Guilds[mc.GuildID]; ok {
		return gp.Permissions, nil
	}

	return 0, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in this guild"}
}
