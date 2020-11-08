package helpers

import (
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
)

type AdminGuildProcessor struct {
	prov  database.DataProvider
	funcs map[string]func(message.Message)
}

func NewAdminGuildProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminGuildProcessor{prov: prov}
	ap.funcs = map[string]func(message.Message){}
	return nil
}
