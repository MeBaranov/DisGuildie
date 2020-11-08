package helpers

import (
	"errors"

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
		"s":        ap.sync,
		"sync":     ap.sync,
		"remove":   ap.remove,
		"cleanup":  ap.cleanup,
	}
	return ap
}

func (ap *AdminUserProcessor) ProcessMessage(m message.Message) {
	cmd := m.CurSegment()

	f, ok := ap.funcs[cmd]
	if !ok {
		go m.SendMessage("Unknown user administration command \"%v\". Send \"!g admin user help\" or \"!g a u h\" for help", m.FullMessage())
		return
	}

	f(m)
}

const (
	register = iota
	sync
)

func (ap *AdminUserProcessor) register(m message.Message) {
	ap.regOrSync(m, register)
}

func (ap *AdminUserProcessor) sync(m message.Message) {
	ap.regOrSync(m, sync)
}

func (ap *AdminUserProcessor) remove(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		go m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.CharsPermissions == 0 {
		go m.SendMessage("I'm sorry, but you don't have permissions to delete users")
		return
	}

	u := m.CurSegment()

	if m.MoreSegments() {
		go m.SendMessage("Unknown extended parameter: '%v'", m.CurSegment())
		return
	}

	if len(m.Mentions()) != 1 {
		go m.SendMessage("Wrong command format. You should mention user to delete")
		return
	}

	uid, err := utility.ParseUserMention(u)
	if err != nil {
		go m.SendMessage("Error: %v", err.Error())
		return
	}

	if uid != m.Mentions()[0] {
		go m.SendMessage("Error: You're doing something tricky. Mention is inconsistent. Try again, please")
		return
	}

	err = ap.removeUser(uid, m.GuildId())
	if err != nil {
		go m.SendMessage("Could not remove user: %v", err.Error())
		return
	}

	go m.SendMessage("User <@!%v> successfully removed", uid)
}

func (ap *AdminUserProcessor) cleanup(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		go m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildCharsPerm == 0 {
		go m.SendMessage("I'm sorry, but you don't have permissions to run guild-wide user management operations")
		return
	}

	if m.MoreSegments() {
		go m.SendMessage("Unknown extended parameter: '%v'", m.CurSegment())
		return
	}

	guildies, err := m.GuildMembers()
	if err != nil {
		go m.SendMessage("Could not get guild members: '%v'", m.CurSegment())
		return
	}

	registered, err := ap.prov.GetUsersInGuild(m.GuildId())
	if err != nil {
		go m.SendMessage("Could not get registered members: '%v'", m.CurSegment())
		return
	}

	count := 0
	for _, u := range registered {
		if _, ok := guildies[u.DiscordId]; !ok {
			if err = ap.removeUser(u.DiscordId, m.GuildId()); err != nil {
				go m.SendMessage("Error while deleting user: %v. Try running the command again.\nCleaned up so far: %v", err.Error(), count)
				return
			}
			count += 1
		}
	}

	go m.SendMessage("Cleaned up %v users", count)
}

func (ap *AdminUserProcessor) help(m message.Message) {
	rv := "Here's a list of user management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		go m.SendMessage(rv)
		return
	}

	if perm&database.CharsPermissions == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		go m.SendMessage(rv)
		return
	}

	rv += "\t -- \"!g admin user register <mention user>\" (\"!g a u r <user>\") - Register user in the system\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user register all\" (\"!g a u r all\") - Register all users from guild in the system\n"
	}
	rv += "\t -- \"!g admin user remove <mention user>\" (\"!g a u remove <user>\") - Remove user from the system\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user cleanup\" (\"!g a u cleanup\") - Cleanup all users that are not in the channel anymore\n"
	}
	rv += "\t -- \"!g admin user assign <mention user> <sub-guild name>\" (\"!g a u a <user> <name>\") - Move user to a sub-guild\n"
	rv += "\t -- \"!g admin user sync <mention user>\" (\"!g a u p <user>\") - Synchronize user permissions\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user sync all\" (\"!g a u s all\") - Synchronize all users permissions\n"
	}

	go m.SendMessage(rv)
}

func (ap *AdminUserProcessor) regOrSync(m message.Message, action int) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		go m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.CharsPermissions == 0 {
		go m.SendMessage("I'm sorry, but you don't have permissions to register or syncronize users")
		return
	}

	u := m.CurSegment()

	if m.MoreSegments() {
		go m.SendMessage("Unknown extended parameter: '%v'", m.CurSegment())
		return
	}

	if u == "all" {
		if perm&database.EditGuildCharsPerm == 0 {
			go m.SendMessage("I'm sorry, but you don't have permissions to run guild-wide user management operations")
			return
		}

		switch action {
		case register:
			ap.registerAllUsers(m)
		case sync:
			ap.syncAllUsers(m)
		}
		return
	}

	if len(m.Mentions()) != 1 {
		go m.SendMessage("Wrong command format. You should mention user to register/sync")
		return
	}

	uid, err := utility.ParseUserMention(u)
	if err != nil {
		go m.SendMessage("Error: %v", err.Error())
		return
	}

	if uid != m.Mentions()[0] {
		go m.SendMessage("Error: You're doing something tricky. Mention is inconsistent. Try again, please")
		return
	}

	guild, dbErr := ap.prov.GetGuildD(m.GuildId())
	if dbErr != nil {
		go m.SendMessage("Error: %v", dbErr.Error())
		return
	}

	switch action {
	case register:
		err = ap.reigsterUser(uid, guild, m)
	case sync:
		err = ap.syncUserId(uid, guild, m)
	}
	if err != nil {
		go m.SendMessage("Could not register user: %v", err.Error())
		return
	}

	go m.SendMessage("User <@!%v> successfully registered", uid)
}

