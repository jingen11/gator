package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jingen11/gator/internal/command"
	"github.com/jingen11/gator/internal/config"
	"github.com/jingen11/gator/internal/database"
	"github.com/jingen11/gator/internal/state"
	_ "github.com/lib/pq"
)

type commands struct {
	Commands map[string]func(*state.State, *command.Command) error
}

func newCommands() commands {
	c := commands{
		Commands: map[string]func(*state.State, *command.Command) error{},
	}
	return c
}

func (c *commands) register(name string, f func(*state.State, *command.Command) error) {
	c.Commands[name] = f
}

func (c *commands) run(name string, s *state.State, cmd *command.Command) error {
	err := c.Commands[name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	c := newCommands()
	conf, err := config.Read()
	if err != nil {
		fmt.Println("error reading config file")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", conf.DbUrl)
	dbQueries := database.New(db)
	if err != nil {
		fmt.Println("error connecting to db")
		os.Exit(1)
	}
	c.register("login", command.HandlerLogin)
	c.register("register", command.HandlerRegister)
	c.register("reset", command.HandlerReset)
	c.register("users", command.HandlerListUsers)
	c.register("agg", command.HandlerAggregate)
	c.register("addfeed", command.MiddlewareLoggedIn(command.HandlerAddFeed))
	c.register("feeds", command.HandlerListFeeds)
	c.register("follow", command.MiddlewareLoggedIn(command.HandlerFollow))
	c.register("following", command.MiddlewareLoggedIn(command.HandlerFollowing))
	c.register("unfollow", command.MiddlewareLoggedIn(command.HandlerUnfollow))
	c.register("browse", command.MiddlewareLoggedIn(command.HandlerBrowse))
	if len(os.Args) == 1 {
		fmt.Println("Invalid arguments")
		os.Exit(1)
	}
	comm := os.Args[1]
	args := os.Args[2:]
	err = c.run(comm, &state.State{
		Conf: &conf,
		Db:   dbQueries,
	}, &command.Command{
		Name:      comm,
		Arguments: args,
	})
	if err != nil {
		fmt.Printf("error running command: %v", err)
		os.Exit(1)
	}
}
