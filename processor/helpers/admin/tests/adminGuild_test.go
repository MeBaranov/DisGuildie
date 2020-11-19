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
	defaultGuildsD := map[string]string{"test": mainGld.DiscordId}

	target := admin.NewAdminGuildProcessor(prov)

	testActions := []struct {
		Name         string
		Command      string
		ErrStr       string
		Result       string
		Validation   func(prefix string, t *testing.T)
		Preparations func()
		GuildsD      map[string]string
		Guilds       map[string]string
	}{
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
			GuildsD: defaultGuildsD,
		},
		{
			Name: "add: success main",
			Preparations: func() {
				msg.CheckGuildModificationPermissionsMock = func(uuid.UUID) (bool, error) { return true, nil }
			},
			Command: "add newsubgld main",
			ErrStr:  "",
			Result:  "Subguild newsubgld registered under main.",
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
			Result:  "Subguild newsubgld2 registered under newsubgld.",
			Guilds:  map[string]string{"newsubgld": "test", "newsubgld2": "newsubgld"},
		},
	}

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
