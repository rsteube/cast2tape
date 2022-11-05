package main

import (
	"fmt"
	"strings"
	"time"
)

// sleepThreshold is the time at which if there has been no activity in the
// tape file we insert a Sleep command
const sleepThreshold = 500 * time.Millisecond

// EscapeSequences is a map of escape sequences to their VHS commands.
var EscapeSequences = map[string]string{
	"\x1b[A":  UP,
	"\x1b[B":  DOWN,
	"\x1b[C":  RIGHT,
	"\x1b[D":  LEFT,
	"\x1b[1~": HOME,
	"\x1b[2~": INSERT,
	"\x1b[3~": DELETE,
	"\x1b[4~": END,
	"\x01":    CTRL + "+A",
	"\x02":    CTRL + "+B",
	"\x03":    CTRL + "+C",
	"\x04":    CTRL + "+D",
	"\x05":    CTRL + "+E",
	"\x06":    CTRL + "+F",
	"\x07":    CTRL + "+G",
	"\x08":    BACKSPACE,
	"\x09":    TAB,
	"\x0b":    CTRL + "+K",
	"\x0c":    CTRL + "+L",
	"\x0d":    ENTER,
	"\x0e":    CTRL + "+N",
	"\x0f":    CTRL + "+O",
	"\x10":    CTRL + "+P",
	"\x11":    CTRL + "+Q",
	"\x12":    CTRL + "+R",
	"\x13":    CTRL + "+S",
	"\x14":    CTRL + "+T",
	"\x15":    CTRL + "+U",
	"\x16":    CTRL + "+V",
	"\x17":    CTRL + "+W",
	"\x18":    CTRL + "+X",
	"\x19":    CTRL + "+Y",
	"\x1a":    CTRL + "+Z",
	"\x1b":    ESCAPE,
	"\x7f":    BACKSPACE,
}

// Record is a command that starts a pseudo-terminal for the user to begin
// writing to, it records all the key presses on stdin and uses them to write
// Tape commands.
//
// vhs record > file.tape

// inputToTape takes input from a PTY stdin and converts it into a tape file.
func inputToTape(input string) string {
	// If the user exited the shell by typing exit don't record this in the
	// command.
	//
	// NOTE: this is not very robust as if someone types exii<BS>t it will not work
	// correctly and the exit will show up. In this case, the user should edit the
	// tape file.
	s := strings.TrimSuffix(strings.TrimSpace(input), "exit")
	for sequence, command := range EscapeSequences {
		s = strings.ReplaceAll(s, sequence, "\n"+command+"\n")
	}

	s = strings.ReplaceAll(s, "\n\n", "\n")

	var sanitized strings.Builder
	lines := strings.Split(s, "\n")

	for i := 0; i < len(lines)-1; i++ {
		// Group repeated commands to compress file and make it more readable.
		repeat := 1
		for lines[i] == lines[i+repeat] {
			repeat++
		}
		i += repeat - 1

		// We've encountered some non-command, assume that we need to type these
		// characters.
		if TokenType(lines[i]) == SLEEP {
			sleep := sleepThreshold * time.Duration(repeat)
			sanitized.WriteString(fmt.Sprintf("%s %s", TokenType(SLEEP), sleep))
		} else if strings.HasPrefix(lines[i], CTRL) {
			for j := 0; j < repeat; j++ {
				sanitized.WriteString("Ctrl" + strings.TrimPrefix(lines[i], CTRL) + "\n")
			}
			continue
		} else if IsCommand(TokenType(lines[i])) {
			sanitized.WriteString(fmt.Sprint(TokenType(lines[i])))
			if repeat > 1 {
				sanitized.WriteString(fmt.Sprint(" ", repeat))
			}
		} else {
			sanitized.WriteString(fmt.Sprintln(TokenType(TYPE), quote(lines[i])))
			continue
		}
		sanitized.WriteRune('\n')
	}

	return sanitized.String()
}

// quote wraps a string in (single or double) quotes
func quote(s string) string {
	if strings.ContainsRune(s, '"') {
		return fmt.Sprintf(`'%s'`, s)
	}
	return fmt.Sprintf(`"%s"`, s)
}
