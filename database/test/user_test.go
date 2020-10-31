package database_test

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
)

func TestUserAdd(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			DiscordId: "did1",
		}

		gid := uuid.New()
		guilds := map[uuid.UUID]int{gid: 10}
		rc, err := d.AddUser(u, gid, 10)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		rc, err = d.AddUser(u, gid, 10)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "The user is already registered in the guild", database.UserAlreadyInGuild, n); e != "" {
			t.Fatalf(e)
		}

		gid2 := uuid.New()
		guilds[gid2] = 11
		rc, err = d.AddUser(u, gid2, 11)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		u = &database.User{
			DiscordId: "did12",
		}
		_, err = d.AddUser(u, gid, 10)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		rc, err = d.AddUser(u, gid2, 11)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserGetD(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			DiscordId: "did2",
		}

		rc, err := d.GetUserD("did2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		gid := uuid.New()
		guilds := map[uuid.UUID]int{gid: 10}
		d.AddUser(u, gid, 10)
		rc, err = d.GetUserD("did2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		d.AddUser(u, gid, 10)
		rc, err = d.GetUserD("did2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		gid2 := uuid.New()
		guilds[gid2] = 11
		d.AddUser(u, gid2, 11)
		rc, err = d.GetUserD("did2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		u = &database.User{
			DiscordId: "did22",
		}
		d.AddUser(u, gid, 10)
		d.AddUser(u, gid2, 11)
		rc, err = d.GetUserD("did22")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserSetPermissions(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			DiscordId: "did3",
		}

		gid, gid2 := uuid.New(), uuid.New()
		rc, err := d.SetUserPermissions("did3", gid, 100)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, gid, 10)

		rc, err = d.SetUserPermissions("did3", gid2, 100)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User is not registered in the guild", database.UserNotInGuild, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, gid2, 11)

		guilds := map[uuid.UUID]int{
			gid:  100,
			gid2: 11,
		}
		rc, err = d.SetUserPermissions("did3", gid, 100)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		guilds[gid2] = 1000
		rc, err = d.SetUserPermissions("did3", gid2, 1000)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserRemove(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			DiscordId: "did4",
		}

		gid, gid2 := uuid.New(), uuid.New()
		rc, err := d.RemoveUserD("did4", gid)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		guilds := make(map[uuid.UUID]int)

		d.AddUser(u, gid, 10)

		rc, err = d.RemoveUserD("did4", gid2)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User is not registered in the guild", database.UserNotInGuild, n); e != "" {
			t.Fatalf(e)
		}

		d.AddUser(u, gid2, 11)

		guilds[gid2] = 11
		rc, err = d.RemoveUserD("did4", gid)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		delete(guilds, gid2)
		rc, err = d.RemoveUserD("did4", gid2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}

func TestUserErase(t *testing.T) {
	for n, d := range testable {
		u := &database.User{
			DiscordId: "did5",
		}

		rc, err := d.EraseUserD("did5")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		gid := uuid.New()
		d.AddUser(u, gid, 10)
		guilds := map[uuid.UUID]int{gid: 10}

		rc, err = d.EraseUserD("did5")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}

		rc, err = d.GetUserD("did5")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "User was not found", database.UserNotFound, n); e != "" {
			t.Fatalf(e)
		}

		guilds[gid] = 11
		rc, err = d.AddUser(u, gid, 11)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.DiscordId != u.DiscordId {
			t.Fatalf("[%v] Wrong user returned. Actual: %v, expected: %v", n, rc, u)
		}
		if !reflect.DeepEqual(rc.Guilds, guilds) {
			t.Fatalf("[%v] Wrong guild set returned. Actual: %v, expected: %v", n, rc.Guilds, guilds)
		}
	}
}
