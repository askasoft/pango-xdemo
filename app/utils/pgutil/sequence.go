package pgutil

import (
	"fmt"
)

func ResetSequence(table string, starts ...int64) string {
	start := int64(1)
	if len(starts) > 0 {
		start = starts[0]
	}
	return fmt.Sprintf("SELECT SETVAL('%s_id_seq', GREATEST((SELECT MAX(id)+1 FROM %s), %d), false)", table, table, start)
}
