package helpers

import (
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
		go m.SendMessage("Unknown user administration command \"%v\". Send \"!g admin user help\" or \"!g a u h\" for help", m.FullMessage())
		return
	}

	f(m)
}

func (ap *AdminUserProcessor) register(m message.Message) {
	ap.regOrSync(m, register)
}

func (ap *AdminUserProcessor) sync(m message.Message) {
	ap.regOrSync(m, sync)
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
	rv += "\t -- \"!g admin user remove <Discord ID>\" (\"!g a u remove <Discord ID>\") - Remove user from the system by discord ID\n"
	rv += "\t -- \"!g admin user assign <mention user> <sub-guild name>\" (\"!g a u a <user> <name>\") - Move user to a sub-guild\n"
	rv += "\t -- \"!g admin user sync <mention user>\" (\"!g a u p <user>\") - Synchronize user permissions\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user sync all\" (\"!g a u s all\") - Synchronize all users permissions\n"
	}

	go m.SendMessage(rv)
}

const (
	register = iota
	sync
)

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
		if err != nil {
			go m.SendMessage("Could not register user: %v", err.Error())
			return
		}
	case sync:
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

func (ap *AdminUserProcessor) reigsterUser(id string, guild *database.Guild, m message.Message) error {
	// TODO: No way. Register == add guild, not just check that user is there
	dbu, dbErr := ap.prov.GetUserD(id)
	if dbErr == nil {
		return nil
	} else if dbErr.Code != database.UserNotFound {
		return dbErr
	}

	dbu = &database.User{
		DiscordId: id,
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

func (ap *AdminUserProcessor) syncUser(id string, guild *database.Guild, m message.Message) error {
	// TODO: Logic

	return nil
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
