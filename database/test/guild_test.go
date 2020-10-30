package database_test

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
)

func TestGuildAdd(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did1",
		}

		rc1, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc1.Name != g.Name || rc1.DiscordId != g.DiscordId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc1, g)
		}

		g = &database.Guild{
			Name:      "test2",
			DiscordId: "did2",
		}
		rc2, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc2.Name != g.Name || rc2.DiscordId != g.DiscordId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc2, g)
		}

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc3, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc3.Name != g.Name || rc3.ParentId != rc1.GuildId || rc3.TopLevelParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc3, g)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc3.GuildId,
		}
		rc4, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc4.Name != g.Name || rc4.ParentId != rc3.GuildId || rc4.TopLevelParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc4, g)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc5, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc5.Name != g.Name || rc5.ParentId != rc2.GuildId || rc5.TopLevelParentId != rc2.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc4, g)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc4.GuildId,
		}
		rce, err := d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Sub-Guild name 'sub2-test1' is already taken", database.SubguildNameTaken, n); e != "" {
			t.Fatalf(e)
		}

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc4.GuildId,
		}
		rce, err = d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Sub-Guild name 'sub1-test1' is already taken", database.SubguildNameTaken, n); e != "" {
			t.Fatalf(e)
		}

		g = &database.Guild{
			Name:     "sub10-test10",
			ParentId: uuid.New(),
		}
		rce, err = d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Invalid parent guild ID", database.InvalidGuildDefinition, n); e != "" {
			t.Fatalf(e)
		}

		g = &database.Guild{
			Name:      "sub10-test10",
			DiscordId: "did1",
		}
		rce, err = d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild 'did1' is already registered", database.GuildAlreadyRegistered, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestGuildGet(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did1",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc3, err := d.AddGuild(g)

		g, err = d.GetGuild(rc3.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc3.Name != g.Name || g.ParentId != rc2.GuildId || g.TopLevelParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, g, rc3)
		}

		names := make(map[string]database.Void)
		names["sub1-test1"] = database.Member
		names["sub2-test1"] = database.Member
		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc1.Name != g.Name || rc1.DiscordId != g.DiscordId || !reflect.DeepEqual(g.ChildNames, names) {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v, with child names: %v", n, g, rc1, names)
		}

		rce, err := d.GetGuild(uuid.New())
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestGuildGetD(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did1",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		d.AddGuild(g)

		names := make(map[string]database.Void)
		names["sub1-test1"] = database.Member
		names["sub2-test1"] = database.Member
		g, err = d.GetGuildD("did1")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc1.Name != g.Name || rc1.DiscordId != g.DiscordId || !reflect.DeepEqual(g.ChildNames, names) {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v, with child names: %v", n, g, rc1, names)
		}

		rce, err := d.GetGuildD("did2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestGuildRename(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did1",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		d.AddGuild(g)

		names := make(map[string]database.Void)
		names["sub1-test2"] = database.Member
		names["sub2-test1"] = database.Member

		rc, err := d.RenameGuild(rc2.GuildId, "sub1-test2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != "sub1-test2" || rc.GuildId != rc2.GuildId || rc.ParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: name: %v, parentId: %v", n, rc, "sub1-test2", rc1.GuildId)
		}

		rce, err := d.RenameGuild(uuid.New(), "test3")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		rce, err = d.RenameGuild(rc2.GuildId, "sub2-test1")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Sub-Guild name 'sub2-test1' is already taken", database.SubguildNameTaken, n); e != "" {
			t.Fatalf(e)
		}

		rc, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != "test" || rc.DiscordId != "did1" || !reflect.DeepEqual(rc.ChildNames, names) {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v, with child names: %v", n, g, rc1, names)
		}
	}
}

func TestGuildSetStat(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		rc, err := d.SetCharacterStat(u, name, "a", "b")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c := &database.Character{
			Name:   name,
			UserId: u,
		}
		d.AddCharacter(c)

		current := make(map[string]interface{})

		rc, err = d.GetCharacter(u, name)
		if rc.Body != nil && !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		current["t1"] = "str"
		rc, err = d.SetCharacterStat(u, name, "t1", "str")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		current["t2"] = 5
		rc, err = d.SetCharacterStat(u, name, "t2", 5)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		current["t1"] = 10
		rc, err = d.SetCharacterStat(u, name, "t1", 10)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		current["t2"] = "str2"
		rc, err = d.SetCharacterStat(u, name, "t2", "str2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.GetCharacter(u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}
	}
}

func TestGuildRemoveStat(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		rc, err := d.RemoveCharacterStat(u, name, "a")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c := &database.Character{
			Name:   name,
			UserId: u,
		}
		d.AddCharacter(c)

		current := make(map[string]interface{})

		rc, err = d.SetCharacterStat(u, name, "t1", "str")
		rc, err = d.SetCharacterStat(u, name, "t2", 5)
		current["t1"] = 10
		rc, err = d.SetCharacterStat(u, name, "t1", 10)
		current["t2"] = "str2"
		rc, err = d.SetCharacterStat(u, name, "t2", "str2")

		rc, err = d.RemoveCharacterStat(u, name, "a")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		delete(current, "t1")
		rc, err = d.RemoveCharacterStat(u, name, "t1")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.GetCharacter(u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		delete(current, "t2")
		rc, err = d.RemoveCharacterStat(u, name, "t2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.RemoveCharacterStat(u, name, "t3")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.GetCharacter(u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}
	}
}

func TestGuildRemove(t *testing.T) {
	for n, d := range testable {
		u, name, name2 := uuid.New(), "test", "test2"

		rc, err := d.RemoveCharacter(u, name)
		if err != nil {
			t.Fatalf("[%v] No errors expected adding character. Received: %v", n, err)
		}
		if rc != nil {
			t.Fatalf("[%v] No character expected. Actual: %v", n, rc)
		}

		c := &database.Character{
			Name:   name,
			UserId: u,
		}
		rc, err = d.AddCharacter(c)

		c2 := &database.Character{
			Name:   name2,
			UserId: u,
		}
		rc, err = d.AddCharacter(c2)

		rc, err = d.RemoveCharacter(u, name)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rcs, err := d.GetCharacters(u)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if len(rcs) != 1 {
			t.Fatalf("[%v] Wrong character count. Actual: %v, expected: %v", n, rcs, 1)
		}
		if rcs[0].Name != c2.Name || rcs[0].UserId != c2.UserId {
			t.Fatalf("[%v] Wrong character is kept. Actual: %v, expected: %v", n, rcs, *c2)
		}

		rc, err = d.RemoveCharacter(u, name2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != c2.Name || rc.UserId != c2.UserId {
			t.Fatalf("[%v] Wrong character returned. Actual: %v, expected: %v", n, *rc, *c2)
		}

		rcs, err = d.GetCharacters(u)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 0 {
			t.Fatalf("[%v] Wrong character count. Actual: %v, expected: %v", n, rcs, 0)
		}
	}
}

/*
	AddGuildStat(g uuid.UUID, n string, t string) (*Guild, error)
	RemoveGuildStat(g uuid.UUID, n string) (*Guild, error)
	RemoveGuild(g uuid.UUID) (*Guild, error)
	RemoveGuildD(d string) (*Guild, error)
*/
