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

func (mdb *MoneyMemoryDb) AddMoney(m *database.Money) (*database.Money, *database.Error) {
	mdb.mux.Lock()
	defer mdb.mux.Unlock()

	if _, ok := mdb.money[m.GuildId]; ok {
		return nil, &database.Error{Code: database.MoneyAlreadyRegistered, Message: "Payment stuff for the guild is already registered"}
	}

	newM := *m
	m = &newM
	mdb.money[m.GuildId] = m

	tmp := *m
	return &tmp, nil
}

func (mdb *MoneyMemoryDb) GetMoney(g string) (*database.Money, *database.Error) {
	if m, ok := mdb.money[g]; ok {
		tmp := *m
		return &tmp, nil
	}
	return nil, &database.Error{Code: database.MoneyNotFound, Message: "Payment stuff for the guild is not found"}
}

func (mdb *MoneyMemoryDb) ChangeMoneyOwner(g string, u string) (*database.Money, *database.Error) {
	m, ok := mdb.money[g]

	if !ok {
		return nil, &database.Error{Code: database.MoneyNotFound, Message: "Payment stuff for the guild is not found"}
	}

	m.UserId = u
	tmp := *m
	return &tmp, nil
}

func (mdb *MoneyMemoryDb) SetMoneyValid(g string, t time.Time) (*database.Money, *database.Error) {
	m, ok := mdb.money[g]

	if !ok {
		return nil, &database.Error{Code: database.MoneyNotFound, Message: "Payment stuff for the guild is not found"}
	}

	m.ValidTo = t

	tmp := *m
	return &tmp, nil
}
