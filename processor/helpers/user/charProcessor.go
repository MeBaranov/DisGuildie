package user

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type CharProcessor struct {
	helpers.BaseMessageProcessor
}

func NewCharProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &CharProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message){
		"h":    ap.help,
		"help": ap.help,
	}
	return ap
}

func (ap *CharProcessor) list(m message.Message) {
	u, err := ap.UserOrAuthorByMention(m.CurSegment(), m)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}

	chars, err := ap.Prov.GetCharacters(m.GuildId(), u.Id)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}

	rv := "List of characters:\n"
	for _, c := range chars {
		rv += "\t"
		if c.Main {
			rv += "[Main] "
		}
		rv += c.Name + "\n"
	}

	m.SendMessage(rv)
}

func (ap *CharProcessor) create(m message.Message) {
	ment, c := m.CurSegment(), m.CurSegment()

	if c == "" {
		c = ment
		ment = ""
	}

	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}
	if !ok {
		m.SendMessage("You don't have permissions to modify this user")
		return
	}

	_, dbErr := ap.Prov.GetCharacter(m.GuildId(), u.Id, c)
	if dbErr == nil {
		m.SendMessage("User <@!%v> already have character %v", u.Id, c)
		return
	}
	if dbErr != nil && dbErr.Code != database.CharacterNotFound {
		m.SendMessage(dbErr.Error())
		return
	}

	ch := database.Character{
		UserId:  u.Id,
		GuildId: m.GuildId(),
		Name:    c,
	}
	_, err = ap.Prov.AddCharacter(&ch)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}

	m.SendMessage("Character %v added", c)
}

func (ap *CharProcessor) main(m message.Message) {
	ment, c := m.CurSegment(), m.CurSegment()

	if c == "" {
		c = ment
		ment = ""
	}

	u, err := ap.UserOrAuthorByMention(ment, m)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}

	ok, err := ap.CheckUserModificationPermissions(m, u.Id)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}
	if !ok {
		m.SendMessage("You don't have permissions to modify this user")
		return
	}

	_, dbErr := ap.Prov.GetCharacter(m.GuildId(), u.Id, c)
	if dbErr != nil {
		m.SendMessage(dbErr.Error())
		return
	}

	_, err = ap.Prov.ChangeMainCharacter(m.GuildId(), u.Id, c)
	if err != nil {
		m.SendMessage(err.Error())
		return
	}

	m.SendMessage("Character %v is set as main for <@!%v>", c, u.Id)
}

func (ap *CharProcessor) help(m message.Message) {
	rv := "Here's a list of character commands you're allowed to use:\n"

	perm, err := m.AuthorPermissions()
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		m.SendMessage(rv)
		return
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

	m.SendMessage(rv)
}
