package admin

import (
	"errors"
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
	"github.com/mebaranov/disguildie/utility"
)

type AdminUserProcessor struct {
	helpers.BaseMessageProcessor
}

func NewAdminUserProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &AdminUserProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message) (string, error){
		"h":        ap.help,
		"help":     ap.help,
		"r":        ap.register,
		"register": ap.register,
		"s":        ap.sync,
		"sync":     ap.sync,
		"a":        ap.assign,
		"assign":   ap.assign,
		"remove":   ap.remove,
		"cleanup":  ap.cleanup,
	}
	return ap
}

const (
	register = iota
	sync
)

func (ap *AdminUserProcessor) register(m message.Message) (string, error) {
	return ap.regOrSync(m, register)
}

func (ap *AdminUserProcessor) sync(m message.Message) (string, error) {
	return ap.regOrSync(m, sync)
}

func (ap *AdminUserProcessor) remove(m message.Message) (string, error) {
	u := m.CurSegment()

	if len(m.Mentions()) != 1 {
		return "", errors.New("Invalid command format")
	}

	uid, err := utility.ParseUserMention(u)
	if err != nil {
		return "parsing mention", err
	}

	ok, err := m.CheckUserModificationPermissions(uid)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to delete this user")
	}

	err = ap.removeUser(uid, m.GuildId())
	if err != nil {
		return "removing user", err
	}

	return fmt.Sprintf("User <@!%v> successfully removed", uid), nil
}

func (ap *AdminUserProcessor) cleanup(m message.Message) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.EditGuildCharsPerm == 0 {
		return "", errors.New("You don't have permissions to run guild-wide user management operations")
	}

	guildies, err := m.GuildMembers()
	if err != nil {
		return "getting guild memebers", err
	}

	registered, err := ap.Prov.GetUsersInGuild(m.GuildId())
	if err != nil {
		return "getting users in guild", err
	}

	count := 0
	for _, u := range registered {
		if _, ok := guildies[u.Id]; !ok {
			if err = ap.removeUser(u.Id, m.GuildId()); err != nil {
				return "deleting user", err
			}
			count += 1
		}
	}

	return fmt.Sprintf("Cleaned up %v users", count), nil
}

func (ap *AdminUserProcessor) assign(m message.Message) (string, error) {
	u := m.CurSegment()
	g := m.CurSegment()
	if u == "" || g == "" || len(m.Mentions()) != 1 {
		return "", errors.New("Invalid command format")
	}

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.EditGuildCharsPerm == 0 {
		return "", errors.New("You don't have permissions to run guild-wide user management operations")
	}

	uid, err := utility.ParseUserMention(u)
	if err != nil {
		return "parsing mention", err
	}

	guild, err := ap.Prov.GetGuildN(m.GuildId(), g)
	if err != nil {
		return "getting subguild", err
	}

	ok, err := m.CheckUserModificationPermissions(uid)
	if err != nil {
		return "checking source modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to assign this user")
	}

	ok, err = m.CheckGuildModificationPermissions(guild.GuildId)
	if err != nil {
		return "checking target modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to move users into this sub-guild")
	}

	_, err = ap.Prov.SetUserSubGuild(uid, &database.GuildPermission{TopGuild: m.GuildId(), GuildId: guild.GuildId})
	if err != nil {
		return "assigning user", err
	}

	return fmt.Sprintf("User <@!%v> assigned to guild %v", uid, g), nil
}

func (ap *AdminUserProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of user management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.CharsPermissions == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		return rv, nil
	}

	rv += "\t -- \"!g admin user register <mention user>\" (\"!g a u r <user>\") - Register user in the system\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user register all\" (\"!g a u r all\") - Register all users from guild in the system\n"
	}
	rv += "\t -- \"!g admin user remove <mention user>\" (\"!g a u remove <user>\") - Remove user from the system\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user cleanup\" (\"!g a u cleanup\") - Cleanup all users that are not in the channel anymore\n"
	}
	rv += "\t -- \"!g admin user assign <mention user> main\" (\"!g a u a <user> main\") - Move user to a top-level guild\n"
	rv += "\t -- \"!g admin user assign <mention user> <sub-guild name>\" (\"!g a u a <user> <name>\") - Move user to a sub-guild\n"
	rv += "\t -- \"!g admin user sync <mention user>\" (\"!g a u p <user>\") - Synchronize user permissions\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user sync all\" (\"!g a u s all\") - Synchronize all users permissions\n"
	}

	return rv, nil
}

