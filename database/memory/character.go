package memory

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
)

type CharMemoryDb struct {
	chars map[string]*database.Character
	mux   sync.Mutex
}

func (cdb *CharMemoryDb) AddCharacter(c *database.Character) (*database.Character, error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c.CharId = getCharacterId(c.UserId, c.Name)
	if _, ok := cdb.chars[c.CharId]; ok {
		return nil, &database.Error{Code: database.CharacterNameTaken, Message: fmt.Sprintf("User already has character with name %v", c.Name)}
	}

	cdb.chars[c.CharId] = c

	return c, nil
}

func (cdb *CharMemoryDb) GetCharacters(u uuid.UUID) ([]*database.Character, error) {
	rv := make([]*database.Character, 0, 10)
	for _, v := range cdb.chars {
		if v.UserId == u {
			rv = append(rv, v)
		}
	}

	return rv, nil
}

func (cdb *CharMemoryDb) GetMainCharacter(u uuid.UUID) (*database.Character, error) {
	var rv *database.Character = nil
	for _, v := range cdb.chars {
		if v.UserId == u {
			if v.Main {
				return v, nil
			}
			if rv == nil {
				rv = v
			}
		}
	}

	if rv == nil {
		return nil, &database.Error{Code: database.CharacterNotFound, Message: fmt.Sprintf("No Characters found")}
	}

	return rv, nil
}

func (cdb *CharMemoryDb) GetCharacter(u uuid.UUID, name string) (*database.Character, error) {
	if name == "" {
		return cdb.GetMainCharacter(u)
	}

	id := getCharacterId(u, name)
	if c, ok := cdb.chars[id]; ok {
		return c, nil
	}

	return nil, &database.Error{Code: database.CharacterNotFound, Message: fmt.Sprintf("Character with name %v was not found", name)}
}

func (cdb *CharMemoryDb) RenameCharacter(u uuid.UUID, old string, name string) (*database.Character, error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.GetCharacter(u, old)
	if err != nil {
		return nil, err
	}

	_, err = cdb.GetCharacter(u, name)
	if err == nil {
		return nil, &database.Error{Code: database.CharacterNameTaken, Message: "Character with that name already exists"}
	}

	c.Name = name
	delete(cdb.chars, c.CharId)
	c.CharId = getCharacterId(u, name)
	cdb.chars[c.CharId] = c

	return c, nil
}

func (cdb *CharMemoryDb) ChangeMainCharacter(u uuid.UUID, name string) (*database.Character, error) {
	c, err := cdb.GetCharacter(u, name)
	if err != nil {
		return nil, err
	}

	old, err := cdb.GetMainCharacter(u)
	if err == nil {
		old.Main = false
	}

	c.Main = true

	return c, nil
}

func (cdb *CharMemoryDb) SetCharacterStat(u uuid.UUID, name string, s string, v interface{}) (*database.Character, error) {
	c, err := cdb.GetCharacter(u, name)
	if err != nil {
		return nil, err
	}

	c.Body[s] = v
	return c, nil
}

func (cdb *CharMemoryDb) ChangeCharacterOwner(old uuid.UUID, name string, u uuid.UUID) (*database.Character, error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.GetCharacter(old, name)
	if err != nil {
		return nil, err
	}

	_, err = cdb.GetCharacter(u, name)
	if err == nil {
		return nil, &database.Error{Code: database.UserHasCharacter, Message: "Target user already has character with that name"}
	}

	c.UserId = u
	delete(cdb.chars, c.CharId)
	c.CharId = getCharacterId(u, name)
	cdb.chars[c.CharId] = c

	return c, nil
}

func (cdb *CharMemoryDb) RemoveCharacterStat(u uuid.UUID, name string, s string) (*database.Character, error) {
	c, err := cdb.GetCharacter(u, name)
	if err != nil {
		return nil, err
	}

	delete(c.Body, s)
	return c, nil
}

func (cdb *CharMemoryDb) RemoveCharacter(u uuid.UUID, name string) (*database.Character, error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.GetCharacter(u, name)
	if err != nil {
		return nil, nil
	}

	delete(cdb.chars, c.CharId)
	return c, nil
}

func getCharacterId(u uuid.UUID, name string) string {
	return fmt.Sprintf("%v:%v", u, name)
}
