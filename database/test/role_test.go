package database_test

import (
	"testing"

	"github.com/mebaranov/disguildie/database"
)

func TestRoleAdd(t *testing.T) {
	for n, d := range testable {
		r := &database.Role{
			GuildId:     "gid1",
			Id:          "rid1",
			Permissions: 10,
		}

		rc, err := d.AddRole(r)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
		if rc == r {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rc, err = d.AddRole(r)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "Role with this ID already exists in this guild", database.RoleAlreadyExists, n); e != "" {
			t.Fatalf(e)
		}

		r = &database.Role{
			GuildId:     "gid1",
			Id:          "rid12",
			Permissions: 10,
		}
		rc, err = d.AddRole(r)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}

		r = &database.Role{
			GuildId:     "gid12",
			Id:          "rid12",
			Permissions: 10,
		}
		rc, err = d.AddRole(r)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
	}
}

func TestRoleGet(t *testing.T) {
	for n, d := range testable {
		gid, rid := "gid2", "rid2"
		r := &database.Role{
			GuildId:     gid,
			Id:          rid,
			Permissions: 10,
		}

		rc, err := d.GetRole(gid, rid)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "Role was not found", database.RoleNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddRole(r)
		rc, err = d.GetRole(gid, rid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
		if rc == r {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		d.AddRole(r)
		rc, err = d.GetRole(gid, rid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}

		r = &database.Role{
			GuildId:     "gid2",
			Id:          "rid22",
			Permissions: 10,
		}
		d.AddRole(r)
		rc, err = d.GetRole(gid, "rid22")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}

		r = &database.Role{
			GuildId:     "gid22",
			Id:          "rid22",
			Permissions: 10,
		}
		d.AddRole(r)
		rc, err = d.GetRole("gid22", "rid22")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
	}
}

func TestRoleGetGuild(t *testing.T) {
	for n, d := range testable {
		gid, rid := "gid3", "rid3"
		r := &database.Role{
			GuildId:     gid,
			Id:          rid,
			Permissions: 10,
		}

		rcs, err := d.GetGuildRoles(gid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 0 {
			t.Fatalf("[%v] Wrong Roles count returned. Actual: %v, expected: %v", n, rcs, 0)
		}

		d.AddRole(r)
		rcs, err = d.GetGuildRoles(gid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs == nil || len(rcs) != 1 {
			t.Fatalf("[%v] Wrong Roles count returned. Actual: %v, expected: %v", n, rcs, 1)
		}
		rc := rcs[0]
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
		if rc == r {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		r = &database.Role{
			GuildId:     gid,
			Id:          "rid32",
			Permissions: 10,
		}
		d.AddRole(r)
		r = &database.Role{
			GuildId:     "gid32",
			Id:          "rid32",
			Permissions: 10,
		}
		d.AddRole(r)
		rcs, err = d.GetGuildRoles(gid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs == nil || len(rcs) != 2 {
			t.Fatalf("[%v] Wrong Roles count returned. Actual: %v, expected: %v", n, rcs, 2)
		}
	}
}

func TestRoleSetPermisisons(t *testing.T) {
	for n, d := range testable {
		gid, rid := "gid4", "rid4"
		r := &database.Role{
			GuildId:     gid,
			Id:          rid,
			Permissions: 10,
		}

		rc, err := d.SetRolePermissions(gid, rid, 100)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "Role was not found", database.RoleNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddRole(r)
		rc, err = d.SetRolePermissions(gid, rid, 100)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != 100 {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
		if rc == r {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}
	}
}

func TestRoleRemove(t *testing.T) {
	for n, d := range testable {
		gid, rid := "gid5", "rid5"
		r := &database.Role{
			GuildId:     gid,
			Id:          rid,
			Permissions: 10,
		}

		rc, err := d.RemoveRole(gid, rid)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "Role was not found", database.RoleNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddRole(r)
		rc, err = d.RemoveRole(gid, rid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
		if rc == r {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		_, err = d.AddRole(r)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}

		r = &database.Role{
			GuildId:     "gid52",
			Id:          "rid52",
			Permissions: 10,
		}
		_, err = d.AddRole(r)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}

		rc, err = d.RemoveRole("gid52", "rid52")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}

		rc, err = d.GetRole(gid, rid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != r.GuildId && rc.Id != r.Id && rc.Permissions != r.Permissions {
			t.Fatalf("[%v] Wrong Role returned. Actual: %v, expected: %v", n, rc, r)
		}
	}
}
