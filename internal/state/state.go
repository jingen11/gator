package state

import (
	"github.com/jingen11/gator/internal/config"
	"github.com/jingen11/gator/internal/database"
)

type State struct {
	Db   *database.Queries
	Conf *config.Config
}
