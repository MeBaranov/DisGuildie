package admin_tests

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mebaranov/disguildie/database/memory"
	"github.com/mebaranov/disguildie/processor/helpers/admin"
	"github.com/mebaranov/disguildie/processor/helpers/tests"

	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
)

type guildTest struct {
	Name         string
	Command      string
	ErrStr       string
	Result       string
	Validation   func(prefix string, t *testing.T)
	Preparations func()
	GuildsD      map[string]string
	Guilds       map[string]string
}

func TestAdd(t *testing.T) {
	msg := &tests.TestMessage{}
	prov := memory.NewMemoryDb()
	mainGld := &database.Guild{
		DiscordId: uuid.New().String(),
		Name:      "test",
	}
	mainGld, _ = prov.AddGuild(mainGld)
	msg.GuildIdMock = func() string { return uuid.New().String() }
	msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, errors.New("Test error") }

	testActions := []guildTest{
		{
			Name:    "add: empty",
			Command: "a",
			ErrStr:  "Invalid command format",
		},
		{
			Name:    "add: a 1 parameter",
			Command: "add newsubgld",
			ErrStr:  "Invalid command format",
		},
		{
			Name:    "add: main guild not found",
			Command: "a newsubgld main",
			ErrStr:  "Guild was not found",
			Result:  "getting parent guild",
		},
		{
			Name:         "add: sub-guild not found",
			Preparations: func() { msg.GuildIdMock = func() string { return mainGld.DiscordId } },
			Command:      "a newsubgld newsubgld2",
			ErrStr:       "Guild was not found",
			Result:       "getting parent guild",
		},
		{
			Name:    "add: error for permissions",
			Command: "add newsubgld main",
			ErrStr:  "Test error",
			Result:  "checking modification permissions",
		},
		{
			Name: "add: no permissions",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, nil }
			},
			Command: "a newsubgld main",
			ErrStr:  "You don't have permissions to modify the sub-guild",
			Result:  "",
		},
		{
			Name: "add: success main",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return true, nil }
			},
			Command: "add newsubgld main",
			ErrStr:  "",
			Result:  "Sub-guild newsubgld registered under main.",
			Guilds:  map[string]string{"newsubgld": "test"},
		},
		{
			Name:    "add: duplicate name",
			Command: "a newsubgld main",
			ErrStr:  "Sub-Guild name 'newsubgld' is already taken",
			Result:  "adding guild",
			Guilds:  map[string]string{"newsubgld": "test"},
		},
		{
			Name:    "add: success sub",
			Command: "a newsubgld2 newsubgld",
			ErrStr:  "",
			Result:  "Sub-guild newsubgld2 registered under newsubgld.",
			Guilds:  map[string]string{"newsubgld": "test", "newsubgld2": "newsubgld"},
		},
	}

	runTest(t, testActions, msg, prov)
}

func TestRename(t *testing.T) {
	msg := &tests.TestMessage{}
	prov := memory.NewMemoryDb()
	mainGld, _ := prov.AddGuild(&database.Guild{
		DiscordId: uuid.New().String(),
		Name:      "test",
	})
	prov.AddGuild(&database.Guild{
		Name:     "renameMe",
		ParentId: mainGld.GuildId,
	})
	prov.AddGuild(&database.Guild{
		Name:     "renameMe2",
		ParentId: mainGld.GuildId,
	})
	msg.GuildIdMock = func() string { return uuid.New().String() }
	msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, errors.New("Test error") }
	msg.GuildIdMock = func() string { return mainGld.DiscordId }

	testActions := []guildTest{
		{
			Name:    "rename: empty",
			Command: "r",
			ErrStr:  "Invalid command format",
		},
		{
			Name:    "rename: a 1 parameter",
			Command: "rename renameMe",
			ErrStr:  "Invalid command format",
		},
		{
			Name:    "rename: source not found",
			Command: "r unknown tmp",
			ErrStr:  "Guild was not found",
			Result:  "getting source guild",
		},
		{
			Name:    "rename: error for permissions",
			Command: "rename renameMe targetName",
			ErrStr:  "Test error",
			Result:  "checking modification permissions",
		},
		{
			Name: "rename: no permissions",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, nil }
			},
			Command: "r renameMe targetName",
			ErrStr:  "You don't have permissions to modify the sub-guild",
			Result:  "",
		},
		{
			Name: "rename: success main",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return true, nil }
			},
			Command: "rename renameMe targetName",
			ErrStr:  "",
			Result:  "Sub-guild 'renameMe' renamed to 'targetName'",
			Guilds:  map[string]string{"targetName": "test", "renameMe2": "test"},
		},
		{
			Name:    "rename: duplicate name",
			Command: "r renameMe2 targetName",
			ErrStr:  "Sub-Guild name 'targetName' is already taken",
			Result:  "renaming guild",
			Guilds:  map[string]string{"targetName": "test", "renameMe2": "test"},
		},
	}

	runTest(t, testActions, msg, prov)
}

