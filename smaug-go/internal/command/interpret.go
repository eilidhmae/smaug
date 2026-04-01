package command

import (
	"sort"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// CmdFunc is the signature for a command handler function.
type CmdFunc func(ch *types.CharData, argument string)

// Command represents a single game command.
type Command struct {
	Name     string
	DoFun    CmdFunc
	Position int // minimum position
	Level    int // minimum trust level
	Log      int // log type
	Flags    int
}

// Registry holds all registered commands and provides lookup and dispatch.
type Registry struct {
	commands       map[string]*Command
	sorted         []*Command // sorted by name for prefix matching
	SocialFallback func(ch *types.CharData, cmd string, argument string) bool
}

// NewRegistry creates a new empty command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
	}
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd *Command) {
	r.commands[strings.ToLower(cmd.Name)] = cmd
	// Rebuild sorted list
	r.sorted = make([]*Command, 0, len(r.commands))
	for _, c := range r.commands {
		r.sorted = append(r.sorted, c)
	}
	// Sort alphabetically
	sort.Slice(r.sorted, func(i, j int) bool {
		return r.sorted[i].Name < r.sorted[j].Name
	})
}

// Find looks up a command by name or prefix.
func (r *Registry) Find(name string, trust int) *Command {
	name = strings.ToLower(name)
	// Exact match first
	if cmd, ok := r.commands[name]; ok && trust >= cmd.Level {
		return cmd
	}
	// Prefix match
	for _, cmd := range r.sorted {
		if trust >= cmd.Level && strings.HasPrefix(strings.ToLower(cmd.Name), name) {
			return cmd
		}
	}
	return nil
}

// Interpret parses player input and dispatches to the appropriate command.
func (r *Registry) Interpret(ch *types.CharData, argument string) {
	argument = strings.TrimSpace(argument)
	if argument == "" {
		return
	}

	cmdWord, rest := util.OneArgument(argument)
	trust := ch.GetTrust()
	cmd := r.Find(cmdWord, trust)

	if cmd == nil {
		// Try social fallback before giving up
		if r.SocialFallback != nil && r.SocialFallback(ch, cmdWord, rest) {
			return
		}
		ch.Send("Huh?\n\r")
		return
	}

	// Position check
	if ch.Position < cmd.Position {
		switch ch.Position {
		case int(types.POS_DEAD):
			ch.Send("Lie still; you are DEAD.\n\r")
		case int(types.POS_MORTAL), int(types.POS_INCAP):
			ch.Send("You are hurt far too bad for that.\n\r")
		case int(types.POS_STUNNED):
			ch.Send("You are too stunned to do that.\n\r")
		case int(types.POS_SLEEPING):
			ch.Send("In your dreams, or what?\n\r")
		case int(types.POS_RESTING):
			ch.Send("Nah... You feel too relaxed...\n\r")
		case int(types.POS_SITTING):
			ch.Send("Better stand up first.\n\r")
		case int(types.POS_FIGHTING):
			ch.Send("No way!  You are still fighting!\n\r")
		}
		return
	}

	cmd.DoFun(ch, rest)
}
