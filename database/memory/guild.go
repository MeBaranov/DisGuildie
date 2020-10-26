package memory

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
)

type GuildMemoryDb struct {
	guilds  map[uuid.UUID]*database.Guild
	guildsD map[string]*database.Guild
	mux     sync.Mutex
}

func (gdb *GuildMemoryDb) AddGuild(g *database.Guild) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()
	if g.DiscordId != "" {
		if guild, ok := gdb.guildsD[g.DiscordId]; ok {
			guild.Name = g.Name
			guild.Stats = g.Stats
			return guild, nil
		}
	} else {
		p, ok := gdb.guilds[g.ParentId]
		if !ok {
			return nil, &database.DbError{Code: database.InvalidGuildDefinition, Message: "Invalid parent guild ID"}
		}

		if p.DiscordId != "" {
			g.TopLevelParentId = p.GuildId
		} else {
			g.TopLevelParentId = p.TopLevelParentId
		}

		p, ok = gdb.guilds[g.TopLevelParentId]
		if !ok {
			return nil, &database.DbError{Code: database.InvalidDatabaseState, Message: fmt.Sprintln("Invalid top level guild ID:", g.TopLevelParentId)}
		}

		if _, ok := p.ChildNames[g.Name]; ok {
			return nil, &database.DbError{Code: database.SubguildNameTaken, Message: "Sub-Guild name already taken"}
		}

		p.ChildNames[g.Name] = database.Member
	}

	g.GuildId = uuid.New()
	gdb.guilds[g.GuildId] = g
	if g.DiscordId != "" {
		gdb.guildsD[g.DiscordId] = g
	}

	return g, nil
}

func (gdb *GuildMemoryDb) GetGuild(g uuid.UUID) (*database.Guild, error) {
	if guild, ok := gdb.guilds[g]; ok {
		return guild, nil
	}

	return nil, &database.DbError{Code: database.GuildNotFound, Message: "Guild was not found"}
}

func (gdb *GuildMemoryDb) GetGuildD(d string) (*database.Guild, error) {
	if guild, ok := gdb.guildsD[d]; ok {
		return guild, nil
	}

	return nil, &database.DbError{Code: database.GuildNotFound, Message: "Guild was not found"}
}

func (gdb *GuildMemoryDb) RenameGuild(g uuid.UUID, name string) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.DbError{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	if guild.Name == name {
		return guild, nil
	}

	p, ok := gdb.guilds[guild.TopLevelParentId]
	if !ok {
		return nil, &database.DbError{Code: database.InvalidDatabaseState, Message: fmt.Sprintln("Invalid top level guild ID:", guild.TopLevelParentId)}
	}

	if _, ok := p.ChildNames[name]; ok {
		return nil, &database.DbError{Code: database.SubguildNameTaken, Message: "Sub-Guild name already taken"}
	}

	delete(p.ChildNames, guild.Name)
	guild.Name = name
	p.ChildNames[name] = database.Member

	return guild, nil
}

func (gdb *GuildMemoryDb) AddGuildStat(g uuid.UUID, n string, t string) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.DbError{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	if et, ok := guild.Stats[n]; ok {
		if et == t {
			return guild, nil
		} else {
			return nil, &database.DbError{Code: database.StatNameConflict, Message: "Stat with same name but different type found"}
		}
	}

	guild.Stats[n] = t
	return guild, nil
}

func (gdb *GuildMemoryDb) RemoveGuildStat(g uuid.UUID, n string) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.DbError{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	delete(guild.Stats, n)
	return guild, nil
}

func (gdb *GuildMemoryDb) RemoveGuild(g uuid.UUID) (*database.Guild, error) {
	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, nil
	}

	delete(gdb.guilds, g)
	if guild.DiscordId != "" {
		delete(gdb.guildsD, guild.DiscordId)
	}

	err := gdb.removeGuildsByParent(guild.GuildId)
	if err != nil {
		return guild, err
	}

	return guild, err
}

func (gdb *GuildMemoryDb) RemoveGuildD(d string) (*database.Guild, error) {
	guild, ok := gdb.guildsD[d]
	if !ok {
		return nil, nil
	}

	guild, err := gdb.RemoveGuild(guild.GuildId)
	if err != nil {
		return guild, err
	}

	return guild, nil
}

func (gdb *GuildMemoryDb) removeGuildsByParent(g uuid.UUID) error {
	removeUs := make([]uuid.UUID, 1, 10)

	for _, v := range gdb.guilds {
		if v.ParentId == g {
			removeUs = append(removeUs, v.GuildId)
		}
	}

	for _, r := range removeUs {
		_, err := gdb.RemoveGuild(r)
		if err != nil {
			return err
		}
	}

	return nil
}