func TestMove(t *testing.T) {
	msg := &tests.TestMessage{}
	prov := memory.NewMemoryDb()
	mainGld, _ := prov.AddGuild(&database.Guild{
		DiscordId: uuid.New().String(),
		Name:      "test",
	})
	moveMe, _ := prov.AddGuild(&database.Guild{
		Name:     "moveMe",
		ParentId: mainGld.GuildId,
	})
	prov.AddGuild(&database.Guild{
		Name:     "target",
		ParentId: mainGld.GuildId,
	})
	msg.GuildIdMock = func() string { return uuid.New().String() }
	msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, errors.New("Test error") }
	msg.GuildIdMock = func() string { return mainGld.DiscordId }

	testActions := []guildTest{
		{
			Name:    "move: empty",
			Command: "m",
			ErrStr:  "Invalid command format",
		},
		{
			Name:    "move: a 1 parameter",
			Command: "move moveMe",
			ErrStr:  "Invalid command format",
		},
		{
			Name:    "move: source not found",
			Command: "m unknown target",
			ErrStr:  "Guild was not found",
			Result:  "getting source guild",
		},
		{
			Name:    "move: target not found",
			Command: "m moveMe unknown",
			ErrStr:  "Guild was not found",
			Result:  "getting target guild",
		},
		{
			Name:    "move: error for permissions",
			Command: "move moveMe target",
			ErrStr:  "Test error",
			Result:  "checking source modification permissions",
		},
		{
			Name: "move: no permissions source",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, nil }
			},
			Command: "m moveMe target",
			ErrStr:  "You don't have permissions to modify the source (moveMe) sub-guild",
			Result:  "",
		},
		{
			Name: "move: no permissions target",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(u uuid.UUID) (bool, error) { return u == moveMe.GuildId, nil }
			},
			Command: "m moveMe target",
			ErrStr:  "You don't have permissions to modify the target (moveMe) sub-guild",
			Result:  "",
		},
		{
			Name: "move: success target",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return true, nil }
			},
			Command: "move moveMe target",
			ErrStr:  "",
			Result:  "Sub-guild 'moveMe' moved under 'target'",
			Guilds:  map[string]string{"target": "test", "moveMe": "target"},
		},
		{
			Name:    "move: success main",
			Command: "move moveMe main",
			ErrStr:  "",
			Result:  "Sub-guild 'moveMe' moved under 'main'",
		},
	}

	runTest(t, testActions, msg, prov)
}

