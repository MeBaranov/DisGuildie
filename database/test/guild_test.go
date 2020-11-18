package database_test

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/mebaranov/disguildie/database"
)

func TestGuildAdd(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did1",
		}

		rc1, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc1.Name != g.Name || rc1.DiscordId != g.DiscordId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc1, g)
		}

		g = &database.Guild{
			Name:      "test2",
			DiscordId: "did2",
		}
		rc2, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc2.Name != g.Name || rc2.DiscordId != g.DiscordId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc2, g)
		}
		if rc2 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc3, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc3.Name != g.Name || rc3.ParentId != rc1.GuildId || rc3.TopLevelParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc3, g)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc3.GuildId,
		}
		rc4, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc4.Name != g.Name || rc4.ParentId != rc3.GuildId || rc4.TopLevelParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc4, g)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc5, err := d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc5.Name != g.Name || rc5.ParentId != rc2.GuildId || rc5.TopLevelParentId != rc2.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, rc4, g)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc4.GuildId,
		}
		rce, err := d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Sub-Guild name 'sub2-test1' is already taken", database.SubguildNameTaken, n); e != "" {
			t.Fatalf(e)
		}

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc4.GuildId,
		}
		rce, err = d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Sub-Guild name 'sub1-test1' is already taken", database.SubguildNameTaken, n); e != "" {
			t.Fatalf(e)
		}

		g = &database.Guild{
			Name:     "sub10-test10",
			ParentId: uuid.New(),
		}
		rce, err = d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Invalid parent guild ID", database.InvalidGuildDefinition, n); e != "" {
			t.Fatalf(e)
		}

		g = &database.Guild{
			Name:      "sub10-test10",
			DiscordId: "did1",
		}
		rce, err = d.AddGuild(g)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild 'did1' is already registered", database.GuildAlreadyRegistered, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestGuildGet(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test22",
			DiscordId: "did122",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc3, err := d.AddGuild(g)

		g, err = d.GetGuild(rc3.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc3.Name != g.Name || g.ParentId != rc2.GuildId || g.TopLevelParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, g, rc3)
		}
		if rc3 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		names := make(map[string]database.Void)
		names["sub1-test1"] = database.Member
		names["sub2-test1"] = database.Member
		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc1.Name != g.Name || rc1.DiscordId != g.DiscordId || !reflect.DeepEqual(g.ChildNames, names) {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v, with child names: %v", n, g, rc1, names)
		}

		rce, err := d.GetGuild(uuid.New())
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestGuildGetD(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test33",
			DiscordId: "did133",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		d.AddGuild(g)

		names := make(map[string]database.Void)
		names["sub1-test1"] = database.Member
		names["sub2-test1"] = database.Member
		g, err = d.GetGuildD("did133")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc1.Name != g.Name || rc1.DiscordId != g.DiscordId || !reflect.DeepEqual(g.ChildNames, names) {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v, with child names: %v", n, g, rc1, names)
		}
		if rc1 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rce, err := d.GetGuildD("did233")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestGuildGetN(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test_getn",
			DiscordId: "did_getn",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test_getn",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test_getn",
			ParentId: rc2.GuildId,
		}
		rc3, err := d.AddGuild(g)

		g, err = d.GetGuildN("did_getn_nonexistent", "sub2-test_getn")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Parent guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.GetGuildN("did_getn", "sub2-test_getn_nonexistent")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.GetGuildN("did_getn", "sub2-test_getn")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc3.Name != g.Name || rc3.GuildId != g.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v.", n, g, rc3)
		}
		if rc3 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		g, err = d.GetGuildN("did_getn", "sub1-test_getn")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc2.Name != g.Name || rc2.GuildId != g.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v.", n, g, rc2)
		}
	}
}

