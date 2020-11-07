package helpers

import (
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/utility"
)

type AdminUserProcessor struct {
	prov  database.DataProvider
	funcs map[string]func(message.Message)
}

func NewAdminUserProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminUserProcessor{prov: prov}
	ap.funcs = map[string]func(message.Message){
		"h":        ap.help,
		"help":     ap.help,
		"r":        ap.register,
		"register": ap.register,
	}
	return ap
}

func (ap *AdminUserProcessor) ProcessMessage(m message.Message) {
	cmd := m.CurSegment()

	f, ok := ap.funcs[cmd]
	if !ok {
		rv := fmt.Sprintf("Unknown user administration command \"%v\". Send \"!g admin user help\" or \"!g a u h\" for help", m.FullMessage())
		go m.SendMessage(&rv)
		return
	}

	f(m)
}

func (ap *AdminUserProcessor) register(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		rv := "Some error happened while getting permissions: " + err.Error()
		go m.SendMessage(&rv)
		return
	}

	if perm&database.CharsPermissions == 0 {
		rv := "I'm sorry, but you don't have permissions to register users"
		go m.SendMessage(&rv)
		return
	}

	u := m.CurSegment()

	if m.MoreSegments() {
		rv := "Unknown extended parameter: \"" + m.CurSegment() + "\".\nWhy did you add it?"
		go m.SendMessage(&rv)
		return
	}

	if u == "all" {
		if perm&database.EditGuildCharsPerm == 0 {
			rv := "I'm sorry, but you don't have permissions to run guild-wide user management operations"
			go m.SendMessage(&rv)
			return
		}

		ap.syncAllUsers(m, false)
		return
	}

	if len(m.Mentions()) != 1 {
		rv := "Wrong command format. You should mention user for registration"
		go m.SendMessage(&rv)
		return
	}

	uid, err := utility.ParseUserMention(u)
	if err != nil {
		rv := "Error: " + err.Error()
		go m.SendMessage(&rv)
		return
	}

	if uid != m.Mentions()[0] {
		rv := "Error: You're doing something tricky. Mention is inconsistent. Try again, please"
		go m.SendMessage(&rv)
		return
	}

	guild, dbErr := ap.prov.GetGuildD(m.GuildId())
	if dbErr != nil {
		rv := "Error: " + dbErr.Error()
		go m.SendMessage(&rv)
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
		go m.SendMessage(&rv)
		return
	}

	rv := "User <@!" + uid + "> successfully registered"
	go m.SendMessage(&rv)
}

func (ap *AdminUserProcessor) syncAllUsers(m message.Message, delete bool) {
	guild, dbErr := ap.prov.GetGuildD(m.GuildId())
	if dbErr != nil {
		rv := "Error: " + dbErr.Error()
		go m.SendMessage(&rv)
		return
	}

	guildies, err := m.GuildMembers()
	if err != nil {
		rv := "Error getting guild memebers: " + err.Error()
		go m.SendMessage(&rv)
		return
	}

	rv := ""

	if delete {
		tmp, err := ap.deleteSyncAllUsers(guildies)
		if err != nil {
			rv := "Error getting guild memebers: " + err.Error()
			go m.SendMessage(&rv)
			return
		}

		rv += tmp + "\n"
	}

	rv += "Users registered:\n"
	for id, nick := range guildies {
		dbu, dbErr := ap.prov.GetUserD(id)
		if dbErr == nil {
			continue
		} else if dbErr.Code != database.UserNotFound {
			rv += "\nError while adding users. Error: \n" + dbErr.Error() + "\nPlease, run the command again to retry"
			go m.SendMessage(&rv)
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

		dbu, dbErr = ap.prov.AddUser(dbu, dbgp)
		if dbErr != nil {
			rv += "\nError while adding users. Error: \n" + dbErr.Error() + "\nPlease, run the command again to retry"
			go m.SendMessage(&rv)
			return
		}
		rv += nick + ", "
	}

	go m.SendMessage(&rv)
}

func (ap *AdminUserProcessor) deleteSyncAllUsers(guildies map[string]string) (string, error) {
	return "", nil
}

func (ap *AdminUserProcessor) help(m message.Message) {
	rv := "Here's a list of user management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		go m.SendMessage(&rv)
		return
	}

	if perm&database.CharsPermissions == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		go m.SendMessage(&rv)
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

	go m.SendMessage(&rv)
}
