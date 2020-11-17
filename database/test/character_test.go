package database_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
)

func TestCharAdd(t *testing.T) {
	for n, d := range testable {
		gid, id := uuid.New().String(), uuid.New().String()

		c := &database.Character{
			GuildId: gid,
			UserId:  id,
			Name:    "test",
		}
		rc, err := d.AddCharacter(c)

		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		c = &database.Character{
			GuildId: gid,
			UserId:  id,
			Name:    "test2",
		}
		rc, err = d.AddCharacter(c)

		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		c = &database.Character{
			GuildId: gid,
			UserId:  id,
			Name:    "test2",
		}
		rc, err = d.AddCharacter(c)

		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, fmt.Sprintf("User already has character with name %v", c.Name), database.CharacterNameTaken, n); e != "" {
			t.Fatalf(e)
		}

		c = &database.Character{
			GuildId: gid,
			UserId:  uuid.New().String(),
			Name:    "test2",
		}
		rc, err = d.AddCharacter(c)

		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong third character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}
	}
}

func TestCharGet(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		rc, err := d.GetCharacter(g, u, name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatalf(e)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rc, err = d.GetCharacter(g, u, "test2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test2 was not found", database.CharacterNotFound, n); e != "" {
			t.Fatalf(e)
		}

		rc, err = d.GetCharacter(g, uuid.New().String(), name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatalf(e)
		}

		rc, err = d.GetCharacter(uuid.New().String(), u, name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestCharGetMain(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		rc, err := d.GetMainCharacter(g, u)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "No Characters found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    "test3",
		}
		d.AddCharacter(c)

		rc, err = d.GetMainCharacter(g, u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		c = &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
			Main:    true,
		}
		d.AddCharacter(c)

		rc, err = d.GetMainCharacter(g, u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		c2 := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c2)

		rc, err = d.GetMainCharacter(g, u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestCharGetNameless(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		rc, err := d.GetCharacter(g, u, "")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "No Characters found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    "test3",
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacter(g, u, "")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		c = &database.Character{
			GuildId: g,
			Name:    name,
			UserId:  u,
			Main:    true,
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacter(g, u, "")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c2 := &database.Character{
			GuildId: g,
			Name:    name,
			UserId:  u,
		}
		d.AddCharacter(c2)

		rc, err = d.GetCharacter(g, u, "")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestCharGets(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		rc, err := d.GetCharacters(g, u)
		if err != nil {
			t.Fatalf("[%v] No error expected getting empty list of characters. Got: %v", n, err)
		}
		if rc != nil && len(rc) > 0 {
			t.Fatalf("[%v] Expected nil or empty array. Got: %v", n, rc)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    "test3",
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacters(g, u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if len(rc) != 1 {
			t.Fatalf("[%v] Expected single character. Got: %v", n, rc)
		}
		if rc[0].Name != c.Name || rc[0].UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc[0] == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		c2 := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
			Main:    true,
		}
		d.AddCharacter(c2)

		c3 := &database.Character{
			GuildId: g,
			UserId:  uuid.New().String(),
			Name:    name,
		}
		d.AddCharacter(c3)

		rc, err = d.GetCharacters(g, u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if len(rc) != 2 {
			t.Fatalf("[%v] Expected two characters. Got: %v", n, rc)
		}

		fi, se := false, false
		for _, r := range rc {
			fi = fi || (r.Name == c.Name && r.UserId == c.UserId)
			se = se || (r.Name == c2.Name && r.UserId == c2.UserId)
		}
		if !fi || !se {
			t.Fatalf("[%v] Expected both characters to be present. Got: %v ", n, rc)
		}

		rc, err = d.GetCharacters(c3.GuildId, c3.UserId)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if len(rc) != 1 {
			t.Fatalf("[%v] Expected single character. Got: %v", n, rc)
		}
		if rc[0].Name != c3.Name || rc[0].UserId != c3.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c3)
		}
	}
}

func TestCharRename(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c)

		rc, err := d.RenameCharacter(g, u, name, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rc, err = d.GetCharacter(g, u, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.GetCharacter(g, u, name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c = &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c)

		rc, err = d.RenameCharacter(g, u, name, "test2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with that name already exists", database.CharacterNameTaken, n); e != "" {
			t.Fatal(e)
		}

		rc, err = d.GetCharacter(g, u, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.GetCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c = &database.Character{
			GuildId: g,
			UserId:  uuid.New().String(),
			Name:    name,
		}
		d.AddCharacter(c)

		rc, err = d.RenameCharacter(c.GuildId, c.UserId, name, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestCharChangeOwner(t *testing.T) {
	for n, d := range testable {
		g, u, name, u2 := uuid.New().String(), uuid.New().String(), "test", uuid.New().String()

		rc, err := d.ChangeCharacterOwner(g, u, name, u2)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c)

		rc, err = d.ChangeCharacterOwner(g, u, name, u2)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != name || rc.UserId != u2 {
			t.Fatalf("[%v] Wrong character returned. Actual: %v, expected: %v", n, *rc, *c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rcs, err := d.GetCharacters(g, u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if len(rcs) != 0 {
			t.Fatalf("[%v] Wrong characters amount returned. Actual: %v, expected: %v", n, rcs, "empty")
		}

		rcs, err = d.GetCharacters(g, u2)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if len(rcs) != 1 {
			t.Fatalf("[%v] Wrong characters amount returned. Actual: %v, expected: %v", n, rcs, 1)
		}

		rc, err = d.GetCharacter(g, u, name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c = &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c)

		rc, err = d.ChangeCharacterOwner(g, u, name, u2)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Target user already has character with name 'test'", database.UserHasCharacter, n); e != "" {
			t.Fatal(e)
		}

		rc, err = d.GetCharacter(g, u2, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != name || rc.UserId != u2 {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.GetCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != name || rc.UserId != u {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestCharChangeMain(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    "test2",
		}
		c, _ = d.AddCharacter(c)

		c2 := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
			Main:    true,
		}
		c2, _ = d.AddCharacter(c2)

		rc, err := d.ChangeMainCharacter(g, u, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}

		rc, err = d.GetMainCharacter(g, u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rc, err = d.ChangeMainCharacter(g, u, "test3")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test3 was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}
	}
}

func TestCharSetStat(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		rc, err := d.SetCharacterStat(g, u, name, "a", "b")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c)

		current := make(map[string]interface{})

		rc, err = d.GetCharacter(g, u, name)
		if rc.Body != nil && !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		current["t1"] = "str"
		rc, err = d.SetCharacterStat(g, u, name, "t1", "str")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		current["t2"] = 5
		rc, err = d.SetCharacterStat(g, u, name, "t2", 5)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		current["t1"] = 10
		rc, err = d.SetCharacterStat(g, u, name, "t1", 10)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		current["t2"] = "str2"
		rc, err = d.SetCharacterStat(g, u, name, "t2", "str2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.GetCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}
	}
}

func TestCharRemoveStat(t *testing.T) {
	for n, d := range testable {
		g, u, name := uuid.New().String(), uuid.New().String(), "test"

		rc, err := d.RemoveCharacterStat(g, u, name, "a")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		if e := assertError(err, "Character with name test was not found", database.CharacterNotFound, n); e != "" {
			t.Fatal(e)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		d.AddCharacter(c)

		current := make(map[string]interface{})

		rc, err = d.SetCharacterStat(g, u, name, "t1", "str")
		rc, err = d.SetCharacterStat(g, u, name, "t2", 5)
		current["t1"] = 10
		rc, err = d.SetCharacterStat(g, u, name, "t1", 10)
		current["t2"] = "str2"
		rc, err = d.SetCharacterStat(g, u, name, "t2", "str2")

		rc, err = d.RemoveCharacterStat(g, u, name, "a")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		delete(current, "t1")
		rc, err = d.RemoveCharacterStat(g, u, name, "t1")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.GetCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		delete(current, "t2")
		rc, err = d.RemoveCharacterStat(g, u, name, "t2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.RemoveCharacterStat(g, u, name, "t3")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}

		rc, err = d.GetCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if !reflect.DeepEqual(rc.Body, current) {
			t.Fatalf("[%v] Unexpected stats. Actual: %v. Expected: %v", n, rc.Body, current)
		}
	}
}

func TestCharRemove(t *testing.T) {
	for n, d := range testable {
		g, u, name, name2 := uuid.New().String(), uuid.New().String(), "test", "test2"

		rc, err := d.RemoveCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] No errors expected adding character. Received: %v", n, err)
		}
		if rc != nil {
			t.Fatalf("[%v] No character expected. Actual: %v", n, rc)
		}

		c := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name,
		}
		rc, err = d.AddCharacter(c)

		c2 := &database.Character{
			GuildId: g,
			UserId:  u,
			Name:    name2,
		}
		rc, err = d.AddCharacter(c2)

		rc, err = d.RemoveCharacter(g, u, name)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong character returned. Actual: %v, expected: %v", n, rc, c)
		}
		if rc == c {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rcs, err := d.GetCharacters(g, u)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if len(rcs) != 1 {
			t.Fatalf("[%v] Wrong character count. Actual: %v, expected: %v", n, rcs, 1)
		}
		if rcs[0].Name != c2.Name || rcs[0].UserId != c2.UserId {
			t.Fatalf("[%v] Wrong character is kept. Actual: %v, expected: %v", n, rcs, *c2)
		}

		rc, err = d.RemoveCharacter(g, u, name2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != c2.Name || rc.UserId != c2.UserId {
			t.Fatalf("[%v] Wrong character returned. Actual: %v, expected: %v", n, *rc, *c2)
		}

		rcs, err = d.GetCharacters(g, u)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rcs != nil && len(rcs) != 0 {
			t.Fatalf("[%v] Wrong character count. Actual: %v, expected: %v", n, rcs, 0)
		}
	}
}
