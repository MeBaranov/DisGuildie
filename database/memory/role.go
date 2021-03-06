package memory

import (
	"fmt"
	"sync"

	"github.com/mebaranov/disguildie/database"
)

type RoleMemoryDb struct {
	Roles map[string]*database.Role
	mux   sync.Mutex
}

func (rdb *RoleMemoryDb) AddRole(r *database.Role) (*database.Role, error) {
	rdb.mux.Lock()
	defer rdb.mux.Unlock()

	id := getRoleId(r.GuildId, r.Id)
	if _, ok := rdb.Roles[id]; ok {
		return nil, &database.Error{Code: database.RoleAlreadyExists, Message: "Role with this ID already exists in this guild"}
	}

	newR := *r
	r = &newR
	rdb.Roles[id] = r

	tmp := *r
	return &tmp, nil
}

func (rdb *RoleMemoryDb) GetRole(g string, r string) (*database.Role, error) {
	id := getRoleId(g, r)
	if r, ok := rdb.Roles[id]; ok {
		tmp := *r
		return &tmp, nil
	}

	return nil, &database.Error{Code: database.RoleNotFound, Message: "Role was not found"}
}

func (rdb *RoleMemoryDb) GetGuildRoles(g string) ([]*database.Role, error) {
	rv := make([]*database.Role, 0, 10)
	for _, r := range rdb.Roles {
		if r.GuildId == g {
			tmp := *r
			rv = append(rv, &tmp)
		}
	}

	return rv, nil
}

func (rdb *RoleMemoryDb) SetRolePermissions(g string, r string, p int) (*database.Role, error) {
	id := getRoleId(g, r)
	role, ok := rdb.Roles[id]
	if !ok {
		return nil, &database.Error{Code: database.RoleNotFound, Message: "Role was not found"}
	}

	role.Permissions = p
	tmp := *role
	return &tmp, nil
}

func (rdb *RoleMemoryDb) RemoveRole(g string, r string) (*database.Role, error) {
	rdb.mux.Lock()
	defer rdb.mux.Unlock()

	id := getRoleId(g, r)
	role, ok := rdb.Roles[id]

	if !ok {
		return nil, &database.Error{Code: database.RoleNotFound, Message: "Role was not found"}
	}

	delete(rdb.Roles, id)
	tmp := *role
	return &tmp, nil
}

func getRoleId(g string, r string) string {
	return fmt.Sprintf("%v:%v", g, r)
}