func TestGuildGetSub(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test_getsub",
			DiscordId: "did_getsub",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test_getsub",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test_getsub",
			ParentId: rc2.GuildId,
		}
		rc3, err := d.AddGuild(g)

		gs, err := d.GetSubGuilds(uuid.New())
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, gs)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		gs, err = d.GetSubGuilds(rc3.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if len(gs) != 1 || rc3.Name != gs[rc3.GuildId].Name || rc3.GuildId != gs[rc3.GuildId].GuildId {
			t.Fatalf("[%v] Wrong guilds returned. Actual: %v, expected: %v.", n, gs, rc3)
		}
		if rc3 == gs[rc3.GuildId] {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		gs, err = d.GetSubGuilds(rc2.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if len(gs) != 2 || rc3.Name != gs[rc3.GuildId].Name || rc3.GuildId != gs[rc3.GuildId].GuildId ||
			rc2.Name != gs[rc2.GuildId].Name || rc2.GuildId != gs[rc2.GuildId].GuildId {

			t.Fatalf("[%v] Wrong guilds returned. Actual: %v, expected: %v elements.", n, gs, 2)
		}

		gs, err = d.GetSubGuilds(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if len(gs) != 3 || rc3.Name != gs[rc3.GuildId].Name || rc3.GuildId != gs[rc3.GuildId].GuildId ||
			rc2.Name != gs[rc2.GuildId].Name || rc2.GuildId != gs[rc2.GuildId].GuildId ||
			rc1.Name != gs[rc1.GuildId].Name || rc1.GuildId != gs[rc1.GuildId].GuildId {

			t.Fatalf("[%v] Wrong guilds returned. Actual: %v, expected: %v elements.", n, gs, 3)
		}
	}
}

func TestGuildRename(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did144",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		d.AddGuild(g)

		names := make(map[string]database.Void)
		names["sub1-test2"] = database.Member
		names["sub2-test1"] = database.Member

		rc, err := d.RenameGuild(rc2.GuildId, "sub1-test2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != "sub1-test2" || rc.GuildId != rc2.GuildId || rc.ParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: name: %v, parentId: %v", n, rc, "sub1-test2", rc1.GuildId)
		}
		if rc == rc2 {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		rce, err := d.RenameGuild(uuid.New(), "test3")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		rce, err = d.RenameGuild(rc2.GuildId, "sub2-test1")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Sub-Guild name 'sub2-test1' is already taken", database.SubguildNameTaken, n); e != "" {
			t.Fatalf(e)
		}

		rc, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.Name != "test" || rc.DiscordId != "did144" || !reflect.DeepEqual(rc.ChildNames, names) {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v, with child names: %v", n, g, rc1, names)
		}
	}
}

func TestGuildAddStat(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did155",
		}
		rc1, _ := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, _ := d.AddGuild(g)

		s1 := &database.Stat{
			ID:          "s1",
			Type:        database.Number,
			Description: "desc1",
		}
		g, err := d.AddGuildStat(rc2.GuildId, s1)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Only top-level guild stats are supported right now", database.GuildLevelError, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.AddGuildStat(uuid.New(), s1)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		stats := map[string]*database.Stat{"s1": s1}

		g, err = d.AddGuildStat(rc1.GuildId, s1)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}
		if rc1 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}
		if g.StatVersion != 1 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 1)
		}

		g, err = d.AddGuildStat(rc1.GuildId, s1)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}

		s1.Type = database.Str
		g, err = d.AddGuildStat(rc1.GuildId, s1)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Stat with same name (s1) but different type (1) found", database.StatNameConflict, n); e != "" {
			t.Fatalf(e)
		}

		s1.Type = database.Number
		s2 := &database.Stat{
			ID:          "s2",
			Type:        database.Str,
			Description: "desc2",
		}
		stats["s2"] = s2
		g, err = d.AddGuildStat(rc1.GuildId, s2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}
		if g.StatVersion != 2 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 2)
		}

		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}
	}
}

func TestGuildSetDefaultStat(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: uuid.New().String(),
		}
		rc1, _ := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, _ := d.AddGuild(g)

		g, err := d.SetDefaultGuildStat(rc2.GuildId, "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Only top-level guild stats are supported right now", database.GuildLevelError, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.SetDefaultGuildStat(uuid.New(), "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.SetDefaultGuildStat(rc1.GuildId, "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Stat was not found", database.StatNotFound, n); e != "" {
			t.Fatalf(e)
		}

		s1 := &database.Stat{
			ID:          "s1",
			Type:        database.Str,
			Description: "desc1",
		}
		d.AddGuildStat(rc1.GuildId, s1)
		g, err = d.SetDefaultGuildStat(rc1.GuildId, "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Stat was not found", database.StatNotFound, n); e != "" {
			t.Fatalf(e)
		}

		s2 := &database.Stat{
			ID:          "s2",
			Type:        database.Str,
			Description: "desc2",
		}
		d.AddGuildStat(rc1.GuildId, s2)
		g, err = d.SetDefaultGuildStat(rc1.GuildId, "s2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.DefaultStat != "s2" {
			t.Fatalf("[%v] Wrong default stat. Actual: %v, expected: %v", n, g.DefaultStat, "s2")
		}
		if rc1 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}
		if g.StatVersion != 2 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 2)
		}

		g, err = d.SetDefaultGuildStat(rc1.GuildId, "s1")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.DefaultStat != "s1" {
			t.Fatalf("[%v] Wrong default stat. Actual: %v, expected: %v", n, g.DefaultStat, "s1")
		}
		if g.StatVersion != 2 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 2)
		}

		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.DefaultStat != "s1" {
			t.Fatalf("[%v] Wrong default stat. Actual: %v, expected: %v", n, g.DefaultStat, "s1")
		}
		if g.StatVersion != 2 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 2)
		}
	}
}

