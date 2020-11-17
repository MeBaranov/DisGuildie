package database_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
)

func TestUserAdd(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			Id: "did1",
		}

		perm := &database.GuildPermission{TopGuild: "gdid11", GuildId: uuid.New(), Permissions: 10}
		guilds := map[string]*database.GuildPermission{perm.TopGuild: perm}
		rc, err := d.AddUser(u, perm)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
		if rc == u {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rc, err = d.AddUser(u, perm)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "The user is already registered in the guild", database.UserAlreadyInGuild, n); e != "" {
			t.Fatalf(e)
		}

		perm2 := &database.GuildPermission{TopGuild: "gdid12", GuildId: uuid.New(), Permissions: 11}
		guilds[perm2.TopGuild] = perm2
		rc, err = d.AddUser(u, perm2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		u = &database.User{
			Id: "did12",
		}
		_, err = d.AddUser(u, perm)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		rc, err = d.AddUser(u, perm2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserGetD(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			Id: "did2",
		}

		rc, err := d.GetUserD("did2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		perm := &database.GuildPermission{TopGuild: "gdid21", GuildId: uuid.New(), Permissions: 10}
		guilds := map[string]*database.GuildPermission{perm.TopGuild: perm}
		d.AddUser(u, perm)
		rc, err = d.GetUserD("did2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
		if rc == u {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		d.AddUser(u, perm)
		rc, err = d.GetUserD("did2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		perm2 := &database.GuildPermission{TopGuild: "gdid22", GuildId: uuid.New(), Permissions: 11}
		guilds[perm2.TopGuild] = perm2
		d.AddUser(u, perm2)
		rc, err = d.GetUserD("did2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		u = &database.User{
			Id: "did22",
		}
		d.AddUser(u, perm)
		d.AddUser(u, perm2)
		rc, err = d.GetUserD("did22")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserGetInGuild(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			Id: "TestUserGetInGuild_udid",
		}
		perm := &database.GuildPermission{TopGuild: "TestUserGetInGuild_gdid", GuildId: uuid.New(), Permissions: 10}
		perm2 := &database.GuildPermission{TopGuild: "TestUserGetInGuild_gdid2", GuildId: uuid.New(), Permissions: 11}

		rcs, err := d.GetUsersInGuild("TestUserGetInGuild_gdid")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 0 {
			t.Fatalf("[%v] Wrong amount of users returned. Received: %v, expected: %v", n, len(rcs), 0)
		}

		d.AddUser(u, perm)

		rcs, err = d.GetUsersInGuild(perm.TopGuild)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 1 {
			t.Fatalf("[%v] Wrong amount of users returned. Received: %v, expected: %v", n, len(rcs), 1)
		}
		if rcs[0] == u {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		d.AddUser(u, perm2)

		rcs, err = d.GetUsersInGuild(perm2.TopGuild)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 1 {
			t.Fatalf("[%v] Wrong amount of users returned. Received: %v, expected: %v", n, len(rcs), 1)
		}

		u = &database.User{
			Id: "did22",
		}
		d.AddUser(u, perm)
		rcs, err = d.GetUsersInGuild(perm.TopGuild)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 2 {
			t.Fatalf("[%v] Wrong amount of users returned. Received: %v, expected: %v", n, len(rcs), 2)
		}

		rcs, err = d.GetUsersInGuild(perm2.TopGuild)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 1 {
			t.Fatalf("[%v] Wrong amount of users returned. Received: %v, expected: %v", n, len(rcs), 1)
		}
	}
}

func TestUserSetPermissions(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			Id: "did3",
		}

		perm := &database.GuildPermission{TopGuild: "gdid31", GuildId: uuid.New(), Permissions: 10}
		perm2 := &database.GuildPermission{TopGuild: "gdid32", GuildId: uuid.New(), Permissions: 11}
		rc, err := d.SetUserPermissions("did3", perm)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, perm)

		rc, err = d.SetUserPermissions("did3", perm2)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User is not registered in the guild", database.UserNotInGuild, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, perm2)

		perm = &database.GuildPermission{TopGuild: "gdid31", GuildId: perm.GuildId, Permissions: 100}
		guilds := map[string]*database.GuildPermission{
			perm.TopGuild:  perm,
			perm2.TopGuild: perm2,
		}
		rc, err = d.SetUserPermissions("did3", perm)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
		if rc == u {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		perm2 = &database.GuildPermission{TopGuild: "gdid32", GuildId: perm2.GuildId, Permissions: 1000}
		guilds[perm2.TopGuild] = perm2
		rc, err = d.SetUserPermissions("did3", perm2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserSetSubguild(t *testing.T) {
	for n, d := range testable {
		udid := "did7"
		u := &database.User{
			Id: udid,
		}

		perm := &database.GuildPermission{TopGuild: "gdid71", GuildId: uuid.New(), Permissions: 10}
		perm2 := &database.GuildPermission{TopGuild: "gdid72", GuildId: uuid.New(), Permissions: 10}
		rc, err := d.SetUserSubGuild(udid, perm)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, perm)

		rc, err = d.SetUserSubGuild(udid, perm2)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User is not registered in the guild", database.UserNotInGuild, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, perm2)

		perm = &database.GuildPermission{TopGuild: "gdid71", GuildId: uuid.New(), Permissions: 10}
		guilds := map[string]*database.GuildPermission{
			perm.TopGuild:  perm,
			perm2.TopGuild: perm2,
		}
		rc, err = d.SetUserSubGuild(udid, perm)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
		if rc == u {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		perm2 = &database.GuildPermission{TopGuild: "gdid72", GuildId: uuid.New(), Permissions: 10}
		guilds[perm2.TopGuild] = perm2
		rc, err = d.SetUserSubGuild(udid, perm2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserRemove(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			Id: "did4",
		}

		perm := &database.GuildPermission{TopGuild: "gdid41", GuildId: uuid.New(), Permissions: 10}
		perm2 := &database.GuildPermission{TopGuild: "gdid42", GuildId: uuid.New(), Permissions: 11}
		rc, err := d.RemoveUserD("did4", perm.TopGuild)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, perm)

		rc, err = d.RemoveUserD("did4", perm2.TopGuild)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User is not registered in the guild", database.UserNotInGuild, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, perm2)

		guilds := map[string]*database.GuildPermission{perm2.TopGuild: perm2}
		rc, err = d.RemoveUserD("did4", perm.TopGuild)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
		if rc == u {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		delete(guilds, perm2.TopGuild)
		rc, err = d.RemoveUserD("did4", perm2.TopGuild)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserErase(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			Id: "did5",
		}

		rc, err := d.EraseUserD("did5")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		perm := &database.GuildPermission{TopGuild: "gdid51", GuildId: uuid.New(), Permissions: 10}
		d.AddUser(u, perm)
		guilds := map[string]*database.GuildPermission{perm.TopGuild: perm}

		rc, err = d.EraseUserD("did5")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
		if rc == u {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rc, err = d.GetUserD("did5")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		perm = &database.GuildPermission{TopGuild: "gdid51", GuildId: uuid.New(), Permissions: 11}
		guilds[perm.TopGuild] = perm
		rc, err = d.AddUser(u, perm)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Id != u.Id {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !guildSetsEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func guildSetsEqual(a map[string]*database.GuildPermission, b map[string]*database.GuildPermission) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		if vb, ok := b[k]; !ok || !permissionsEqual(va, vb) {
			return false
		}
	}

	return true
}

func permissionsEqual(a *database.GuildPermission, b *database.GuildPermission) bool {
	return a.GuildId == b.GuildId && a.Permissions == b.Permissions && a.TopGuild == b.TopGuild
}
