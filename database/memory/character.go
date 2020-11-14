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

func (cdb *CharMemoryDb) AddCharacter(c *database.Character) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c.CharId = getCharacterId(c.UserId, c.Name)
	if _, ok := cdb.chars[c.CharId]; ok {
		return nil, &database.Error{Code: database.CharacterNameTaken, Message: fmt.Sprintf("User already has character with name %v", c.Name)}
	}

	newC := *c
	c = &newC
	cdb.chars[c.CharId] = c

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) GetCharacters(u uuid.UUID) ([]*database.Character, *database.Error) {
	rv := make([]*database.Character, 0, 10)
	for _, v := range cdb.chars {
		if v.UserId == u {
			tmp := *v
			rv = append(rv, &tmp)
		}
	}

	return rv, nil
}

func (cdb *CharMemoryDb) GetMainCharacter(u uuid.UUID) (*database.Character, *database.Error) {
	rv, err := cdb.getMainCharacter(u)
	if err != nil {
		return nil, err
	}

	tmp := *rv
	return &tmp, nil
}

func (cdb *CharMemoryDb) GetCharacter(u uuid.UUID, name string) (*database.Character, *database.Error) {
	rv, err := cdb.getCharacter(u, name)
	if err != nil {
		return nil, err
	}

	tmp := *rv
	return &tmp, nil
}

func (cdb *CharMemoryDb) RenameCharacter(u uuid.UUID, old string, name string) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.getCharacter(u, old)
	if err != nil {
		return nil, err
	}

	_, err = cdb.getCharacter(u, name)
	if err == nil {
		return nil, &database.Error{Code: database.CharacterNameTaken, Message: "Character with that name already exists"}
	}

	c.Name = name
	delete(cdb.chars, c.CharId)
	c.CharId = getCharacterId(u, name)
	cdb.chars[c.CharId] = c

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) ChangeMainCharacter(u uuid.UUID, name string) (*database.Character, *database.Error) {
	c, err := cdb.getCharacter(u, name)
	if err != nil {
		return nil, err
	}

	old, err := cdb.getMainCharacter(u)
	if err == nil {
		old.Main = false
	}

	c.Main = true

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) SetCharacterStat(u uuid.UUID, name string, s string, v interface{}) (*database.Character, *database.Error) {
	c, err := cdb.getCharacter(u, name)
	if err != nil {
		return nil, err
	}

	if c.Body == nil {
		c.Body = make(map[string]interface{})
	}
	c.Body[s] = v

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) ChangeCharacterOwner(old uuid.UUID, name string, u uuid.UUID) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.getCharacter(old, name)
	if err != nil {
		return nil, err
	}

	_, err = cdb.getCharacter(u, name)
	if err == nil {
		return nil, &database.Error{Code: database.UserHasCharacter, Message: fmt.Sprintf("Target user already has character with name '%v'", name)}
	}

	c.UserId = u
	delete(cdb.chars, c.CharId)
	c.CharId = getCharacterId(u, name)
	cdb.chars[c.CharId] = c

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) RemoveCharacterStat(u uuid.UUID, name string, s string) (*database.Character, *database.Error) {
	c, err := cdb.getCharacter(u, name)
	if err != nil {
		return nil, err
	}

	if c.Body == nil {
		tmp := *c
		return &tmp, nil
	}

	delete(c.Body, s)
	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) RemoveCharacter(u uuid.UUID, name string) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.getCharacter(u, name)
	if err != nil {
		return nil, nil
	}

	delete(cdb.chars, c.CharId)
	tmp := *c
	return &tmp, nil
}

func getCharacterId(u uuid.UUID, name string) string {
	return fmt.Sprintf("%v:%v", u, name)
}

func (cdb *CharMemoryDb) getCharacter(u uuid.UUID, name string) (*database.Character, *database.Error) {
	if name == "" {
		return cdb.getMainCharacter(u)
	}

	id := getCharacterId(u, name)
	if c, ok := cdb.chars[id]; ok {
		return c, nil
	}

	return nil, &database.Error{Code: database.CharacterNotFound, Message: fmt.Sprintf("Character with name %v was not found", name)}
}

func (cdb *CharMemoryDb) getMainCharacter(u uuid.UUID) (*database.Character, *database.Error) {
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
