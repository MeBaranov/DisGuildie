package memory

import (
	"sync"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
)

type UserMemoryDb struct {
	users  map[uuid.UUID]*database.User
	usersD map[string]*database.User
	mux    sync.Mutex
}

func (udb *UserMemoryDb) AddUser(u *database.User) *database.User {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	if user, ok := udb.usersD[u.DiscordId]; ok {
		for val, _ := range u.GuildIds {
			if _, e := user.GuildIds[val]; !e {
				user.GuildIds[val] = database.Member
			}
		}

		return user
	}

	uid := uuid.New()
	u.UserId = uid
	udb.users[uid] = u
	udb.usersD[u.DiscordId] = u

	return u
}

func (udb *UserMemoryDb) GetUser(d string, g uuid.UUID) *database.User {
	if user, ok := udb.usersD[d]; ok {
		return user
	}

	return nil
}

func (udb *UserMemoryDb) SetUserPermissions(u uuid.UUID, p int) {
	if user, ok := udb.users[u]; ok {
		user.Permissions = p
	}
}

func (udb *UserMemoryDb) RemoveUser(u uuid.UUID) *database.User {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	user, ok := udb.users[u]
	if !ok {
		return nil
	}

	delete(udb.users, u)
	delete(udb.usersD, user.DiscordId)

	return user
}

func (udb *UserMemoryDb) RemoveUserD(d string) *database.User {
	udb.mux.Lock()
	defer udb.mux.Unlock()

	user, ok := udb.usersD[d]
	if !ok {
		return nil
	}

	delete(udb.users, user.UserId)
	delete(udb.usersD, d)

	return user
}