func (ap *AdminUserProcessor) registerAllUsers(m message.Message) {
	guild, dbErr := ap.prov.GetGuildD(m.GuildId())
	if dbErr != nil {
		go m.SendMessage("Error: %v", dbErr.Error())
		return
	}

	guildies, err := m.GuildMembers()
	if err != nil {
		go m.SendMessage("Error getting guild memebers: %v", err.Error())
		return
	}

	rv := "Users registered:\n"
	for id, nick := range guildies {
		if err = ap.reigsterUser(id, guild, m); err != nil {
			rv += "\nError while adding users. Error: \n" + err.Error() + "\nPlease, run the command again to retry"
			go m.SendMessage(rv)
			return
		}

		rv += nick + ", "
	}

	go m.SendMessage(rv)
}

func (ap *AdminUserProcessor) syncAllUsers(m message.Message) {
	guild, dbErr := ap.prov.GetGuildD(m.GuildId())
	if dbErr != nil {
		go m.SendMessage("Error: %v", dbErr.Error())
		return
	}

	guildies, dbErr := ap.prov.GetUsersInGuild(guild.DiscordId)
	if dbErr != nil {
		go m.SendMessage("Error getting guild memebers: %v", dbErr.Error())
		return
	}

	for _, u := range guildies {
		if err := ap.syncUser(u, guild, m); err != nil {
			go m.SendMessage("\nError while adding users: %v\nPlease, run the command again to retry", err.Error())
			return
		}
	}

	go m.SendMessage("All users permissions syncronized")
}

func (ap *AdminUserProcessor) reigsterUser(id string, guild *database.Guild, m message.Message) error {
	dbu, dbErr := ap.prov.GetUserD(id)
	if dbErr == nil {
		if _, ok := dbu.Guilds[guild.DiscordId]; ok {
			return nil
		}
	} else if dbErr.Code != database.UserNotFound {
		return dbErr
	} else {
		dbu = &database.User{
			DiscordId: id,
		}
	}

	p, err := ap.userPermissions(id, m)
	if err != nil {
		return err
	}

	dbgp := &database.GuildPermission{
		Permissions: p,
		GuildId:     guild.GuildId,
		TopGuild:    guild.DiscordId,
	}

	dbu, dbErr = ap.prov.AddUser(dbu, dbgp)
	if dbErr != nil {
		return dbErr
	}

	return nil
}

func (ap *AdminUserProcessor) syncUserId(id string, guild *database.Guild, m message.Message) error {
	dbu, dbErr := ap.prov.GetUserD(id)
	if dbErr != nil {
		return dbErr
	}

	return ap.syncUser(dbu, guild, m)
}

func (ap *AdminUserProcessor) syncUser(dbu *database.User, guild *database.Guild, m message.Message) error {
	uperms, ok := dbu.Guilds[guild.DiscordId]
	if !ok {
		return errors.New("User is not registered in the guild")
	}

	p, err := ap.userPermissions(dbu.DiscordId, m)
	if err != nil {
		return err
	}

	if uperms.Permissions == p {
		return nil
	}
	uperms.Permissions = p

	dbu, dbErr := ap.prov.SetUserPermissions(dbu.DiscordId, uperms)
	if dbErr != nil {
		return dbErr
	}

	return nil
}

func (ap *AdminUserProcessor) removeUser(id string, guildId string) error {
	dbu, err := ap.prov.GetUserD(id)
	if err != nil {
		return err
	}

	if _, ok := dbu.Guilds[guildId]; !ok {
		return nil
	}

	_, err = ap.prov.RemoveUserD(id, guildId)
	return err
}

func (ap *AdminUserProcessor) userPermissions(id string, m message.Message) (int, error) {
	roles, err := m.UserRoles(id)
	if err != nil {
		return 0, err
	}

	rv := 0
	for _, r := range roles {
		p, err := ap.rolePermission(m.GuildId(), r)
		if err != nil {
			return 0, err
		}
		rv |= p
	}

	return rv, nil
}

func (ap *AdminUserProcessor) rolePermission(g string, r string) (int, error) {
	role, err := ap.prov.GetRole(g, r)
	if err != nil {
		if err.Code == database.RoleNotFound {
			return 0, nil
		}
		return 0, err
	}

	return role.Permissions, nil
}
