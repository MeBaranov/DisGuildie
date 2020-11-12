package helpers

import (
	"errors"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/utility"
)

type AdminRoleProcessor struct {
	BaseMessageProcessor
}

func NewAdminRoleProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminRoleProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message){
		"h":    ap.help,
		"help": ap.help,
		"a":    ap.add,
		"add":  ap.add,
	}
	return ap
}

func (ap *AdminRoleProcessor) add(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		m.SendMessage("I'm sorry, but you don't have permissions to run guild-wide structure management operations")
		return
	}

	roleStr, rid, err := ap.getRoleId(m)
	if err != nil {
		m.SendMessage("Could not parse role: %v", err.Error())
		return
	}

	permStr, p, err := ap.getPermission(m)
	if err != nil {
		m.SendMessage("Could not parse permission: %v", err.Error())
		return
	}

	role, dbErr := ap.Prov.GetRole(m.GuildId(), rid)
	if dbErr != nil {
		if dbErr.Code != database.RoleNotFound {
			m.SendMessage("Error getting role: %v", dbErr.Error())
			return
		}

		role = &database.Role{
			GuildId:     m.GuildId(),
			Id:          rid,
			Permissions: p,
		}
		if _, dbErr = ap.Prov.AddRole(role); dbErr != nil {
			m.SendMessage("Could not add role: %v", dbErr.Error())
			return
		}
		m.SendMessage("Permission %v added for the role %v", permStr, roleStr)
		return
	}

	p = p | role.Permissions
	if p != role.Permissions {
		if _, err = ap.Prov.SetRolePermissions(m.GuildId(), rid, p); err != nil {
			m.SendMessage("Could not add permission: %v", err.Error())
			return
		}
	}

	usrs, err := m.GuildMembersWithRole(rid)
	if err != nil {
		m.SendMessage("Role permissions were updated. But got error getting users: %v", err.Error())
		return
	}

	for uid, _ := range usrs {
		u, err := ap.Prov.GetUserD(uid)
		if err != nil {
			m.SendMessage("Error getting a user for update: %v", err.Error())
			return
		}
		if uper, ok := u.Guilds[m.GuildId()]; ok {
			uper.Permissions |= p
			_, err = ap.Prov.SetUserPermissions(uid, uper)
			if err != nil {
				m.SendMessage("Error while updating users: %v", err.Error())
				return
			}
		}
	}

	m.SendMessage("Permission %v added for the role %v", permStr, roleStr)
}

func (ap *AdminRoleProcessor) remove(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		m.SendMessage("I'm sorry, but you don't have permissions to run guild-wide structure management operations")
		return
	}

	roleStr, rid, err := ap.getRoleId(m)
	if err != nil {
		m.SendMessage("Could not parse role: %v", err.Error())
		return
	}

	permStr, p, err := ap.getPermission(m)
	if err != nil {
		m.SendMessage("Could not parse permission: %v", err.Error())
		return
	}

	role, dbErr := ap.Prov.GetRole(m.GuildId(), rid)
	if dbErr != nil {
		m.SendMessage("Error getting role: %v", dbErr.Error())
		return
	}

	p = ^p & role.Permissions
	if p != role.Permissions {
		if _, err = ap.Prov.SetRolePermissions(m.GuildId(), rid, p); err != nil {
			m.SendMessage("Could not add permission: %v", err.Error())
			return
		}
	}

	m.SendMessage("Permission %v removed from the role %v", permStr, roleStr)
}

func (ap *AdminRoleProcessor) reset(m message.Message) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		m.SendMessage("Some error happened while getting permissions: %v", err.Error())
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		m.SendMessage("I'm sorry, but you don't have permissions to run guild-wide structure management operations")
		return
	}

	roleStr, rid, err := ap.getRoleId(m)
	if err != nil {
		m.SendMessage("Could not parse role: %v", err.Error())
		return
	}

	if _, err = ap.Prov.RemoveRole(m.GuildId(), rid); err != nil {
		m.SendMessage("Error removing role: %v", err.Error())
		return
	}

	m.SendMessage("Permissions for the role %v were reset", roleStr)
}

func (ap *AdminRoleProcessor) help(m message.Message) {
	rv := "Here's a list of guild management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		m.SendMessage(rv)
		return
	}

	if perm&database.EditGuildStructurePerm == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		m.SendMessage(rv)
		return
	}

	rv += "\t -- \"!g admin role add <role> <permission>\" (\"!g a r a <role> <permission>\") - Add permission to a role\n"
	rv += "\t -- \"!g admin role remove <role> <permission>\" (\"!g a r r <role> <permission>\") - Remove permission from a role\n"
	rv += "\t -- \"!g admin role reset <role>\" (\"!g a r reset <role>\") - Remove permissions for a role\n"
	rv += "\n"
	rv += "In the explanation above <role> can be role name or role mention\n"
	rv += "<permission> is one of the following:\n"
	rv += "-- \"SubEditUser\" (\"su\") - lets role members edit users and characters in their subguild (and all guilds under it)"
	rv += "-- \"SubEditGuild\" (\"sg\") - lets role members edit structure of their subguild (and all guilds under it)"
	rv += "-- \"OneUpEditUser\" (\"uu\") - lets role members edit users and characters of a subguild above their (and all under)"
	rv += "-- \"OneUpEditGuild\" (\"ug\") - lets role members edit structure of a subguild above their (and all under)"
	rv += "-- \"GuildEditUser\" (\"gu\") - lets role members edit users and characters of the entire guild"
	rv += "-- \"GuildEditGuild\" (\"gg\") - lets role members edit structure of the entire guild"
	rv += "Notice that last two permissions grant group-wide operations access. Like this one."

	m.SendMessage(rv)
}

func (ap *AdminRoleProcessor) getRoleId(m message.Message) (string, string, error) {
	roleStr := m.CurSegment()
	if roleStr == "" {
		return roleStr, "", errors.New("Malformed command. Role is not present")
	}

	rid, err := utility.ParseRoleMention(roleStr)
	if err != nil {
		rid, err = m.GetRoleId(roleStr)
		if err != nil {
			return roleStr, "", err
		}
	}

	return roleStr, rid, nil
}

func (ap *AdminRoleProcessor) getPermission(m message.Message) (string, int, error) {
	permStr := m.CurSegment()

	if permStr == "" {
		return permStr, 0, errors.New("Malformed command. Permission is not present")
	}

	p, err := database.StringToPermission(permStr)
	if err != nil {
		return permStr, p, err
	}

	return permStr, p, nil
}
