package core

import (
	"database/sql"

	"github.com/diamondburned/arikawa/v3/state"
)

var (
	State *state.State
	DB    *sql.DB
)
