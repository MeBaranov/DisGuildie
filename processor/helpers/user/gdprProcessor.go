package user

import (
	"errors"
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type GdprProcessor struct {
	helpers.BaseMessageProcessor
}

func NewGdprProcessor(prov database.DataProvider) helpers.MessageProcessor {
	ap := &GdprProcessor{}
	ap.Prov = prov
	ap.Funcs = map[string]func(message.Message) (string, error){
		"h":      ap.help,
		"help":   ap.help,
		"l":      ap.list,
		"list":   ap.list,
		"remove": ap.remove,
	}
	return ap
}

func (ap *GdprProcessor) list(m message.Message) (string, error) {
	a, err := m.Author()
	if err != nil {
		return "getting author", err
	}

	rv := "Your guilds and characters:\n"
	for _, g := range a.Guilds {
		tgld, err := ap.Prov.GetGuildD(g.TopGuild)
		if err != nil {
			return "getting top level guild", err
		}
		sgld, err := ap.Prov.GetGuild(g.GuildId)
		if err != nil {
			return "getting sub-guild", err
		}
		mon, err := ap.Prov.GetMoney(g.TopGuild)
		if err != nil {
			return "getting payments", err
		}
		rv += fmt.Sprintf("-Guild: %v (ID: %v), Sub-Guild: %v", tgld.Name, tgld.DiscordId, sgld.Name)
		if mon.UserId == a.Id {
			rv += " [You payed for it]"
		}
		rv += "\n"
		chars, err := ap.Prov.GetCharacters(tgld.DiscordId, a.Id)
		if err != nil {
			return "getting characters", err
		}
		for _, c := range chars {
			rv += "--" + c.Name + "\n"
		}
	}

	return rv, nil
}

func (ap *GdprProcessor) remove(m message.Message) (string, error) {
	me, id := m.CurSegment(), m.CurSegment()
	if me != "me" {
		return "", errors.New("Invalid command format. It has very specific syntax. Consult \"!g g h\"")
	}

	a, err := m.Author()
	if err != nil {
		return "getting author", err
	}

	if id == "" {
		return fmt.Sprintf("Please, use the following command: \"!g g remove me %v\" to approve deletion." + m.GuildId()), nil
	}

	gld, err := ap.Prov.GetGuildD(id)
	if err != nil {
		return "getting guild", err
	}

	money, err := ap.Prov.GetMoney(gld.DiscordId)
	if err != nil {
		return "getting payments", err
	}

	chars, err := ap.Prov.GetCharacters(gld.DiscordId, a.Id)
	if err != nil {
		return "getting characters", err
	}

	if money.UserId == a.Id {
		_, err := ap.Prov.ChangeMoneyOwner(gld.DiscordId, "")
		if err != nil {
			return "changing payment owner", nil
		}
	}

	for _, c := range chars {
		_, err = ap.Prov.RemoveCharacter(c.GuildId, c.UserId, c.Name)
		if err != nil {
			return "removing character", err
		}
	}

	_, err = ap.Prov.RemoveUserD(a.Id, gld.DiscordId)
	if err != nil {
		return "removing user", err
	}

	return "You were removed from guild with ID " + id, nil
}

func (ap *GdprProcessor) forget(m message.Message) (string, error) {
	me, id := m.CurSegment(), m.CurSegment()
	if me != "me" {
		return "", errors.New("Invalid command format. It has very specific syntax. Consult \"!g g h\"")
	}

	a, err := m.Author()
	if err != nil {
		return "getting author", err
	}

	if id == "" {
		return fmt.Sprintf("Please, use the following command: \"!g g forget me %v\" to approve deletion." + a.Id), nil
	}

	for _, g := range a.Guilds {
		gld, err := ap.Prov.GetGuildD(g.TopGuild)
		if err != nil {
			return "getting guild", err
		}

		money, err := ap.Prov.GetMoney(gld.DiscordId)
		if err != nil {
			return "getting payments", err
		}

		chars, err := ap.Prov.GetCharacters(gld.DiscordId, a.Id)
		if err != nil {
			return "getting characters", err
		}

		if money.UserId == a.Id {
			_, err := ap.Prov.ChangeMoneyOwner(gld.DiscordId, "")
			if err != nil {
				return "changing payment owner", nil
			}
		}

		for _, c := range chars {
			_, err = ap.Prov.RemoveCharacter(c.GuildId, c.UserId, c.Name)
			if err != nil {
				return "removing character", err
			}
		}

		_, err = ap.Prov.RemoveUserD(a.Id, gld.DiscordId)
		if err != nil {
			return "removing user", err
		}
	}

	return "You were totally removed from the system. You're always welcome to come back.", nil
}

func (ap *GdprProcessor) help(m message.Message) (string, error) {
	rv := "Here's a list of gdpr commands you're allowed to use:\n"
	rv += "\t -- \"!g gdpr list\" (\"!g g l\") - List which guilds you belong to and your characters there\n"
	rv += "\t -- \"!g gdpr remove me\" (\"!g g remove me\") - Remove yourself and your characters from this guild\n"
	rv += "\t -- \"!g gdpr remove me <guild id>\" (\"!g g remove me <guild id>\") - Remove yourself and your characters from guild by id. See \"!g g l\" for guild ids\n"
	rv += "\t -- \"!g gdpr forget me\" (\"!g g forget me\") - Remove yourself and your characters from all guilds\n"
	rv += "\t -- \"!g gdpr forget me <ID>\" (\"!g g forget me <ID>\") - Submit removal of yourself and your characters from all guilds. For ID use \"!g g forget me\"\n"

	rv += "\nInformation that I store: your unique discord ID, your guild memberships, your ownership if you were the last one who payed for a guild"

	return rv, nil
}
