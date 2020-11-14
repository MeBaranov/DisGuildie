package helpers

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/utility"
)

type MessageProcessor interface {
	ProcessMessage(m message.Message)
}

type BaseMessageProcessor struct {
	Prov  database.DataProvider
	Funcs map[string]func(message.Message)
}

func (ap *BaseMessageProcessor) ProcessMessage(m message.Message) {
	cmd := m.PeekSegment()
	if utility.IsUserMention(cmd) {
		cmd = ""
	} else {
		cmd = m.CurSegment()
	}

	f, ok := ap.Funcs[cmd]
	if !ok {
		m.SendMessage("Unknown command \"%v\". Use \"!g help\" or \"!g h\" for help", m.FullMessage())
		return
	}

	f(m)
}

func (ap *BaseMessageProcessor) CheckGuildModificationPermissions(m message.Message, gid uuid.UUID) (bool, error) {
	var err error

	perm, err := m.AuthorPermissions()
	if err != nil {
		return false, err
	}

	if perm&database.EditGuildCharsPerm != 0 {
		return true, nil
	}

	auth, err := m.Author()
	if err != nil {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	gper, ok := auth.Guilds[m.GuildId()]
	if !ok {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	ok, err = utility.ValidateGuildAccess(ap.Prov, gper, gid)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (ap *BaseMessageProcessor) CheckUserModificationPermissions(m message.Message, uid string) (bool, error) {
	if uid == m.AuthorId() {
		return true, nil
	}

	var err error
	trgUser, err := ap.Prov.GetUserD(uid)
	if err != nil {
		return false, errors.New("User you're trying to modify doesn't seem to be a part of this guild")
	}

	trgPerm, ok := trgUser.Guilds[m.GuildId()]
	if !ok {
		return false, errors.New("User you're trying to modify doesn't seem to be a part of this guild")
	}

	perm, err := m.AuthorPermissions()
	if err != nil {
		return false, err
	}

	if perm&database.EditGuildCharsPerm != 0 {
		return true, nil
	}

	auth, err := m.Author()
	if err != nil {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	gper, ok := auth.Guilds[m.GuildId()]
	if !ok {
		return false, errors.New("You don't seem to be a part of this guild Oo. Try again later please")
	}

	ok, err = utility.ValidateUserAccess(ap.Prov, gper, trgPerm.GuildId)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (ap *BaseMessageProcessor) UserOrAuthorByMention(ment string, m message.Message) (*database.User, error) {
	if ment != "" {
		uid, err := utility.ParseUserMention(ment)
		if err != nil {
			return nil, err
		}

		u, err := ap.Prov.GetUserD(uid)
		if err != nil {
			return nil, err
		}
		return u, nil
	} else {
		u, err := m.Author()

		if err != nil {
			return nil, err
		}
		return u, nil
	}

}
