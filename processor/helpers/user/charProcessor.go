package user

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

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
		"list":   ap.list,
		"l":      ap.list,
		"c":      ap.create,
		"create": ap.create,
		"main":   ap.main,
		"m":      ap.main,
	}
	return ap
}

func (ap *CharProcessor) list(m message.Message) (string, error) {
	u, err := ap.UserOrAuthorByMention(m.CurSegment(), m)
	if err != nil {
		return "getting target user", err
	}

	chars, err := ap.Prov.GetCharacters(m.GuildId(), u.Id)
	if err != nil {
		return "getting characters", err
	}

	rv := "List of characters:\n"
	for _, c := range chars {
		rv += "\t"
		if c.Main {
			rv += "[Main] "
		}
		rv += c.Name + "\n"
	}

	return rv, nil
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

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		return "checking modification permissions", err
	}
	if !ok {
		return "", errors.New("You don't have permissions to modify this user")
	}

	_, dbErr := ap.Prov.GetCharacter(m.GuildId(), u.Id, c)
	if dbErr == nil {
		return "", errors.New(fmt.Sprintf("User <@!%v> already have character %v", u.Id, c))
	}
	if dbErr != nil && dbErr.Code != database.CharacterNotFound {
		return "getting character", dbErr
	}

	ch := database.Character{
		UserId:  u.Id,
		GuildId: m.GuildId(),
		Name:    c,
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

func (ap *CharProcessor) stat(m message.Message) (string, error) {
	v1, v2, v3, v4 := m.CurSegment(), m.CurSegment(), m.CurSegment(), m.CurSegment()
	if v3 == "" && v4 == "" {
		if v1 != "" && utility.IsUserMention(v1) {
			return ap.getStat(m, v1, v2)
		}
		if v2 == "" {
			return ap.getStat(m, "", v1)
		}

		return ap.setStat(m, "", "", v1, v2)
	}

	if v4 != "" {
		return ap.setStat(m, v1, v2, v3, v4)
	}

	if utility.IsUserMention(v1) {
		return ap.setStat(m, v1, "", v2, v3)
	}

	return ap.setStat(m, "", v1, v2, v3)
}

func (ap *CharProcessor) getStat(m message.Message, ment string, char string) (string, error) {
	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		return "getting target user", err
	}

	c, err := ap.Prov.GetCharacter(m.GuildId(), u.Id, char)
	if err != nil {
		return "getting character", err
	}

	gld, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	rvs := make([]string, 0, len(c.Body))
	for n, v := range c.Body {
		s, ok := gld.Stats[n]
		if !ok {
			ap.Prov.RemoveCharacterStat(c.GuildId, c.UserId, c.Name, n)
			continue
		}
		switch s.Type {
		case database.Number:
			if _, err := v.(int); err {
				ap.Prov.RemoveCharacterStat(c.GuildId, c.UserId, c.Name, n)
				continue
			}
		case database.Str:
			if _, err := v.(string); err {
				ap.Prov.RemoveCharacterStat(c.GuildId, c.UserId, c.Name, n)
				continue
			}
		}

		rvs = append(rvs, fmt.Sprintf("\t%v:%v\n", n, v))
	}

	sort.Strings(rvs)
	return fmt.Sprintf("Stats are:\n\tmain:%v\n\tname:%v\n%v", c.Main, c.Name, rvs), nil
}

func (ap *CharProcessor) setStat(m message.Message, ment string, char string, stat string, value string) (string, error) {
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

	gld, err := ap.Prov.GetGuildD(m.GuildId())
	if err != nil {
		return "getting guild", err
	}

	s, ok := gld.Stats[stat]
	if !ok {
		return "", errors.New(fmt.Sprintf("Stat %v is not defined in your guild", stat))
	}

	var val interface{}
	switch s.Type {
	case database.Number:
		val, err = strconv.Atoi(value)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Expected numeric value. Got %v", value))
		}
	case database.Str:
		val = value
	}

	_, err = ap.Prov.SetCharacterStat(c.GuildId, c.UserId, c.Name, stat, val)
	if err != nil {
		return "setting character stat", err
	}

	return fmt.Sprintf("Stat %v set to %v for character %v", stat, value, c.Name), nil
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

	_, dbErr := ap.Prov.GetCharacter(m.GuildId(), u.Id, newN)
	if dbErr == nil {
		return "", errors.New(fmt.Sprintf("Character %v already extists", newN))
	}
	if dbErr.Code != database.CharacterNotFound {
		return "getting new character", dbErr
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

	_, dbErr := ap.Prov.GetCharacter(m.GuildId(), n.Id, char)
	if dbErr == nil {
		return "", errors.New(fmt.Sprintf("User <@!%v> already has character %v", n.Id, char))
	}
	if dbErr.Code != database.CharacterNotFound {
		return "getting new character", dbErr
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

	rv += "\t -- \"!g chars\" (\"!g c\") - List your characters\n"
	rv += "\t -- \"!g chars <mention user>\" (\"!g c <mention>\") - List users characters\n"

	rv += "\t -- \"!g char create <char name>\" (\"!g c c <name>\") - Create your character\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char create <mention user> <char name>\" (\"!g c c <mention> <name>\") - Create a character for user\n"
	}

	rv += "\t -- \"!g char main <char name>\" (\"!g c m <name>\") - Set main character for yourself\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char main <mention user> <char name>\" (\"!g c m <mention> <name>\") - Set main character for a user\n"
	}

	rv += "\t -- \"!g char stat\" (\"!g c s\") - Get stats for your main character\n"
	rv += "\t -- \"!g char stat <char name>\" (\"!g c s <name>\") - Get stats for your character\n"
	rv += "\t -- \"!g char stat <mention user>\" (\"!g c s <mention>\") - Get stats for users main character\n"
	rv += "\t -- \"!g char stat <mention user> <char name>\" (\"!g c s <mention> <name>\") - Get stats for users character\n"

	rv += "\t -- \"!g char stat <stat name> <stat value>\" (\"!g c s <stat name> <stat value>\") - Set stat for your main character\n"
	rv += "\t -- \"!g char stat <char name> <stat name> <stat value>\" (\"!g c s <char name> <stat name> <stat value>\") - Set stat for your character\n"
	if perm&database.CharsPermissions != 0 {
		rv += "\t -- \"!g char stat <mention user> <stat name> <stat value>\" (\"!g c s <mention user> <stat name> <stat value>\") - Set stat for other users main character\n"
		rv += "\t -- \"!g char stat <mention user> <char name> <stat name> <stat value>\" (\"!g c s <mention user> <char name> <stat name> <stat value>\") - Set stat for other users character\n"
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
