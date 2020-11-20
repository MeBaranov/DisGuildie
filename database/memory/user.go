package memory

import (
	"sync"

	"github.com/mebaranov/disguildie/database"
)

type UserMemoryDb struct {
	UsersD map[string]*database.User
	mux    sync.Mutex
}

func (udb *UserMemoryDb) AddUser(d string, gp *database.GuildPermission) (*database.User, error) {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	if user, ok := udb.UsersD[d]; ok {
		if _, ok = user.Guilds[gp.TopGuild]; !ok {
			user.Guilds[gp.TopGuild] = gp
			tmp := *user
			return &tmp, nil
		}

		return nil, &database.Error{Code: database.UserAlreadyInGuild, Message: "The user is already registered in the guild"}
	}

	tmpGp := *gp
	u := &database.User{
		Id:     d,
		Guilds: map[string]*database.GuildPermission{gp.TopGuild: &tmpGp},
	}
	udb.UsersD[d] = u

	tmp := *u
	return &tmp, nil
}

func (udb *UserMemoryDb) GetUserD(d string) (*database.User, error) {
	rv, err := udb.getUserD(d)
	if err != nil {
		return nil, err
	}

	tmp := *rv
	return &tmp, nil
}

func (udb *UserMemoryDb) GetUsersInGuild(d string) ([]*database.User, error) {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	rv := make([]*database.User, 0, 100)
	for _, u := range udb.UsersD {
		if _, ok := u.Guilds[d]; ok {
			tmp := *u
			rv = append(rv, &tmp)
		}
	}

	return rv, nil
}

func (udb *UserMemoryDb) SetUserPermissions(u string, gp *database.GuildPermission) (*database.User, error) {
	user, err := udb.getUserD(u)
	if err != nil {
		return nil, err
	}

	curGp, ok := user.Guilds[gp.TopGuild]
	if !ok {
		return nil, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in the guild"}
	}

	curGp.Permissions = gp.Permissions
	tmp := *user
	return &tmp, nil
}

func (udb *UserMemoryDb) SetUserSubGuild(u string, gp *database.GuildPermission) (*database.User, error) {
	user, err := udb.getUserD(u)
	if err != nil {
		return nil, err
	}

	curGp, ok := user.Guilds[gp.TopGuild]
	if !ok {
		return nil, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in the guild"}
	}

	curGp.GuildId = gp.GuildId
	tmp := *user
	return &tmp, nil
}

func (udb *UserMemoryDb) RemoveUserD(u string, g string) (*database.User, error) {
	user, err := udb.getUserD(u)
	if err != nil {
		return nil, err
	}

	if _, ok := user.Guilds[g]; !ok {
		return nil, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in the guild"}
	}

	delete(user.Guilds, g)

	tmp := *user
	return &tmp, nil
}

func (udb *UserMemoryDb) EraseUserD(u string) (*database.User, error) {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	user, err := udb.getUserD(u)
	if err != nil {
		return nil, err
	}

	delete(udb.UsersD, u)
	tmp := *user
	return &tmp, nil
}

func (udb *UserMemoryDb) getUserD(d string) (*database.User, error) {
	if user, ok := udb.UsersD[d]; ok {
		return user, nil
	}

	return nil, &database.Error{Code: database.UserNotFound, Message: "User was not found"}
}