func TestRemove(t *testing.T) {
	msg := &tests.TestMessage{}
	prov := memory.NewMemoryDb()
	mainGld, _ := prov.AddGuild(&database.Guild{
		DiscordId: uuid.New().String(),
		Name:      "test",
	})
	removeMe, _ := prov.AddGuild(&database.Guild{
		Name:     "removeMe",
		ParentId: mainGld.GuildId,
	})
	sub, _ := prov.AddGuild(&database.Guild{
		Name:     "sub",
		ParentId: removeMe.GuildId,
	})
	prov.AddUser("u1", &database.GuildPermission{GuildId: sub.GuildId, TopGuild: mainGld.DiscordId})
	prov.AddUser("u2", &database.GuildPermission{GuildId: removeMe.GuildId, TopGuild: mainGld.DiscordId})
	prov.AddUser("u3", &database.GuildPermission{GuildId: mainGld.GuildId, TopGuild: mainGld.DiscordId})
	msg.GuildIdMock = func() string { return uuid.New().String() }
	msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, errors.New("Test error") }
	msg.GuildIdMock = func() string { return mainGld.DiscordId }

	testActions := []guildTest{
		{
			Name:    "remove: empty",
			Command: "remove",
			ErrStr:  "Invalid command format",
		},
		{
			Name:    "remove: source not found",
			Command: "remove unknown",
			ErrStr:  "Guild was not found",
			Result:  "getting guild",
		},
		{
			Name:    "remove: error for permissions",
			Command: "remove removeMe",
			ErrStr:  "Test error",
			Result:  "checking modification permissions",
		},
		{
			Name: "remove: no permissions source",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return false, nil }
			},
			Command: "remove removeMe",
			ErrStr:  "You don't have permissions to modify the sub-guild",
			Result:  "",
		},
		{
			Name: "remove: success target",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return true, nil }
			},
			Command: "remove removeMe",
			ErrStr:  "",
			Result:  "Sub-guild 'removeMe' removed",
			Guilds:  map[string]string{},
			Validation: func(prefix string, t *testing.T) {
				for _, u := range prov.UsersD {
					if len(u.Guilds) != 1 {
						t.Errorf("[%v] Something has gone wrong with guilds amount. Got: %v, Wish: 1 guild", prefix, u)
					}
					tmp, ok := u.Guilds[mainGld.DiscordId]
					if !ok {
						t.Errorf("[%v] User %v is not a part of the guild at all anymore", prefix, u)
					}
					if tmp.GuildId != mainGld.GuildId {
						t.Errorf("[%v] Wrong guild membership. Got: %v, wish: %v", prefix, u, mainGld.GuildId)
					}
				}
			},
		},
	}

	runTest(t, testActions, msg, prov)
}

func runTest(t *testing.T, testActions []guildTest, msg *tests.TestMessage, prov *memory.MemoryDB) {

	defaultGuildsD := make(map[string]string)
	for _, v := range prov.GuildsD {
		defaultGuildsD[v.Name] = v.DiscordId
	}
	defaultGuilds := make(map[string]string)
	for _, v := range prov.Guilds {
		if v.DiscordId == "" {
			defaultGuilds[v.Name] = prov.Guilds[v.ParentId].Name
		}
	}
	target := admin.NewAdminGuildProcessor(prov)

	for _, cur := range testActions {
		msg.CurMsg = cur.Command
		if cur.Preparations != nil {
			cur.Preparations()
		}

		rv, err := target.ProcessMessage(msg)
		if cur.Result != rv {
			t.Errorf("[%v] Wrong processing result. Got: %v, Wish: %v", cur.Name, rv, cur.Result)
		}
		if (cur.ErrStr != "" && err == nil) || (err != nil && cur.ErrStr != err.Error()) {
			t.Errorf("[%v] Wrong processing error. Got: %v, Wish: %v", cur.Name, err, cur.ErrStr)
		}

		gldsD := make(map[string]string)
		for _, g := range prov.GuildsD {
			gldsD[g.Name] = g.DiscordId
		}
		glds := make(map[string]string)
		for _, g := range prov.Guilds {
			if g.DiscordId == "" {
				glds[g.Name] = prov.Guilds[g.ParentId].Name
			}
		}

		if cur.GuildsD == nil {
			cur.GuildsD = defaultGuildsD
		}
		if cur.Guilds == nil {
			cur.Guilds = defaultGuilds
		}

		if len(cur.GuildsD) != len(gldsD) || (len(gldsD) > 0 && !reflect.DeepEqual(cur.GuildsD, gldsD)) {
			t.Errorf("[%v] Unexpected guildsD. Actual: %v. Expected: %v", cur.Name, gldsD, cur.GuildsD)
		}
		if len(cur.Guilds) != len(glds) || (len(glds) > 0 && !reflect.DeepEqual(cur.Guilds, glds)) {
			t.Errorf("[%v] Unexpected guilds. Actual: %v. Expected: %v", cur.Name, glds, cur.Guilds)
		}

		if cur.Validation != nil {
			cur.Validation(cur.Name, t)
		}
	}
}
