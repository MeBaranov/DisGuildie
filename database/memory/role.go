package memory

import (
	"fmt"
	"sync"

	"github.com/mebaranov/disguildie/database"
)

type RoleMemoryDb struct {
	roles map[string]*database.Role
	mux   sync.Mutex
}

func (rdb *RoleMemoryDb) AddRole(r *database.Role) (*database.Role, error) {
	rdb.mux.Lock()
	defer rdb.mux.Unlock()

	id := getRoleId(r.GuildId, r.Id)
	if _, ok := rdb.roles[id]; ok {
		return nil, &database.DbError{Code: database.RoleAlreadyExists, Message: "Role with that name already exists"}
	}

	rdb.roles[id] = r
	return r, nil
}

func (rdb *RoleMemoryDb) GetRole(g string, r string) (*database.Role, error) {
	id := getRoleId(g, r)
	if r, ok := rdb.roles[id]; ok {
		return r, nil
	}

	return nil, &database.DbError{Code: database.NoMainCharacterSpecified, Message: "No main character specified"}
}

func (rdb *RoleMemoryDb) GetGuildRoles(g string) ([]*database.Role, error) {
	rv := make([]*database.Role, 0, 10)
	for _, r := range rdb.roles {
		if r.GuildId == g {
			rv = append(rv, r)
		}
	}

	return rv, nil
}

func (rdb *RoleMemoryDb) SetRolePermissions(g string, r string, p int) (*database.Role, error) {
	id := getRoleId(g, r)
	role, ok := rdb.roles[id]

	if !ok {
		return nil, &database.DbError{Code: database.RoleNotFound, Message: "Role was not found"}
	}

	role.Permissions = p
	return role, nil
}

func (rdb *RoleMemoryDb) RemoveRole(g string, r string) (*database.Role, error) {
	rdb.mux.Lock()
	defer rdb.mux.Unlock()

	id := getRoleId(g, r)
	role, ok := rdb.roles[id]

	if !ok {
		return nil, nil
	}

	delete(rdb.roles, id)
	return role, nil
}

func getRoleId(g string, r string) string {
	return fmt.Sprintf("%v:%v", g, r)
}
