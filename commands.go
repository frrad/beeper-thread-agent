package main

import "strings"

type Command struct {
	Name  string
	Text  string
	Since string
}

func ParseCommand(input string) (Command, bool) {
	text := strings.TrimSpace(input)
	if text == "" {
		return Command{}, false
	}
	if !strings.HasPrefix(text, "/") {
		return Command{Name: "steer", Text: text}, true
	}
	fields := strings.Fields(strings.TrimPrefix(text, "/"))
	if len(fields) == 0 {
		return Command{}, false
	}
	name := strings.ToLower(fields[0])
	if name == "start" {
		cmd := Command{Name: name}
		if len(fields) > 1 {
			cmd.Since = fields[1]
		}
		return cmd, true
	}
	for _, allowed := range []string{"pause", "resume", "status", "stop", "help"} {
		if name == allowed {
			return Command{Name: name}, true
		}
	}
	return Command{}, false
}
