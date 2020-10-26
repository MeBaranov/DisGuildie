package memory

import (
	"sync"
	"time"

	"github.com/mebaranov/disguildie/database"
)

type MoneyMemoryDb struct {
	money map[string]*database.Money
	mux   sync.Mutex
}

func (mdb *MoneyMemoryDb) AddMoney(m *database.Money) (*database.Money, error) {
	mdb.mux.Lock()
	defer mdb.mux.Unlock()

	if _, ok := mdb.money[m.GuildId]; ok {
		return nil, &database.DbError{Code: database.MoneyAlreadyRegistered, Message: "Payment stuff for the guild is already registered"}
	}

	mdb.money[m.GuildId] = m
	return m, nil
}

func (mdb *MoneyMemoryDb) GetMoneyGuid(g string) (*database.Money, error) {
	if m, ok := mdb.money[g]; ok {
		return m, nil
	}
	return nil, &database.DbError{Code: database.MoneyNotFound, Message: "Payment stuff for the guild is not found"}
}

func (mdb *MoneyMemoryDb) ChangeGuildOwner(g string, u string) (*database.Money, error) {
	m, ok := mdb.money[g]

	if !ok {
		return nil, &database.DbError{Code: database.MoneyNotFound, Message: "Payment stuff for the guild is not found"}
	}

	m.UserId = u
	return m, nil
}

func (mdb *MoneyMemoryDb) SetMoneyValid(g string, t time.Time) (*database.Money, error) {
	m, ok := mdb.money[g]

	if !ok {
		return nil, &database.DbError{Code: database.MoneyNotFound, Message: "Payment stuff for the guild is not found"}
	}

	m.ValidTo = t

	return m, nil
}
