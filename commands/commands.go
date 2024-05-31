package commands

import (
	"log"
	"strings"
)

type cmd func()
type Commands struct {
	log *log.Logger
	commands map[string]cmd
}

func NewCommands(log *log.Logger) *Commands {
	return &Commands{log: log, commands: make(map[string]cmd)}
}

func (c *Commands) Exec(command string) {
	if cmd := c.findCommandByLongestPrefix(command); cmd != nil {
		cmd()
	} else {
		c.log.Printf("Command %s not found\n", command)
	}
}

func (c *Commands) findCommandByLongestPrefix(commandPrefix string) cmd {
	longest := -1
	var longestCmd cmd
	for name, cmd := range c.commands {
		if strings.HasPrefix(name, commandPrefix) && len(name) > longest {
			longest = len(name)
			longestCmd = cmd
		}
	}
	return longestCmd
}

func (c *Commands) Register(name string, command cmd) {
	c.commands[name] = command
}
