package database_test

import (
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/database/memory"
)

var testable map[string]database.DataProvider = map[string]database.DataProvider{
	"memory": memory.NewMemoryDb(),
}

func assertError(e *database.Error, message string, code database.ErrorCode, dbn string) string {
	if e.Code != code || e.Message != message {
		return fmt.Sprintf("[%v] Wrong error message. Actual: [%v, %v], Expected: [%v, %v]", dbn, e.Code, e.Message, code, message)
	}

	return ""
}
