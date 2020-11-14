package utility

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"

	"github.com/bwmarrin/discordgo"
)

const charLimit = 2000
const delay = 50 * time.Millisecond
const longDelay = 1 * time.Second
const limit = 5

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

func IsUserMention(m string) bool {
	if len(m) < 4 || m[0] != '<' || m[1] != '@' || m[2] != '!' || m[len(m)-1] != '>' {
		return false
	}
	return true
}

func ParseUserMention(m string) (string, error) {
	if !IsUserMention(m) {
		return "", errors.New("Wrong format for user name")
	}

	return m[3 : len(m)-1], nil
}

func ParseRoleMention(m string) (string, error) {
	if len(m) < 4 || m[0] != '<' || m[1] != '@' || m[2] != '&' || m[len(m)-1] != '>' {
		return "", errors.New("Wrong format for a role")
	}

	return m[3 : len(m)-1], nil
}

func ValidateUserAccess(prov database.DataProvider, from *database.GuildPermission, subGuild uuid.UUID) (bool, error) {
	if from.Permissions&database.CharsPermissions == 0 {
		return false, nil
	}

	if from.Permissions&database.EditGuildCharsPerm != 0 {
		return true, nil
	}

	g := from.GuildId
	if from.Permissions&database.EditOneUpCharsPerm != 0 {
		guild, err := prov.GetGuild(g)
		if err != nil {
			return false, err
		}

		if guild.DiscordId == "" {
			g = guild.ParentId
		}
	}

	return CheckIfSubguild(prov, g, subGuild)
}

func ValidateGuildAccess(prov database.DataProvider, from *database.GuildPermission, subGuild uuid.UUID) (bool, error) {
	if from.Permissions&database.StructurePermissions == 0 {
		return false, nil
	}

	if from.Permissions&database.EditGuildStructurePerm != 0 {
		return true, nil
	}

	g := from.GuildId
	if from.Permissions&database.EditOneUpStructurePerm != 0 {
		guild, err := prov.GetGuild(g)
		if err != nil {
			return false, err
		}

		if guild.DiscordId == "" {
			g = guild.ParentId
		}
	}

	return CheckIfSubguild(prov, g, subGuild)
}

func CheckIfSubguild(prov database.DataProvider, g uuid.UUID, sub uuid.UUID) (bool, error) {
	for sub != g {
		sg, err := prov.GetGuild(sub)
		if err != nil {
			return false, err
		}

		if sg.DiscordId != "" {
			return false, nil
		}

		sub = sg.ParentId
	}

	return true, nil
}
