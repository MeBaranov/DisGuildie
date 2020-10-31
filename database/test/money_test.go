package database_test

import (
	"testing"
	"time"

	"github.com/mebaranov/disguildie/database"
)

func TestMoneyAdd(t *testing.T) {
	for n, d := range testable {
		m := &database.Money{
			GuildId: "gid1",
			UserId:  "uid1",
			Price:   10,
			ValidTo: time.Now(),
		}

		rc, err := d.AddMoney(m)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != m.UserId || rc.Price != m.Price || rc.ValidTo != m.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}

		m = &database.Money{
			GuildId: "gid1",
			UserId:  "uid11",
			Price:   11,
			ValidTo: time.Now(),
		}

		rce, err := d.AddMoney(m)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rce)
		}
		if e := assertError(err, "Payment stuff for the guild is already registered", database.MoneyAlreadyRegistered, n); e != "" {
			t.Fatalf(e)
		}

		m = &database.Money{
			GuildId: "gid11",
			UserId:  "uid11",
			Price:   11,
			ValidTo: time.Now(),
		}

		rc, err = d.AddMoney(m)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != m.UserId || rc.Price != m.Price || rc.ValidTo != m.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}
	}
}

func TestMoneyGet(t *testing.T) {
	for n, d := range testable {
		g := "gid12"
		m := &database.Money{
			GuildId: g,
			UserId:  "uid12",
			Price:   10,
			ValidTo: time.Now(),
		}

		d.AddMoney(m)
		rc, err := d.GetMoney(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != m.UserId || rc.Price != m.Price || rc.ValidTo != m.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}

		m2 := &database.Money{
			GuildId: g,
			UserId:  "uid22",
			Price:   11,
			ValidTo: time.Now(),
		}

		d.AddMoney(m2)
		rc, err = d.GetMoney(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != m.UserId || rc.Price != m.Price || rc.ValidTo != m.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}

		g2 := "gid22"
		m2 = &database.Money{
			GuildId: g2,
			UserId:  "uid22",
			Price:   11,
			ValidTo: time.Now(),
		}

		rc, err = d.AddMoney(m2)
		rc, err = d.GetMoney(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != m.UserId || rc.Price != m.Price || rc.ValidTo != m.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}

		rc, err = d.GetMoney(g2)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m2.GuildId || rc.UserId != m2.UserId || rc.Price != m2.Price || rc.ValidTo != m2.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}

		rc, err = d.GetMoney("unknown")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "Payment stuff for the guild is not found", database.MoneyNotFound, n); e != "" {
			t.Fatalf(e)
		}
	}
}

func TestMoneyChangeOwner(t *testing.T) {
	for n, d := range testable {
		g := "gid13"
		m := &database.Money{
			GuildId: g,
			UserId:  "uid13",
			Price:   10,
			ValidTo: time.Now(),
		}

		rc, err := d.ChangeMoneyOwner(g, "uid23")
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "Payment stuff for the guild is not found", database.MoneyNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddMoney(m)
		rc, err = d.ChangeMoneyOwner(g, "uid23")
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != "uid23" || rc.Price != m.Price || rc.ValidTo != m.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}

		rc, err = d.GetMoney(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != "uid23" || rc.Price != m.Price || rc.ValidTo != m.ValidTo {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}
	}
}

func TestMoneySetValid(t *testing.T) {
	for n, d := range testable {
		g := "gid14"
		m := &database.Money{
			GuildId: g,
			UserId:  "uid14",
			Price:   10,
			ValidTo: time.Now(),
		}

		target := time.Now().Add(time.Second * 100)
		rc, err := d.SetMoneyValid(g, target)
		if err == nil {
			t.Fatalf("[%v] Error expected. Received: %v", n, rc)
		}
		if e := assertError(err, "Payment stuff for the guild is not found", database.MoneyNotFound, n); e != "" {
			t.Fatalf(e)
		}

		d.AddMoney(m)
		rc, err = d.SetMoneyValid(g, target)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != m.UserId || rc.Price != m.Price || rc.ValidTo != target {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}

		rc, err = d.GetMoney(g)
		if err != nil {
			t.Fatalf("[%v] No errors expected. Received: %v", n, err)
		}
		if rc.GuildId != m.GuildId || rc.UserId != m.UserId || rc.Price != m.Price || rc.ValidTo != target {
			t.Fatalf("[%v] Wrong money returned. Actual: %v, expected: %v", n, rc, m)
		}
	}
}
