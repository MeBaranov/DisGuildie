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

func (udb *UserMemoryDb) AddUser(u *database.User, g uuid.UUID, p int) (*database.User, error) {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	if user, ok := udb.usersD[u.DiscordId]; ok {
		if _, ok = user.Guilds[g]; !ok {
			user.Guilds[g] = p
			return user, nil
		}

		return nil, &database.Error{Code: database.UserAlreadyInGuild, Message: "The user is already registered in the guild"}
	}

	uid := uuid.New()
	u.UserId = uid
	u.Guilds = map[uuid.UUID]int{g: p}
	udb.usersD[u.DiscordId] = u

	return u, nil
}

func (udb *UserMemoryDb) GetUserD(d string) (*database.User, error) {
	if user, ok := udb.usersD[d]; ok {
		return user, nil
	}

	return nil, &database.Error{Code: database.UserNotFound, Message: "User was not found"}
}

func (udb *UserMemoryDb) SetUserPermissions(u string, g uuid.UUID, p int) (*database.User, error) {
	user, err := udb.GetUserD(u)
	if err != nil {
		return nil, err
	}

	if _, ok := user.Guilds[g]; !ok {
		return nil, &database.Error{Code: database.UserNotInGuild, Message: "User is not registered in the guild"}
	}

	user.Guilds[g] = p
	return user, nil
}

func (udb *UserMemoryDb) RemoveUserD(u string, g uuid.UUID) (*database.User, error) {
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
