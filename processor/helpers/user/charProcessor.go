package user

import (
	"errors"
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
	"github.com/mebaranov/disguildie/utility"
)

type CharProcessor struct {
	helpers.BaseMessageProcessor
}

func NewCharProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &CharProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message) (string, error){
		"h":      ap.help,
		"help":   ap.help,
		"c":      ap.create,
		"create": ap.create,
		"main":   ap.main,
		"m":      ap.main,
	}
	return ap
}

func (ap *CharProcessor) create(m message.Message) (string, error) {
	ment, c := m.CurSegment(), m.CurSegment()

	if c == "" {
		c = ment
		ment = ""
	}

	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	gld, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to modify this user")
	}

	_, err = ap.Prov.GetCharacter(m.GuildId(), u.Id, c)
	if err == nil {
		return "", errors.New(fmt.Sprintf("User <@!%v> already have character %v", u.Id, c))
	}
	if err != nil {
		dbErr := database.ErrToDbErr(err)
		if dbErr == nil || dbErr.Code != database.CharacterNotFound {
			return "getting character", err
		}
	}

	stats := make(map[string]interface{})
	for _, s := range gld.Stats {
		switch s.Type {
		case database.Number:
			stats[s.ID] = 0
		case database.Str:
			stats[s.ID] = ""
		default:
			return "", errors.New("Undefined stat type for " + s.ID)
		}
	}

	ch := database.Character{
		UserId:      u.Id,
		GuildId:     m.GuildId(),
		Name:        c,
		Body:        stats,
		StatVersion: gld.StatVersion,
	}
	_, err = ap.Prov.AddCharacter(&ch)
	if err != nil {
		return "adding character", err
	}

	return fmt.Sprintf("Character %v added", c), nil
}

func (ap *CharProcessor) main(m message.Message) (string, error) {
	ment, c := m.CurSegment(), m.CurSegment()

	if c == "" {
		c = ment
		ment = ""
	}

	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to modify this user")
	}

	_, err = ap.Prov.GetCharacter(m.GuildId(), u.Id, c)
	if err != nil {
		return "getting character", err
	}

	_, err = ap.Prov.ChangeMainCharacter(m.GuildId(), u.Id, c)
	if err != nil {
		return "changing main character", err
	}

	return fmt.Sprintf("Character %v is set as main for <@!%v>", c, u.Id), nil
}

func (ap *CharProcessor) rename(m message.Message) (string, error) {
	ment, oldN, newN := m.CurSegment(), m.CurSegment(), m.CurSegment()
	if !utility.IsUserMention(ment) {
		if newN != "" {
			return "", errors.New("Invalid command format")
		}
		newN = oldN
		oldN = ment
		ment = ""
	}

	if newN == "" {
		newN = oldN
		oldN = ""
	}

	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to change this user")
	}

	c, err := ap.Prov.GetCharacter(m.GuildId(), u.Id, oldN)
	if err != nil {
		return "getting character", err
	}

	_, err = ap.Prov.GetCharacter(m.GuildId(), u.Id, newN)
	if err == nil {
		return "", errors.New(fmt.Sprintf("Character %v already extists", newN))
	}
	dbErr := database.ErrToDbErr(err)
	if dbErr == nil || dbErr.Code != database.CharacterNotFound {
		return "getting new character", err
	}

	_, err = ap.Prov.RenameCharacter(m.GuildId(), u.Id, c.Name, newN)
	if err != nil {
		return "renaming character", err
	}

	return fmt.Sprintf("Character %v renamed to %v", c.Name, newN), nil
}

func (ap *CharProcessor) give(m message.Message) (string, error) {
	oldOwner, char, newOwner := m.CurSegment(), m.CurSegment(), m.CurSegment()
	if !utility.IsUserMention(oldOwner) {
		if newOwner != "" {
			return "", errors.New("Invalid command format")
		}
		newOwner = char
		char = oldOwner
		oldOwner = ""
	}

	if !utility.IsUserMention(newOwner) {
		return "", errors.New("Invalid command format")
	}

	o, err := ap.UserOrAuthorByMention(oldOwner, m)
	if err != nil {
		return "getting source user", err
	}

	ok, err := ap.CheckUserModificationPermissions(m, o.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to change the owner")
	}

	n, err := ap.UserOrAuthorByMention(newOwner, m)
	if err != nil {
		return "getting target user", err
	}

	ok, err = ap.CheckUserModificationPermissions(m, n.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to change target user")
	}

	c, err := ap.Prov.GetCharacter(m.GuildId(), o.Id, char)
	if err != nil {
		return "getting character", err
	}

	_, err = ap.Prov.GetCharacter(m.GuildId(), n.Id, char)
	if err == nil {
		return "", errors.New(fmt.Sprintf("User <@!%v> already has character %v", n.Id, char))
	}
	dbErr := database.ErrToDbErr(err)
	if dbErr == nil || dbErr.Code != database.CharacterNotFound {
		return "getting new character", err
	}

	_, err = ap.Prov.ChangeCharacterOwner(m.GuildId(), o.Id, c.Name, n.Id)
	if err != nil {
		return "changing owner", err
	}

	return fmt.Sprintf("Character %v was given to <@!%v>", c.Name, n.Id), nil
}

func (ap *CharProcessor) remove(m message.Message) (string, error) {
	ment, char := m.CurSegment(), m.CurSegment()
	if !utility.IsUserMention(ment) {
		if char != "" {
			return "", errors.New("Invalid command format")
		}
		char = ment
		ment = ""
	}

	if char == "" {
		return "", errors.New("Invalid command format")
	}

	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to change this user")
	}

	c, err := ap.Prov.GetCharacter(m.GuildId(), u.Id, char)
	if err != nil {
		return "getting character", err
	}

	_, err = ap.Prov.RemoveCharacter(m.GuildId(), c.UserId, c.Name)
	if err != nil {
		return "removing character", err
	}

	return fmt.Sprintf("Character %v was removed", c.Name), nil
}

func (ap *CharProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of character commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	rv += "\t -- \"!g char create <char name>\" (\"!g c c <name>\") - Create your character\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char create <mention user> <char name>\" (\"!g c c <mention> <name>\") - Create a character for user\n"
	}

	rv += "\t -- \"!g char main <char name>\" (\"!g c m <name>\") - Set main character for yourself\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char main <mention user> <char name>\" (\"!g c m <mention> <name>\") - Set main character for a user\n"
	}

	rv += "\t -- \"!g char rename <new name>\" (\"!g c n <new name>\") - Change name of your main character\n"
	rv += "\t -- \"!g char rename <old name> <new name>\" (\"!g c n <old name> <new name>\") - Change name of your character\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char rename <mention user> <new name>\" (\"!g c n <mention> <new name>\") - Change name of users main character\n"
		rv += "\t -- \"!g char rename <mention user> <old name> <new name>\" (\"!g c n <mention> <old name> <new name>\") - Change name of users character\n"
	}

	rv += "\t -- \"!g char give <char name> <mention user>\" (\"!g c g <char name> <mention>\") - Give your char to the user\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char give <mention owner> <char name> <mention user>\" (\"!g c g <mention> <char name> <mention>\") - Give users char to the other user\n"
	}

	rv += "\t -- \"!g char remove <char name>\" (\"!g c remove <name>\") - Remove your character\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char remove <mention user> <char name>\" (\"!g c remove <mention> <name>\") - Remove user's character\n"
	}
	rv += "Be aware that your ability to modify other members characters depends on your subguilds.\n"

	return rv, nil
}
