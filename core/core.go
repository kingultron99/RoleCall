package core

import (
	"database/sql"
	"flag"

	"github.com/diamondburned/arikawa/v3/state"
)

var (
	DevPtr = flag.Bool("dev", false, "Should RoleCall run in dev mode?")
	State  *state.State
	DB     *sql.DB
)
