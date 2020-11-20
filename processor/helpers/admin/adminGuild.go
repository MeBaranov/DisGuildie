package admin

import (
	"errors"
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type AdminGuildProcessor struct {
	helpers.BaseMessageProcessor
}

func NewAdminGuildProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &AdminGuildProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message) (string, error){
		"h":      ap.help,
		"help":   ap.help,
		"a":      ap.add,
		"add":    ap.add,
		"r":      ap.rename,
		"rename": ap.rename,
		"m":      ap.move,
		"move":   ap.move,
		"remove": ap.remove,
	}
	return ap
}

func (ap *AdminGuildProcessor) add(m message.Message) (string, error) {
	gldName := m.CurSegment()
	parent := m.CurSegment()

	if parent == "" || gldName == "" {
		return "", errors.New("Invalid command format")
	}

	var pguild *database.Guild
	var err error
	if parent == "main" {
		pguild, err = ap.Prov.GetGuildD(m.GuildId())
	} else {
		pguild, err = ap.Prov.GetGuildN(m.GuildId(), parent)
	}
	if err != nil {
		return "getting parent guild", err
	}

	ok, err := m.CheckGuildModificationPermissions(pguild.GuildId)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to modify the sub-guild")
	}

	g := &database.Guild{
		Name:     gldName,
		ParentId: pguild.GuildId,
	}
	if _, err = ap.Prov.AddGuild(g); err != nil {
		return "adding guild", err
	}

	return fmt.Sprintf("Sub-guild %v registered under %v.", gldName, parent), nil
}

func (ap *AdminGuildProcessor) rename(m message.Message) (string, error) {
	oldName := m.CurSegment()
	newName := m.CurSegment()
	if oldName == "" || newName == "" {
		return "", errors.New("Invalid command format")
	}

	var err error
	g, err := ap.Prov.GetGuildN(m.GuildId(), oldName)
	if err != nil {
		return "getting source guild", err
	}

	ok, err := m.CheckGuildModificationPermissions(g.GuildId)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to modify the sub-guild")
	}

	if _, err = ap.Prov.RenameGuild(g.GuildId, newName); err != nil {
		return "renaming guild", err
	}

	return fmt.Sprintf("Sub-guild '%v' renamed to '%v'", oldName, newName), nil
}

func (ap *AdminGuildProcessor) move(m message.Message) (string, error) {
	name := m.CurSegment()
	parent := m.CurSegment()
	if name == "" || parent == "" {
		return "", errors.New("Invalid command format")
	}

	var err error
	g, err := ap.Prov.GetGuildN(m.GuildId(), name)
	if err != nil {
		return "getting source guild", err
	}

	var pguild *database.Guild
	if parent == "main" {
		pguild, err = ap.Prov.GetGuildD(m.GuildId())
	} else {
		pguild, err = ap.Prov.GetGuildN(m.GuildId(), parent)
	}
	if err != nil {
		return "getting target guild", err
	}

	ok, err := m.CheckGuildModificationPermissions(g.GuildId)
	if err != nil {
		return "checking source modification permissions", err
	}
	if !ok {
		return "", errors.New(fmt.Sprintf("You don't have permissions to modify the source (%v) sub-guild", name))
	}

	ok, err = m.CheckGuildModificationPermissions(pguild.GuildId)
	if err != nil {
		return "checking target modification pemissions", err
	}
	if !ok {
		return "", errors.New(fmt.Sprintf("You don't have permissions to modify the target (%v) sub-guild", name))
	}

	if _, err = ap.Prov.MoveGuild(g.GuildId, pguild.GuildId); err != nil {
		return "moving guild", err
	}

	return fmt.Sprintf("Sub-guild '%v' moved under '%v'", name, parent), nil
}

func (ap *AdminGuildProcessor) remove(m message.Message) (string, error) {
	name := m.CurSegment()
	if name == "" {
		return "", errors.New("Invalid command format")
	}

	var err error
	g, err := ap.Prov.GetGuildN(m.GuildId(), name)
	if err != nil {
		return "getting guild", err
	}

	ok, err := m.CheckGuildModificationPermissions(g.GuildId)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to modify the sub-guild")
	}

	subs, err := ap.Prov.GetSubGuilds(g.GuildId)
	if err != nil {
		return "getting sub-guilds", err
	}

	users, err := ap.Prov.GetUsersInGuild(m.GuildId())
	if err != nil {
		return "getting users in guild", err
	}

	for _, s := range users {
		if perm, ok := s.Guilds[m.GuildId()]; ok {
			if _, ok = subs[perm.GuildId]; ok {
				_, err = ap.Prov.SetUserSubGuild(s.Id, &database.GuildPermission{TopGuild: m.GuildId(), GuildId: g.ParentId})
				if err != nil {
					return "moving users out from sub-guild", err
				}
			}
		}
	}

	if _, err = ap.Prov.RemoveGuild(g.GuildId); err != nil {
		return "removing sub-guild", err
	}

	return fmt.Sprintf("Sub-guild '%v' removed", name), nil
}

func (ap *AdminGuildProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of guild management commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting permissions", err
	}

	if perm&database.StructurePermissions == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		return rv, nil
	}

	rv += "\t -- \"!g admin guild add <child guild name> main\" (\"!g a g a <name>\") - Add sub-guild to the main level\n"
	rv += "\t -- \"!g admin guild add <child guild name> <parent guild name>\" (\"!g a g a <child> <parent>\") - Add sub-guild to a parent sub-guild\n"
	rv += "\t -- \"!g admin guild rename <old sub-guild name> <new name>\" (\"!g a g r <old> <new>\") - Rename sub-guild\n"
	rv += "\t -- \"!g admin guild move <child guild name> main\" (\"!g a g m <name> main\") - Move sub-guild to a the main level\n"
	rv += "\t -- \"!g admin guild move <child guild name> <new parent guild>\" (\"!g a g m <name> <new parent>\") - Move sub-guild to a new parent\n"
	rv += "\t -- \"!g admin guild remove <child guild name>\" (\"!g a g remove <name>\") - Remove sub-guild\n"
	rv += "Be aware that your ability to modify structure depends on the guild you're assigned to.\n"

	return rv, nil
}
