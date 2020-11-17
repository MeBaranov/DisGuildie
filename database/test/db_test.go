package database_test

import (
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/database/memory"
)

var testable map[string]database.DataProvider = map[string]database.DataProvider{
	"memory": memory.NewMemoryDb(),
}

func assertError(oe error, message string, code database.ErrorCode, dbn string) string {
	e := database.ErrToDbErr(oe)
	if e == nil {
		return fmt.Sprintf("Expected Database Error. Received: %v", oe)
	}

	if e.Code != code || e.Message != message {
		return fmt.Sprintf("[%v] Wrong error message. Actual: [%v, %v], Expected: [%v, %v]", dbn, e.Code, e.Message, code, message)
	}

	return ""
}
