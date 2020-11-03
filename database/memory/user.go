package memory

import (
	"sync"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
)

type UserMemoryDb struct {
	usersD map[string]*database.User
	mux    sync.Mutex
}

func (udb *UserMemoryDb) AddUser(u *database.User, gp *database.GuildPermission) (*database.User, error) {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	if user, ok := udb.usersD[u.DiscordId]; ok {
		if _, ok = user.Guilds[gp.TopGuild]; !ok {
			user.Guilds[gp.TopGuild] = gp
			return user, nil
		}

		return nil, &database.Error{Code: database.UserAlreadyInGuild, Message: "The user is already registered in the guild"}
	}

	uid := uuid.New()
	u.UserId = uid
	u.Guilds = map[string]*database.GuildPermission{gp.TopGuild: gp}
	udb.usersD[u.DiscordId] = u

	return u, nil
}

func (udb *UserMemoryDb) GetUserD(d string) (*database.User, error) {
	if user, ok := udb.usersD[d]; ok {
		return user, nil
	}

	return nil, &database.Error{Code: database.UserNotFound, Message: "User was not found"}
}

func (udb *UserMemoryDb) GetUsersInGuild(d string) ([]*database.User, error) {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	rv := make([]*database.User, 0, 100)
	for _, u := range udb.usersD {
		if _, ok := u.Guilds[d]; ok {
			rv = append(rv, u)
		}
	}

	return rv, nil
}

func (udb *UserMemoryDb) SetUserPermissions(u string, gp *database.GuildPermission) (*database.User, error) {
	user, err := udb.GetUserD(u)
	if err != nil {
		return nil, err
	}

	curGp, ok := user.Guilds[gp.TopGuild]
	if !ok {
		return nil, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in the guild"}
	}

	curGp.Permissions = gp.Permissions
	return user, nil
}

func (udb *UserMemoryDb) SetUserSubGuild(u string, gp *database.GuildPermission) (*database.User, error) {
	user, err := udb.GetUserD(u)
	if err != nil {
		return nil, err
	}

	curGp, ok := user.Guilds[gp.TopGuild]
	if !ok {
		return nil, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in the guild"}
	}

	curGp.GuildId = gp.GuildId
	return user, nil
}

func (udb *UserMemoryDb) RemoveUserD(u string, g string) (*database.User, error) {
	user, err := udb.GetUserD(u)
	if err != nil {
		return nil, err
	}

	if _, ok := user.Guilds[g]; !ok {
		return nil, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in the guild"}
	}

	delete(user.Guilds, g)

	return user, nil
}

func (udb *UserMemoryDb) EraseUserD(u string) (*database.User, error) {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	user, err := udb.GetUserD(u)
	if err != nil {
		return nil, err
	}

	delete(udb.usersD, u)
	return user, nil
}
