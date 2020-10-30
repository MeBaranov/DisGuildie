package database_test

import (
	"fmt"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/database/memory"
)

var testable map[string]database.DataProvider = map[string]database.DataProvider{
	"memory": memory.NewMemoryDb(),
}

func assertError(e error, message string, code database.ErrorCode, dbn string) string {
	wish := fmt.Sprintf("Error '%v': %v", code, message)
	if e.Error() != wish {
		return fmt.Sprintf("[%v] Wrong error message. Actual: %v, Expected: %v", dbn, e.Error(), wish)
	}

	return ""
}
