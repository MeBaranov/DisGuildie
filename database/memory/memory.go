package memory

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
)

type MemoryDB struct {
	CharMemoryDb
	GuildMemoryDb
	MoneyMemoryDb
	RoleMemoryDb
	UserMemoryDb
}

// constructor function
func NewMemoryDb() *MemoryDB {
	m := MemoryDB{}
	m.Chars = make(map[string]*database.Character)
	m.Guilds = make(map[uuid.UUID]*database.Guild)
	m.GuildsD = make(map[string]*database.Guild)
	m.Money = make(map[string]*database.Money)
	m.Roles = make(map[string]*database.Role)
	m.UsersD = make(map[string]*database.User)
	return &m
}

func (m *MemoryDB) Export() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(io.Writer(&buf))
	encoder.Encode(m)

	return buf.Bytes(), nil
}

func (m *MemoryDB) Import(b []byte) error {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(io.Reader(buf))
	err := decoder.Decode(m)

	return &database.Error{Code: database.IOErrorDuringImport, Message: err.Error()}
}

func init() {
	var _ database.DataProvider = (*MemoryDB)(nil)
}
