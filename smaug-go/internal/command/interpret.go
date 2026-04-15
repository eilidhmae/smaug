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
	// ObjCommandHook is consulted only when no registered command matches the
	// input. It gives obj-progs a chance to consume an unknown command. If it
	// returns true, dispatch stops; otherwise the room hook + social fallback
	// are tried. Wired from main to mudprog.OprogCommandTrigger to avoid an
	// import cycle (mudprog imports command).
	ObjCommandHook func(ch *types.CharData, line string) bool
	// RoomCommandHook is consulted after ObjCommandHook (still in the
	// unknown-command path) before the social fallback. Wired from main to
	// mudprog.RprogCommandTrigger. Full precedence (C-matching, see
	// src/interp.c): normal command → obj-prog CMD → room-prog CMD →
	// social fallback → "Huh?". Hooks fire in the unknown-command path so a
	// greedy CMD prog cannot swallow real commands like "quit" or "north".
	RoomCommandHook func(ch *types.CharData, line string) bool
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

// All returns the sorted list of registered commands. The returned slice
// is a shared view for iteration (wizhelp, help, etc.) — callers must not
// mutate it. Safe because the game is single-threaded.
func (r *Registry) All() []*Command {
	return r.sorted
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

	// Alias expansion. Aliases only apply to PCs, and use C alias.c:42's
	// prefix-match semantics (`!str_prefix(argument, pal->name)`): typing "g"
	// fires an alias named "get". We track CmdRecurse on the character to
	// prevent infinite alias loops (e.g. "alias g get" where g expands to a
	// chain that eventually hits g again).
	if ch != nil && !ch.IsNPC() && ch.PCData != nil && len(ch.PCData.Aliases) > 0 {
		cmdWord, rest := util.OneArgument(argument)
		lc := strings.ToLower(cmdWord)
		for _, a := range ch.PCData.Aliases {
			if a.Cmd != "" && strings.HasPrefix(strings.ToLower(a.Name), lc) {
				if ch.CmdRecurse < 0 {
					// Marked exhausted earlier this dispatch chain.
					ch.CmdRecurse = 0
					return
				}
				ch.CmdRecurse++
				if ch.CmdRecurse > 50 {
					ch.Send("Unable to further process command, recurses too much.\n\r")
					ch.CmdRecurse = -1
					return
				}
				expanded := a.Cmd
				if rest != "" {
					expanded += " " + rest
				}
				r.Interpret(ch, expanded)
				ch.CmdRecurse--
				if ch.CmdRecurse < 0 {
					ch.CmdRecurse = 0
				}
				return
			}
		}
	}

	cmdWord, rest := util.OneArgument(argument)
	trust := ch.GetTrust()
	cmd := r.Find(cmdWord, trust)

	if cmd == nil {
		// No registered command matched. Mirror C's interp.c order: try the
		// obj-prog CMD hook first, then the room-prog CMD hook, then the
		// social fallback, before finally giving up with "Huh?".
		if r.ObjCommandHook != nil && r.ObjCommandHook(ch, argument) {
			return
		}
		if r.RoomCommandHook != nil && r.RoomCommandHook(ch, argument) {
			return
		}
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

// InterpretWithTrustCap is like Interpret but caps the effective trust level
// for command lookup. Used by force to prevent privilege escalation.
func (r *Registry) InterpretWithTrustCap(ch *types.CharData, argument string, maxTrust int) {
	argument = strings.TrimSpace(argument)
	if argument == "" {
		return
	}

	cmdWord, rest := util.OneArgument(argument)
	trust := ch.GetTrust()
	if trust > maxTrust {
		trust = maxTrust
	}
	cmd := r.Find(cmdWord, trust)

	if cmd == nil {
		if r.SocialFallback != nil && r.SocialFallback(ch, cmdWord, rest) {
			return
		}
		ch.Send("Huh?\n\r")
		return
	}

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
