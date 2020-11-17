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
		if _, ok := gdb.guildsD[g.DiscordId]; ok {
			return nil, &database.Error{Code: database.GuildAlreadyRegistered, Message: fmt.Sprintf("Guild '%v' is already registered", g.DiscordId)}
		}
	} else {
		p, ok := gdb.guilds[g.ParentId]
		if !ok {
			return nil, &database.Error{Code: database.InvalidGuildDefinition, Message: "Invalid parent guild ID"}
		}

		if p.DiscordId != "" {
			g.TopLevelParentId = p.GuildId
		} else {
			g.TopLevelParentId = p.TopLevelParentId
		}

		p, ok = gdb.guilds[g.TopLevelParentId]
		if !ok {
			return nil, &database.Error{Code: database.InvalidDatabaseState, Message: fmt.Sprintln("Invalid top level guild ID:", g.TopLevelParentId)}
		}

		if _, ok := p.ChildNames[g.Name]; ok {
			return nil, &database.Error{Code: database.SubguildNameTaken, Message: fmt.Sprintf("Sub-Guild name '%v' is already taken", g.Name)}
		}

		if p.ChildNames == nil {
			p.ChildNames = make(map[string]database.Void)
		}
		p.ChildNames[g.Name] = database.Member
	}

	newG := *g
	g = &newG
	g.GuildId = uuid.New()
	if g.DiscordId != "" {
		g.TopLevelParentId = g.GuildId
		g.StatVersion = -1
		gdb.guildsD[g.DiscordId] = g
	}
	gdb.guilds[g.GuildId] = g

	tmp := *g
	return &tmp, nil
}

func (gdb *GuildMemoryDb) GetGuild(g uuid.UUID) (*database.Guild, error) {
	if guild, ok := gdb.guilds[g]; ok {
		tmp := *guild
		return &tmp, nil
	}

	return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
}

func (gdb *GuildMemoryDb) GetGuildD(d string) (*database.Guild, error) {
	if guild, ok := gdb.guildsD[d]; ok {
		tmp := *guild
		return &tmp, nil
	}

	return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
}

func (gdb *GuildMemoryDb) GetGuildN(p string, n string) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	parent, ok := gdb.guildsD[p]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Parent guild was not found"}
	}

	for _, g := range gdb.guilds {
		if g.Name == n && g.TopLevelParentId == parent.GuildId {
			tmp := *g
			return &tmp, nil
		}
	}

	return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
}

func (gdb *GuildMemoryDb) GetSubGuilds(g uuid.UUID) (map[uuid.UUID]*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	gld, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	allGuilds := make(map[uuid.UUID]*database.Guild)
	for _, g := range gdb.guilds {
		if g.TopLevelParentId == gld.TopLevelParentId {
			allGuilds[g.GuildId] = g
		}
	}

	subGuilds := map[uuid.UUID]*database.Guild{
		gld.GuildId: gld,
	}
	prevLen, length := -1, 0
	for length > prevLen {
		prevLen = length
		for _, g := range allGuilds {
			if _, ok := subGuilds[g.ParentId]; ok {
				tmp := *g
				subGuilds[g.GuildId] = &tmp
			}
		}
		length = len(subGuilds)
	}

	return subGuilds, nil
}

func (gdb *GuildMemoryDb) RenameGuild(g uuid.UUID, name string) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	if guild.Name == name {
		tmp := *guild
		return &tmp, nil
	}

	p, ok := gdb.guilds[guild.TopLevelParentId]
	if !ok {
		return nil, &database.Error{Code: database.InvalidDatabaseState, Message: fmt.Sprintln("Invalid top level guild ID:", guild.TopLevelParentId)}
	}

	if _, ok := p.ChildNames[name]; ok {
		return nil, &database.Error{Code: database.SubguildNameTaken, Message: fmt.Sprintf("Sub-Guild name '%v' is already taken", name)}
	}

	// This is sanity check. Should never happen. Never ever
	if p.ChildNames == nil {
		p.ChildNames = make(map[string]database.Void)
	}
	delete(p.ChildNames, guild.Name)
	guild.Name = name
	p.ChildNames[name] = database.Member

	tmp := *guild
	return &tmp, nil
}

