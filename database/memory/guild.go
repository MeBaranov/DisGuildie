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
		p := g
		for ; p != nil && p.ParentId != uuid.Nil && p.DiscordId == ""; p, _ = gdb.guilds[p.ParentId] {
		}

		if p.ParentId != uuid.Nil && p.DiscordId != "" {
			return nil, &database.DbError{Code: database.InvalidGuildDBState, Message: "Sub-Guild contains DiscordID"}
		}
		if p == g {
			return nil, &database.DbError{Code: database.InvalidGuildDefinition, Message: "Sub-Guild does not contain valid parent"}
		}
		if p.DiscordId == "" {
			return nil, &database.DbError{Code: database.InvalidGuildDBState, Message: fmt.Sprintln("Top-level guild", p.GuildId, "does not have discordID")}
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

func (gdb *GuildMemoryDb) GetGuild(d string) *database.Guild {
	panic("TODO")
}

func (gdb *GuildMemoryDb) GetGuildD(g uuid.UUID) *database.Guild {
	panic("TODO")
}

func (gdb *GuildMemoryDb) RenameGuild(g uuid.UUID, name string) *database.Guild {
	panic("TODO")
}

func (gdb *GuildMemoryDb) RestatGuild(g uuid.UUID, name string) *database.Guild {
	panic("TODO")
}

func (gdb *GuildMemoryDb) RemoveGuild(g uuid.UUID) *database.Guild {
	panic("TODO")
}

func (gdb *GuildMemoryDb) RemoveGuildD(d string) *database.Guild {
	panic("TODO")
}
