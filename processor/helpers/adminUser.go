package helpers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/utility"
)

type AdminUserProcessor struct {
	prov  database.DataProvider
	funcs map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate)
}

func NewAdminUserProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminUserProcessor{prov: prov}
	ap.funcs = map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate){
		"h":        ap.help,
		"help":     ap.help,
		"r":        ap.register,
		"register": ap.register,
	}
	return ap
}

func (ap *AdminUserProcessor) ProcessMessage(s *discordgo.Session, m *string, mc *discordgo.MessageCreate) {
	cmd, obj := utility.NextCommand(m)

	f, ok := ap.funcs[cmd]
	if !ok {
		rv := fmt.Sprintf("Unknown user administration command \"%v\". Send \"!g admin user help\" or \"!g a u h\" for help", mc.Message.Content)
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	f(s, &obj, mc)
}

func (ap *AdminUserProcessor) register(s *discordgo.Session, m *string, mc *discordgo.MessageCreate) {
	perm, err := utility.GetPermissions(s, mc, ap.prov)
	if err != nil {
		rv := "Some error happened while getting permissions: " + err.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	if perm&database.CharsPermissions == 0 {
		rv := "I'm sorry, but you don't have permissions to register users"
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	u, extra := utility.NextCommand(m)

	if extra != "" {
		rv := "Unknown extended parameter: \"" + extra + "\".\nWhy did you add it?"
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	if u == "all" {
		if perm&database.EditGuildCharsPerm == 0 {
			rv := "I'm sorry, but you don't have permissions to run guild-wide user management operations"
			go utility.SendMonitored(s, &mc.ChannelID, &rv)
			return
		}

		ap.syncAllUsers(s, mc, false)
		return
	}

	if len(mc.Message.Mentions) != 1 {
		rv := "Wrong command format. You should mention user for registration"
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	uid, err := utility.ParseUserMention(u)
	if err != nil {
		rv := "Error: " + err.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	if uid != mc.Message.Mentions[0].ID {
		rv := "Error: You're doing something tricky. Mention is inconsistent. Try again, please"
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	guild, dbErr := ap.prov.GetGuildD(mc.Message.GuildID)
	if dbErr != nil {
		rv := "Error: " + dbErr.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	dbu := &database.User{
		DiscordId: uid,
	}
	dbgp := &database.GuildPermission{
		Permissions: 0,
		GuildId:     guild.GuildId,
		TopGuild:    guild.DiscordId,
	}
	dbu, dbErr = ap.prov.AddUser(dbu, dbgp)
	if dbErr != nil {
		rv := "Error: " + dbErr.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	rv := "User <@!" + uid + "> successfully registered"
	go utility.SendMonitored(s, &mc.ChannelID, &rv)
}

func (ap *AdminUserProcessor) syncAllUsers(s *discordgo.Session, mc *discordgo.MessageCreate, delete bool) {
	guild, err := ap.prov.GetGuildD(mc.GuildID)
	if err != nil {
		rv := "Error: " + err.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	guildies := make(map[string]string)

	cur := ""
	for count := 1000; count == 100; {
		gld, err := s.GuildMembers(mc.GuildID, cur, 1000)
		if err != nil {
			rv := "Error getting guild memebers: " + err.Error()
			go utility.SendMonitored(s, &mc.ChannelID, &rv)
			return
		}

		count = len(gld)
		for _, m := range gld {
			guildies[m.User.ID] = m.Nick
		}
	}

	rv := ""

	if delete {
		tmp, err := ap.deleteSyncAllUsers(guildies)
		if err != nil {
			rv := "Error getting guild memebers: " + err.Error()
			go utility.SendMonitored(s, &mc.ChannelID, &rv)
			return
		}

		rv += tmp + "\n"
	}

	rv += "Users registered:\n"
	for id, nick := range guildies {
		dbu, err := ap.prov.GetUserD(id)
		if err == nil {
			continue
		} else if err.Code != database.UserNotFound {
			rv += "\nError while adding users. Error: \n" + err.Error() + "\nPlease, run the command again to retry"
			go utility.SendMonitored(s, &mc.ChannelID, &rv)
			return
		}

		dbu = &database.User{
			DiscordId: id,
		}

		// TODO: add roles processing here too

		dbgp := &database.GuildPermission{
			Permissions: 0,
			GuildId:     guild.GuildId,
			TopGuild:    guild.DiscordId,
		}

		dbu, err = ap.prov.AddUser(dbu, dbgp)
		if err != nil {
			rv += "\nError while adding users. Error: \n" + err.Error() + "\nPlease, run the command again to retry"
			go utility.SendMonitored(s, &mc.ChannelID, &rv)
			return
		}
		rv += nick + ", "
	}

	go utility.SendMonitored(s, &mc.ChannelID, &rv)
}

func (ap *AdminUserProcessor) deleteSyncAllUsers(guildies map[string]string) (string, error) {
	return "", nil
}

func (ap *AdminUserProcessor) help(s *discordgo.Session, _ *string, mc *discordgo.MessageCreate) {
	rv := "Here's a list of user management commands you're allowed to use:\n"

	perm, err := utility.GetPermissions(s, mc, ap.prov)
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	if perm&database.CharsPermissions == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	rv += "\t -- \"!g admin user register <mention user>\" (\"!g a u r <user>\") - Register user in the system\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user register all\" (\"!g a u r all\") - Register all users from guild in the system\n"
	}
	rv += "\t -- \"!g admin user remove <mention user>\" (\"!g a u remove <user>\") - Remove user from the system\n"
	rv += "\t -- \"!g admin user remove <Discord ID>\" (\"!g a u remove <Discord ID>\") - Remove user from the system by discord ID\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user sync\" (\"!g a u s\") - Register all users in guild in the system. And remove all users no longer in guild\n"
	}
	rv += "\t -- \"!g admin user assign <mention user> <sub-guild name>\" (\"!g a u a <user> <name>\") - Move user to a sub-guild\n"
	rv += "\t -- \"!g admin user permissions <mention user>\" (\"!g a u p <user>\") - Re-synchronize user roles and permissions\n"

	go utility.SendMonitored(s, &mc.ChannelID, &rv)
}