func (gdb *GuildMemoryDb) MoveGuild(g uuid.UUID, p uuid.UUID) (*database.Guild, error) {
	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	if _, ok := gdb.guilds[p]; !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Parent guild was not found"}
	}

	guild.ParentId = p

	tmp := *guild
	return &tmp, nil
}

func (gdb *GuildMemoryDb) RemoveGuild(g uuid.UUID) (*database.Guild, error) {
	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	err := gdb.removeGuildsByParent(guild.GuildId)
	if err != nil {
		return nil, err
	}

	delete(gdb.guilds, g)
	if guild.DiscordId != "" {
		delete(gdb.guildsD, guild.DiscordId)
	} else if parent, ok := gdb.guilds[guild.TopLevelParentId]; ok {
		delete(parent.ChildNames, guild.Name)
	}

	tmp := *guild
	return &tmp, nil
}

func (gdb *GuildMemoryDb) RemoveGuildD(d string) (*database.Guild, error) {
	guild, ok := gdb.guildsD[d]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}

	guild, err := gdb.RemoveGuild(guild.GuildId)
	if err != nil {
		return nil, err
	}

	tmp := *guild
	return &tmp, nil
}

func (gdb *GuildMemoryDb) AddGuildStat(g uuid.UUID, s *database.Stat) (*database.Guild, error) {
	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}
	if guild.DiscordId == "" {
		return nil, &database.Error{Code: database.GuildLevelError, Message: "Only top-level guild stats are supported right now"}
	}

	if et, ok := guild.Stats[s.ID]; ok {
		if et.Type == s.Type {
			et.Description = s.Description
			tmp := *guild
			return &tmp, nil
		} else {
			return nil, &database.Error{Code: database.StatNameConflict, Message: fmt.Sprintf("Stat with same name (%v) but different type (%v) found", s.ID, et.Type)}
		}
	}

	if guild.Stats == nil {
		guild.Stats = make(map[string]*database.Stat)
		guild.DefaultStat = s.ID
	}
	tmpStat := *s
	guild.Stats[s.ID] = &tmpStat
	guild.StatVersion += 1
	tmp := *guild
	return &tmp, nil
}

func (gdb *GuildMemoryDb) SetDefaultGuildStat(g uuid.UUID, sn string) (*database.Guild, error) {
	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}
	if guild.DiscordId == "" {
		return nil, &database.Error{Code: database.GuildLevelError, Message: "Only top-level guild stats are supported right now"}
	}

	if _, ok := guild.Stats[sn]; !ok {
		return nil, &database.Error{Code: database.StatNotFound, Message: "Could not find stat with name " + sn}
	}

	guild.DefaultStat = sn
	tmp := *guild
	return &tmp, nil
}

func (gdb *GuildMemoryDb) RemoveGuildStat(g uuid.UUID, n string) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}
	if guild.DiscordId == "" {
		return nil, &database.Error{Code: database.GuildLevelError, Message: "Only top-level guild stats are supported right now"}
	}

	if _, ok := guild.Stats[n]; !ok {
		return nil, &database.Error{Code: database.StatNotFound, Message: "Stat was not found"}
	}

	delete(guild.Stats, n)
	if n == guild.DefaultStat {
		if len(guild.Stats) == 0 {
			guild.DefaultStat = ""
		} else {
			// Oh my, what a hack
			for n, _ := range guild.Stats {
				guild.DefaultStat = n
				break
			}
		}
	}
	guild.StatVersion += 1
	tmp := *guild
	return &tmp, nil
}

func (gdb *GuildMemoryDb) RemoveAllGuildStats(g uuid.UUID) (*database.Guild, error) {
	gdb.mux.Lock()
	defer gdb.mux.Unlock()

	guild, ok := gdb.guilds[g]
	if !ok {
		return nil, &database.Error{Code: database.GuildNotFound, Message: "Guild was not found"}
	}
	if guild.DiscordId == "" {
		return nil, &database.Error{Code: database.GuildLevelError, Message: "Only top-level guild stats are supported right now"}
	}

	guild.Stats = nil
	guild.StatVersion += 1
	tmp := *guild
	return &tmp, nil
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
			dbErr := database.ErrToDbErr(err)
			if dbErr == nil || dbErr.Code != database.GuildNotFound {
				return err
			}
		}
	}

	return nil
}