func (ap *AdminUserProcessor) regOrSync(m message.Message, action int) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting permissions", err
	}

	if perm&database.CharsPermissions == 0 {
		return "", errors.New("You don't have permissions to register or syncronize users")
	}

	u := m.CurSegment()

	if u == "all" {
		if perm&database.EditGuildCharsPerm == 0 {
			return "", errors.New("You don't have permissions to run guild-wide user management operations")
		}

		switch action {
		case register:
			return ap.registerAllUsers(m)
		case sync:
			return ap.syncAllUsers(m)
		}
	}

	if len(m.Mentions()) != 1 {
		return "", errors.New("Invalid command format")
	}

	uid, err := utility.ParseUserMention(u)
	if err != nil {
		return "parsing mention", err
	}

	guild, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guilld", err
	}

	switch action {
	case register:
		err = ap.reigsterUser(uid, guild, m)
	case sync:
		err = ap.syncUserId(uid, guild, m)
	}

	if err != nil {
		return "registering/syncing user", err
	}

	return "User successfully registered/synced", nil
}

func (ap *AdminUserProcessor) registerAllUsers(m message.Message) (string, error) {
	var err error
	guild, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	guildies, err := m.GuildMembers()
	if err != nil {
		return "getting guild members", err
	}

	rv := "Users registered:\n"
	for id, nick := range guildies {
		if err = ap.reigsterUser(id, guild, m); err != nil {
			return "adding users", err
		}

		rv += nick + ", "
	}

	return rv, nil
}

func (ap *AdminUserProcessor) syncAllUsers(m message.Message) (string, error) {
	var err error
	guild, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	guildies, err := ap.Prov.GetUsersInGuild(guild.DiscordId)
	if err != nil {
		return "getting users in guild", err
	}

	for _, u := range guildies {
		if err := ap.syncUser(u, guild, m); err != nil {
			return "adding users", err
		}
	}

	return "All users permissions syncronized", nil
}

func (ap *AdminUserProcessor) reigsterUser(id string, guild *database.Guild, m message.Message) error {
	dbu, err := ap.Prov.GetUserD(id)
	if err == nil {
		if _, ok := dbu.Guilds[guild.DiscordId]; ok {
			return nil
		}
	} else {
		dbErr := database.ErrToDbErr(err)
		if dbErr == nil || dbErr.Code != database.UserNotFound {
			return err
		} else {
			dbu = &database.User{
				Id: id,
			}
		}
	}

	authorPerms, err := m.AuthorPermissions()
	if err != nil {
		return err
	}

	guildToAdd := guild.GuildId
	if authorPerms&database.EditGuildCharsPerm == 0 {
		auth, err := m.Author()
		if err != nil {
			return errors.New("You don't seem to be a part of this guild. Try again later.")
		}

		gper, ok := auth.Guilds[m.GuildId()]
		if !ok {
			return errors.New("You don't seem to be a part of this guild. Try again later.")
		}

		guildToAdd = gper.GuildId
	}

	p, err := ap.userPermissions(id, m)
	if err != nil {
		return err
	}

	dbgp := &database.GuildPermission{
		Permissions: p,
		GuildId:     guildToAdd,
		TopGuild:    guild.DiscordId,
	}

	dbu, err = ap.Prov.AddUser(dbu.Id, dbgp)
	if err != nil {
		return err
	}

	return nil
}

func (ap *AdminUserProcessor) syncUserId(id string, guild *database.Guild, m message.Message) error {
	dbu, err := ap.Prov.GetUserD(id)
	if err != nil {
		return err
	}

	return ap.syncUser(dbu, guild, m)
}

func (ap *AdminUserProcessor) syncUser(dbu *database.User, guild *database.Guild, m message.Message) error {
	uperms, ok := dbu.Guilds[guild.DiscordId]
	if !ok {
		return errors.New("User is not registered in the guild")
	}

	p, err := ap.userPermissions(dbu.Id, m)
	if err != nil {
		return err
	}

	if uperms.Permissions == p {
		return nil
	}
	uperms.Permissions = p

	dbu, err = ap.Prov.SetUserPermissions(dbu.Id, uperms)
	if err != nil {
		return err
	}

	return nil
}

func (ap *AdminUserProcessor) removeUser(id string, guildId string) error {
	dbu, err := ap.Prov.GetUserD(id)
	if err != nil {
		return err
	}

	if _, ok := dbu.Guilds[guildId]; !ok {
		return nil
	}

	_, err = ap.Prov.RemoveUserD(id, guildId)
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
	role, err := ap.Prov.GetRole(g, r)
	if err != nil {
		dbErr := database.ErrToDbErr(err)
		if dbErr != nil && dbErr.Code == database.RoleNotFound {
			return 0, nil
		}
		return 0, err
	}

	return role.Permissions, nil
}