func TestGuildRemoveStat(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: uuid.New().String(),
		}
		rc1, _ := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, _ := d.AddGuild(g)

		g, err := d.RemoveGuildStat(rc2.GuildId, "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Only top-level guild stats are supported right now", database.GuildLevelError, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.RemoveGuildStat(uuid.New(), "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.RemoveGuildStat(rc1.GuildId, "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Stat was not found", database.StatNotFound, n); e != "" {
			t.Fatalf(e)
		}

		s1 := &database.Stat{
			ID:          "s1",
			Type:        database.Str,
			Description: "desc1",
		}
		d.AddGuildStat(rc1.GuildId, s1)
		g, err = d.RemoveGuildStat(rc1.GuildId, "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Stat was not found", database.StatNotFound, n); e != "" {
			t.Fatalf(e)
		}

		s2 := &database.Stat{
			ID:          "s2",
			Type:        database.Str,
			Description: "desc2",
		}
		stats := map[string]*database.Stat{"s1": s1}
		d.AddGuildStat(rc1.GuildId, s2)
		g, err = d.RemoveGuildStat(rc1.GuildId, "s2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}
		if rc1 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}
		if g.StatVersion != 3 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 3)
		}

		delete(stats, "s1")
		g, err = d.RemoveGuildStat(rc1.GuildId, "s1")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.Stats == nil || !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}
		if g.StatVersion != 4 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 4)
		}

		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.Stats == nil || !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}
		if g.StatVersion != 4 {
			t.Fatalf("[%v] Wrong stats version. Actual: %v, expected: %v", n, g.StatVersion, 4)
		}
	}
}

func TestGuildRemoveAllStats(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: uuid.New().String(),
		}
		rc1, f := d.AddGuild(g)
		_ = f
		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, _ := d.AddGuild(g)

		g, err := d.RemoveAllGuildStats(rc2.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Only top-level guild stats are supported right now", database.GuildLevelError, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.RemoveAllGuildStats(uuid.New())
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		s1 := &database.Stat{
			ID:          "s1",
			Type:        database.Str,
			Description: "desc1",
		}
		d.AddGuildStat(rc1.GuildId, s1)

		s2 := &database.Stat{
			ID:          "s2",
			Type:        database.Str,
			Description: "desc2",
		}
		d.AddGuildStat(rc1.GuildId, s2)

		g, err = d.RemoveAllGuildStats(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.Stats != nil && len(g.Stats) > 0 {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, "empty")
		}

		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.Stats != nil && len(g.Stats) > 0 {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, "empty")
		}
	}
}

func TestGuildMove(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test_move",
			DiscordId: "did_move",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test_move",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test_move",
			ParentId: rc2.GuildId,
		}
		rc3, err := d.AddGuild(g)

		g, err = d.MoveGuild(uuid.New(), rc1.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.MoveGuild(rc3.GuildId, uuid.New())
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Parent guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.MoveGuild(rc3.GuildId, rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.GuildId != rc3.GuildId || g.ParentId != rc1.GuildId || g.TopLevelParentId != rc1.GuildId {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v, with parent: %v", n, g, rc3, rc1.GuildId)
		}
		if rc3 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}
	}
}

func TestGuildRemove(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did177",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc3, err := d.AddGuild(g)

		g, err = d.RemoveGuild(uuid.New())
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.RemoveGuild(rc3.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.GuildId != rc3.GuildId || g.Name != rc3.Name {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, g, rc3)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc3, err = d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}

		g, err = d.RemoveGuild(rc2.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.GuildId != rc2.GuildId || g.Name != rc2.Name {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, g, rc2)
		}
		if rc2 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		_, err = d.GetGuild(rc2.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		_, err = d.GetGuild(rc3.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.GuildId != rc1.GuildId || g.Name != rc1.Name {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, g, rc1)
		}

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err = d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc3, err = d.AddGuild(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}

		g, err = d.RemoveGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.GuildId != rc1.GuildId || g.Name != rc1.Name {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, g, rc1)
		}

		_, err = d.GetGuild(rc3.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestGuildRemoveD(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did18",
		}
		rc1, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub1-test1",
			ParentId: rc1.GuildId,
		}
		rc2, err := d.AddGuild(g)

		g = &database.Guild{
			Name:     "sub2-test1",
			ParentId: rc2.GuildId,
		}
		rc3, err := d.AddGuild(g)

		g, err = d.RemoveGuildD("unknown did")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.RemoveGuildD(rc2.Name)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.RemoveGuildD(rc1.DiscordId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.GuildId != rc1.GuildId || g.Name != rc1.Name {
			t.Fatalf("[%v] Wrong guild returned. Actual: %v, expected: %v", n, g, rc1)
		}
		if rc1 == g {
			t.Fatalf("[%v] Duplicate of character expected, received original", n)
		}

		_, err = d.GetGuild(rc1.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		_, err = d.GetGuild(rc2.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		_, err = d.GetGuild(rc3.GuildId)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}
