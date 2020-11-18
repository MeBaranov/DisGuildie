package memory

import (
	"fmt"
	"sort"
	"sync"

	"github.com/mebaranov/disguildie/database"
)

type CharMemoryDb struct {
	chars map[string]*database.Character
	mux   sync.Mutex
}

func (cdb *CharMemoryDb) AddCharacter(c *database.Character) (*database.Character, error) {
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

func (cdb *CharMemoryDb) GetCharacters(g string, u string) ([]*database.Character, error) {
	rv := make([]*database.Character, 0, 10)
	for _, v := range cdb.chars {
		if v.UserId == u && v.GuildId == g {
			tmp := *v
			rv = append(rv, &tmp)
		}
	}

	return rv, nil
}

func (cdb *CharMemoryDb) GetCharactersSorted(g string, s string, t int, asc bool, limit int) ([]*database.Character, error) {
	rv := make([]*database.Character, 0, 600)
	for _, v := range cdb.chars {
		if v.GuildId == g {
			tmp := *v
			rv = append(rv, &tmp)
		}
	}

	var f func(i int, j int) bool
	switch t {
	case database.Number:
		f = func(i int, j int) bool {
			av, ok := rv[i].Body[s]
			if !ok {
				return false
			}
			bv, ok := rv[j].Body[s]
			if !ok {
				return true
			}

			ai, ok := av.(int)
			if !ok {
				return false
			}

			bi, ok := bv.(int)
			if !ok {
				return true
			}

			return (asc && ai < bi) || (!asc && ai > bi)
		}
	case database.Str:
		f = func(i int, j int) bool {
			av, ok := rv[i].Body[s]
			if !ok {
				return false
			}
			bv, ok := rv[j].Body[s]
			if !ok {
				return true
			}

			ai, ok := av.(string)
			if !ok {
				return false
			}

			bi, ok := bv.(string)
			if !ok {
				return true
			}

			return (asc && ai < bi) || (!asc && ai > bi)
		}
	default:
		return nil, &database.Error{Code: database.UnknownStatType, Message: "Stat type for " + s + " is not defined"}
	}

	sort.Slice(rv, f)
	if limit > 0 {
		rv = rv[:limit]
	}
	return rv, nil
}

func (cdb *CharMemoryDb) GetCharactersOutdated(g string, v int) ([]*database.Character, error) {
	rv := make([]*database.Character, 0, 600)
	for _, c := range cdb.chars {
		if c.GuildId == g && c.StatVersion < v {
			tmp := *c
			rv = append(rv, &tmp)
		}
	}

	return rv, nil
}

func (cdb *CharMemoryDb) GetCharactersByName(g string, n string) ([]*database.Character, error) {
	rv := make([]*database.Character, 0, 600)
	for _, c := range cdb.chars {
		if c.GuildId == g && c.Name == n {
			tmp := *c
			rv = append(rv, &tmp)
		}
	}

	return rv, nil
}

func (cdb *CharMemoryDb) GetMainCharacter(g string, u string) (*database.Character, error) {
	rv, err := cdb.getMainCharacter(g, u)
	if err != nil {
		return nil, err
	}

	tmp := *rv
	return &tmp, nil
}

func (cdb *CharMemoryDb) GetCharacter(g string, u string, name string) (*database.Character, error) {
	rv, err := cdb.getCharacter(g, u, name)
	if err != nil {
		return nil, err
	}

	tmp := *rv
	return &tmp, nil
}

func (cdb *CharMemoryDb) RenameCharacter(g string, u string, old string, name string) (*database.Character, error) {
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

func (cdb *CharMemoryDb) ChangeMainCharacter(g string, u string, name string) (*database.Character, error) {
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

func (cdb *CharMemoryDb) SetCharacterStat(g string, u string, name string, s string, v interface{}) (*database.Character, error) {
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

func (cdb *CharMemoryDb) SetCharacterStatVersion(g string, u string, name string, stats map[string]*database.Stat, version int) (*database.Character, error) {
	c, err := cdb.getCharacter(g, u, name)
	if err != nil {
		return nil, err
	}
	if c.StatVersion >= version {
		return c, nil
	}

	if c.Body == nil {
		c.Body = make(map[string]interface{})
	} else {
		rm := make([]string, 0, len(c.Body))
		for k, v := range c.Body {
			s, ok := stats[k]
			if !ok {
				rm = append(rm, k)
				continue
			}

			switch s.Type {
			case database.Number:
				_, ok = v.(int)
			case database.Str:
				_, ok = v.(string)
			default:
				return nil, &database.Error{Code: database.UnknownStatType, Message: fmt.Sprintf("Stat type for %v is not defined", s.ID)}
			}
			if !ok {
				rm = append(rm, k)
				continue
			}
		}

		for _, s := range rm {
			delete(c.Body, s)
		}
	}

	for _, s := range stats {
		if _, ok := c.Body[s.ID]; !ok {
			switch s.Type {
			case database.Number:
				c.Body[s.ID] = 0
			case database.Str:
				c.Body[s.ID] = ""
			default:
				return nil, &database.Error{Code: database.UnknownStatType, Message: fmt.Sprintf("Stat type for %v is not defined", s.ID)}
			}
		}
	}
	c.StatVersion = version

	tmp := *c
	return &tmp, nil
}

func (cdb *CharMemoryDb) ChangeCharacterOwner(g string, old string, name string, u string) (*database.Character, error) {
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

func (cdb *CharMemoryDb) RemoveCharacterStat(g string, u string, name string, s string) (*database.Character, error) {
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

func (cdb *CharMemoryDb) RemoveCharacter(g string, u string, name string) (*database.Character, error) {
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

func (cdb *CharMemoryDb) getCharacter(g string, u string, name string) (*database.Character, error) {
	if name == "" {
		return cdb.getMainCharacter(g, u)
	}

	id := getCharacterId(g, u, name)
	if c, ok := cdb.chars[id]; ok {
		return c, nil
	}

	return nil, &database.Error{Code: database.CharacterNotFound, Message: fmt.Sprintf("Character with name %v was not found", name)}
}

func (cdb *CharMemoryDb) getMainCharacter(g string, u string) (*database.Character, error) {
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
