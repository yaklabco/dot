// Package prompt provides utilities for interactive user prompts.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Prompter handles interactive user prompts.
type Prompter struct {
	in  io.Reader
	out io.Writer
}

// New creates a new Prompter that reads from in and writes to out.
func New(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{
		in:  in,
		out: out,
	}
}

// Confirm prompts the user with a yes/no question.
// Returns true if the user answers yes, false otherwise.
// The default behavior is to return false (no) if the user just presses enter.
func (p *Prompter) Confirm(message string) (bool, error) {
	fmt.Fprintf(p.out, "%s [y/N]: ", message)

	scanner := bufio.NewScanner(p.in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("read input: %w", err)
		}
		// EOF
		return false, nil
	}

	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

// ConfirmWithDefault prompts the user with a yes/no question with a specific default.
// If defaultYes is true, the default is yes ([Y/n]), otherwise it's no ([y/N]).
func (p *Prompter) ConfirmWithDefault(message string, defaultYes bool) (bool, error) {
	var prompt string
	if defaultYes {
		prompt = fmt.Sprintf("%s [Y/n]: ", message)
	} else {
		prompt = fmt.Sprintf("%s [y/N]: ", message)
	}

	fmt.Fprint(p.out, prompt)

	scanner := bufio.NewScanner(p.in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("read input: %w", err)
		}
		// EOF - return default
		return defaultYes, nil
	}

	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))

	// If empty answer, use default
	if answer == "" {
		return defaultYes, nil
	}

	return answer == "y" || answer == "yes", nil
}

// Input prompts the user for text input.
func (p *Prompter) Input(message string) (string, error) {
	fmt.Fprintf(p.out, "%s: ", message)

	scanner := bufio.NewScanner(p.in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}
		// EOF
		return "", nil
	}

	return strings.TrimSpace(scanner.Text()), nil
}

// Select prompts the user to choose from a list of options.
// Returns the index of the selected option (0-based) or -1 if invalid/cancelled.
func (p *Prompter) Select(message string, options []string) (int, error) {
	fmt.Fprintln(p.out, message)
	for i, opt := range options {
		fmt.Fprintf(p.out, "  %d) %s\n", i+1, opt)
	}
	fmt.Fprint(p.out, "Enter selection: ")

	scanner := bufio.NewScanner(p.in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return -1, fmt.Errorf("read input: %w", err)
		}
		// EOF
		return -1, nil
	}

	answer := strings.TrimSpace(scanner.Text())
	var selection int
	if _, err := fmt.Sscanf(answer, "%d", &selection); err != nil {
		return -1, nil
	}

	// Convert to 0-based index and validate range
	selection--
	if selection < 0 || selection >= len(options) {
		return -1, nil
	}

	return selection, nil
}
