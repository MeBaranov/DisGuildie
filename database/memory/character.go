package memory

import (
	"fmt"
	"sync"

	"github.com/mebaranov/disguildie/database"
)

type CharMemoryDb struct {
	chars map[string]*database.Character
	mux   sync.Mutex
}

func (cdb *CharMemoryDb) AddCharacter(c *database.Character) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	id := getCharacterId(c.GuildId, c.UserId, c.Name)
	if _, ok := cdb.chars[id]; ok {
		return nil, &database.Error{Code: database.CharacterNameTaken, Message: fmt.Sprintf("User already has character with name %v", c.Name)}
	}

	newC := *c
	c = &newC
	cdb.chars[id] = c

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) GetCharacters(g string, u string) ([]*database.Character, *database.Error) {
	rv := make([]*database.Character, 0, 10)
	for _, v := range cdb.chars {
		if v.UserId == u {
			tmp := *v
			rv = append(rv, &tmp)
		}
	}

	return rv, nil
}

func (cdb *CharMemoryDb) GetMainCharacter(g string, u string) (*database.Character, *database.Error) {
	rv, err := cdb.getMainCharacter(g, u)
	if err != nil {
		return nil, err
	}

	tmp := *rv
	return &tmp, nil
}

func (cdb *CharMemoryDb) GetCharacter(g string, u string, name string) (*database.Character, *database.Error) {
	rv, err := cdb.getCharacter(g, u, name)
	if err != nil {
		return nil, err
	}

	tmp := *rv
	return &tmp, nil
}

func (cdb *CharMemoryDb) RenameCharacter(g string, u string, old string, name string) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.getCharacter(g, u, old)
	if err != nil {
		return nil, err
	}

	_, err = cdb.getCharacter(g, u, name)
	if err == nil {
		return nil, &database.Error{Code: database.CharacterNameTaken, Message: "Character with that name already exists"}
	}

	c.Name = name
	idO, idN := getCharacterId(g, u, old), getCharacterId(g, u, name)
	delete(cdb.chars, idO)
	cdb.chars[idN] = c

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) ChangeMainCharacter(g string, u string, name string) (*database.Character, *database.Error) {
	c, err := cdb.getCharacter(g, u, name)
	if err != nil {
		return nil, err
	}

	old, err := cdb.getMainCharacter(g, u)
	if err == nil {
		old.Main = false
	}

	c.Main = true

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) SetCharacterStat(g string, u string, name string, s string, v interface{}) (*database.Character, *database.Error) {
	c, err := cdb.getCharacter(g, u, name)
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

func (cdb *CharMemoryDb) ChangeCharacterOwner(g string, old string, name string, u string) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.getCharacter(g, old, name)
	if err != nil {
		return nil, err
	}

	_, err = cdb.getCharacter(g, u, name)
	if err == nil {
		return nil, &database.Error{Code: database.UserHasCharacter, Message: fmt.Sprintf("Target user already has character with name '%v'", name)}
	}

	c.UserId = u
	ido, idn := getCharacterId(g, old, name), getCharacterId(g, u, name)
	delete(cdb.chars, ido)
	cdb.chars[idn] = c

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) RemoveCharacterStat(g string, u string, name string, s string) (*database.Character, *database.Error) {
	c, err := cdb.getCharacter(g, u, name)
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

func (cdb *CharMemoryDb) RemoveCharacter(g string, u string, name string) (*database.Character, *database.Error) {
	cdb.mux.Lock()
	defer cdb.mux.Unlock()

	c, err := cdb.getCharacter(g, u, name)
	if err != nil {
		return nil, nil
	}

	id := getCharacterId(g, u, name)
	delete(cdb.chars, id)
	tmp := *c
	return &tmp, nil
}

func getCharacterId(g string, u string, name string) string {
	return fmt.Sprintf("%v:%v:%v", g, u, name)
}

func (cdb *CharMemoryDb) getCharacter(g string, u string, name string) (*database.Character, *database.Error) {
	if name == "" {
		return cdb.getMainCharacter(g, u)
	}

	id := getCharacterId(g, u, name)
	if c, ok := cdb.chars[id]; ok {
		return c, nil
	}

	return nil, &database.Error{Code: database.CharacterNotFound, Message: fmt.Sprintf("Character with name %v was not found", name)}
}

func (cdb *CharMemoryDb) getMainCharacter(g string, u string) (*database.Character, *database.Error) {
	var rv *database.Character = nil
	for _, v := range cdb.chars {
		if v.UserId == u && v.GuildId == g {
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
