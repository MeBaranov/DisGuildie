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

		rce, err := d.GetGuildD("did233")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
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

		g, err := d.AddGuildStat(rc2.GuildId, "s1", "t1")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Only top-level guild stats are supported right now", database.GuildLevelError, n); e != "" {
			t.Fatalf(e)
		}

		g, err = d.AddGuildStat(uuid.New(), "s1", "t1")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Guild was not found", database.GuildNotFound, n); e != "" {
			t.Fatalf(e)
		}

		stats := map[string]string{"s1": "t1"}

		g, err = d.AddGuildStat(rc1.GuildId, "s1", "t1")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}

		g, err = d.AddGuildStat(rc1.GuildId, "s1", "t1")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}

		g, err = d.AddGuildStat(rc1.GuildId, "s1", "t2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Stat with same name (s1) but different type (t1) found", database.StatNameConflict, n); e != "" {
			t.Fatalf(e)
		}

		stats["s2"] = "t2"
		g, err = d.AddGuildStat(rc1.GuildId, "s2", "t2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
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

func TestGuildRemoveStat(t *testing.T) {
	for n, d := range testable {
		g := &database.Guild{
			Name:      "test",
			DiscordId: "did166",
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

		d.AddGuildStat(rc1.GuildId, "s1", "t1")
		g, err = d.RemoveGuildStat(rc1.GuildId, "s2")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, g)
		}
		if e := assertError(err, "Stat was not found", database.StatNotFound, n); e != "" {
			t.Fatalf(e)
		}

		stats := map[string]string{"s1": "t1"}
		d.AddGuildStat(rc1.GuildId, "s2", "t2")
		g, err = d.RemoveGuildStat(rc1.GuildId, "s2")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}

		delete(stats, "s1")
		g, err = d.RemoveGuildStat(rc1.GuildId, "s1")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.Stats != nil && !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
		}

		g, err = d.GetGuild(rc1.GuildId)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if g.Stats != nil && !reflect.DeepEqual(g.Stats, stats) {
			t.Fatalf("[%v] Wrong stats. Actual: %v, expected: %v", n, g.Stats, stats)
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
