package memory_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/database/memory"
)

// Here we should test character DB
var testable map[string]database.DataProvider = map[string]database.DataProvider{
	"memory": memory.NewMemoryDb(),
}

func TestAdd(t *testing.T) {
	for n, d := range testable {
		id := uuid.New()

		c := &database.Character{
			Name:   "test",
			UserId: id,
		}
		rc, err := d.AddCharacter(c)

		if err != nil {
			t.Fatalf("[%v] No errors expected adding character. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c = &database.Character{
			Name:   "test2",
			UserId: id,
		}
		rc, err = d.AddCharacter(c)

		if err != nil {
			t.Fatalf("[%v] No errors expected adding second character. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c = &database.Character{
			Name:   "test2",
			UserId: id,
		}
		rc, err = d.AddCharacter(c)

		if err == nil {
			t.Fatalf("[%v] Errors expected adding third character. Received: %v", n, rc)
		}
		assertError(t, err, fmt.Sprintf("User already has character with name %v", c.Name), database.CharacterNameTaken, n)

		c = &database.Character{
			Name:   "test2",
			UserId: uuid.New(),
		}
		rc, err = d.AddCharacter(c)

		if err != nil {
			t.Fatalf("[%v] No errors expected adding third character. Received: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong third character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestGet(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		rc, err := d.GetCharacter(u, name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "Character with name test was not found", database.CharacterNotFound, n)

		c := &database.Character{
			Name:   name,
			UserId: u,
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacter(u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.GetCharacter(u, "test2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "Character with name test2 was not found", database.CharacterNotFound, n)

		rc, err = d.GetCharacter(uuid.New(), name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "Character with name test was not found", database.CharacterNotFound, n)
	}
}

func TestGetMain(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		rc, err := d.GetMainCharacter(u)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "No Characters found", database.CharacterNotFound, n)

		c := &database.Character{
			Name:   "test3",
			UserId: u,
		}
		d.AddCharacter(c)

		rc, err = d.GetMainCharacter(u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c = &database.Character{
			Name:   name,
			UserId: u,
			Main:   true,
		}
		d.AddCharacter(c)

		rc, err = d.GetMainCharacter(u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c2 := &database.Character{
			Name:   name,
			UserId: u,
		}
		d.AddCharacter(c2)

		rc, err = d.GetMainCharacter(u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestGetNameless(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		rc, err := d.GetCharacter(u, "")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "No Characters found", database.CharacterNotFound, n)

		c := &database.Character{
			Name:   "test3",
			UserId: u,
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacter(u, "")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c = &database.Character{
			Name:   name,
			UserId: u,
			Main:   true,
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacter(u, "")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c2 := &database.Character{
			Name:   name,
			UserId: u,
		}
		d.AddCharacter(c2)

		rc, err = d.GetCharacter(u, "")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestGets(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		rc, err := d.GetCharacters(u)
		if err != nil {
			t.Fatalf("[%v] No error expected getting empty list of characters. Got: %v", n, err)
		}
		if rc != nil && len(rc) > 0 {
			t.Fatalf("[%v] Expected nil or empty array. Got: %v", n, rc)
		}

		c := &database.Character{
			Name:   "test3",
			UserId: u,
		}
		d.AddCharacter(c)

		rc, err = d.GetCharacters(u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if len(rc) != 1 {
			t.Fatalf("[%v] Expected single character. Got: %v", n, rc)
		}
		if rc[0].Name != c.Name || rc[0].UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c2 := &database.Character{
			Name:   name,
			UserId: u,
			Main:   true,
		}
		d.AddCharacter(c2)

		c3 := &database.Character{
			Name:   name,
			UserId: uuid.New(),
		}
		d.AddCharacter(c3)

		rc, err = d.GetCharacters(u)
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

		rc, err = d.GetCharacters(c3.UserId)
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

func TestRename(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		c := &database.Character{
			Name:   name,
			UserId: u,
		}
		d.AddCharacter(c)

		rc, err := d.RenameCharacter(u, name, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.GetCharacter(u, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.GetCharacter(u, name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "Character with name test was not found", database.CharacterNotFound, n)

		c = &database.Character{
			Name:   name,
			UserId: u,
		}
		d.AddCharacter(c)

		rc, err = d.RenameCharacter(u, name, "test2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "Character with that name already exists", database.CharacterNameTaken, n)

		rc, err = d.GetCharacter(u, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.GetCharacter(u, name)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		c = &database.Character{
			Name:   name,
			UserId: uuid.New(),
		}
		d.AddCharacter(c)

		rc, err = d.RenameCharacter(c.UserId, name, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != "test2" || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}
	}
}

func TestChangeMain(t *testing.T) {
	for n, d := range testable {
		u, name := uuid.New(), "test"

		c := &database.Character{
			Name:   "test2",
			UserId: u,
		}
		d.AddCharacter(c)

		c2 := &database.Character{
			Name:   name,
			UserId: u,
			Main:   true,
		}
		d.AddCharacter(c2)

		rc, err := d.ChangeMainCharacter(u, "test2")
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}

		rc, err = d.GetMainCharacter(u)
		if err != nil {
			t.Fatalf("[%v] Error not expected. Got: %v", n, err)
		}
		if rc.Name != c.Name || rc.UserId != c.UserId {
			t.Fatalf("[%v] Wrong second character returned. Actual: %v, expected: %v", n, rc, c)
		}

		rc, err = d.ChangeMainCharacter(u, "test3")
		if err == nil {
			t.Fatalf("[%v] Error expected. Got: %v", n, rc)
		}
		assertError(t, err, "Character with name test3 was not found", database.CharacterNotFound, n)
	}
}

func assertError(t *testing.T, e error, message string, code database.ErrorCode, dbn string) {
	wish := fmt.Sprintf("Error '%v': %v", code, message)
	if e.Error() != wish {
		t.Fatalf("[%v] Wrong error message. Actual: %v, Expected: %v", dbn, e.Error(), wish)
	}
}

/*
	SetCharacterStat(u uuid.UUID, name string, s string, v interface{}) (*Character, error)
	ChangeCharacterOwner(old uuid.UUID, name string, u uuid.UUID) (*Character, error)
	RemoveCharacterStat(u uuid.UUID, name string, s string) (*Character, error)
	RemoveCharacter(u uuid.UUID, name string) (*Character, error)
*/
