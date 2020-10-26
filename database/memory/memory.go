package memory

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/mebaranov/disguildie/database"
)

type MemoryDB struct {
	CharMemoryDb
	GuildMemoryDb
	MoneyMemoryDb
	RoleMemoryDb
	UserMemoryDb
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

	return err
}

func init() {
	var _ database.DataProvider = (*MemoryDB)(nil)
}
