package admin

import (
	"errors"
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
	"github.com/mebaranov/disguildie/utility"
)

type AdminRoleProcessor struct {
	helpers.BaseMessageProcessor
}

func NewAdminRoleProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &AdminRoleProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message) (string, error){
		"h":    ap.help,
		"help": ap.help,
		"a":    ap.add,
		"add":  ap.add,
	}
	return ap
}

func (ap *AdminRoleProcessor) add(m message.Message) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		return "", errors.New("You don't have permissions to run guild-wide structure management operations")
	}

	roleStr, rid, err := ap.getRoleId(m)
	if err != nil {
		return "parsing role", err
	}

	permStr, p, err := ap.getPermission(m)
	if err != nil {
		return "parsing permission", err
	}

	role, err := ap.Prov.GetRole(m.GuildId(), rid)
	if err != nil {
		dbErr := database.ErrToDbErr(err)
		if dbErr == nil || dbErr.Code != database.RoleNotFound {
			return "getting role", err
		}

		role = &database.Role{
			GuildId:     m.GuildId(),
			Id:          rid,
			Permissions: p,
		}
		if _, err = ap.Prov.AddRole(role); err != nil {
			return "adding role", err
		}
		return fmt.Sprintf("Permission %v added for the role %v", permStr, roleStr), nil
	}

	p = p | role.Permissions
	if p != role.Permissions {
		if _, err = ap.Prov.SetRolePermissions(m.GuildId(), rid, p); err != nil {
			return "setting role permissions", err
		}
	}

	usrs, err := m.GuildMembersWithRole(rid)
	if err != nil {
		return "getting users", err
	}

	for uid, _ := range usrs {
		u, err := ap.Prov.GetUserD(uid)
		if err != nil {
			return "getting a user for update", err
		}
		if uper, ok := u.Guilds[m.GuildId()]; ok {
			uper.Permissions |= p
			_, err = ap.Prov.SetUserPermissions(uid, uper)
			if err != nil {
				return "updating user", err
			}
		}
	}

	return fmt.Sprintf("Permission %v added for the role %v", permStr, roleStr), nil
}

func (ap *AdminRoleProcessor) remove(m message.Message) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		return "", errors.New("You don't have permissions to run guild-wide structure management operations")
	}

	roleStr, rid, err := ap.getRoleId(m)
	if err != nil {
		return "parsing role", err
	}

	permStr, p, err := ap.getPermission(m)
	if err != nil {
		return "parsing permission", err
	}

	role, err := ap.Prov.GetRole(m.GuildId(), rid)
	if err != nil {
		return "getting role", err
	}

	p = ^p & role.Permissions
	if p != role.Permissions {
		if _, err = ap.Prov.SetRolePermissions(m.GuildId(), rid, p); err != nil {
			return "adding permission", err
		}
	}

	return fmt.Sprintf("Permission %v removed from the role %v", permStr, roleStr), nil
}

func (ap *AdminRoleProcessor) reset(m message.Message) (string, error) {
	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		return "", errors.New("You don't have permissions to run guild-wide structure management operations")
	}

	roleStr, rid, err := ap.getRoleId(m)
	if err != nil {
		return "parsing role", err
	}

	if _, err = ap.Prov.RemoveRole(m.GuildId(), rid); err != nil {
		return "removing role", err
	}

	return fmt.Sprintf("Permissions for the role %v were reset", roleStr), nil
}

func (ap *AdminRoleProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of role management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting permissions", err
	}

	if perm&database.EditGuildStructurePerm == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		return rv, nil
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

	return rv, nil
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
