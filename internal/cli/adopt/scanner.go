// Package adopt provides interactive file adoption.
//
// This file contains Bubble Tea UI code which is excluded from coverage
// requirements as interactive terminal UI cannot be reliably unit tested.
package adopt

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/pkg/dot"
)

// scannerModel represents the loading state while scanning for dotfiles.
type scannerModel struct {
	spinner    spinner.Model
	scanning   bool
	candidates []DotfileCandidate
	err        error
	ctx        context.Context
	fs         domain.FS
	opts       DiscoveryOptions
	client     *dot.Client
	targetDir  string
	currentDir string
	dirCount   int
	totalDirs  int
}

// scanCompleteMsg is sent when discovery completes.
type scanCompleteMsg struct {
	candidates []DotfileCandidate
	err        error
}

// scanProgressMsg is sent when scanning progresses to a new directory.
type scanProgressMsg struct {
	dir   string
	index int
	total int
}

// newScannerModel creates a new scanner model.
func newScannerModel(ctx context.Context, fs domain.FS, opts DiscoveryOptions, client *dot.Client, targetDir string) scannerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return scannerModel{
		spinner:   s,
		scanning:  true,
		ctx:       ctx,
		fs:        fs,
		opts:      opts,
		client:    client,
		targetDir: targetDir,
		totalDirs: len(opts.ScanDirs),
		dirCount:  0,
	}
}

// Init starts the spinner and discovery.
func (m scannerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.discoverDotfiles,
	)
}

// Update handles messages.
func (m scannerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case scanProgressMsg:
		m.currentDir = msg.dir
		m.dirCount = msg.index
		m.totalDirs = msg.total
		return m, nil

	case scanCompleteMsg:
		m.scanning = false
		m.candidates = msg.candidates
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the scanning UI.
func (m scannerModel) View() string {
	if !m.scanning {
		return ""
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	msg := fmt.Sprintf("\n %s Scanning for dotfiles...", m.spinner.View())

	if m.currentDir != "" {
		msg += fmt.Sprintf("\n   %s (%d/%d)",
			dimStyle.Render(m.currentDir),
			m.dirCount,
			m.totalDirs)
	}

	return msg + "\n\n"
}

// discoverDotfiles runs the discovery in the background.
func (m scannerModel) discoverDotfiles() tea.Msg {
	candidates, err := discoverDotfilesWithProgress(m.ctx, m.fs, m.opts, m.client, m.targetDir)
	return scanCompleteMsg{
		candidates: candidates,
		err:        err,
	}
}

// discoverDotfilesWithProgress wraps DiscoverDotfiles and sends progress updates.
func discoverDotfilesWithProgress(
	ctx context.Context,
	fs domain.FS,
	opts DiscoveryOptions,
	client *dot.Client,
	targetDir string,
) ([]DotfileCandidate, error) {
	var allCandidates []DotfileCandidate

	for i, scanDir := range opts.ScanDirs {
		// Send progress update
		if progressChan != nil {
			select {
			case progressChan <- scanProgressMsg{dir: scanDir, index: i + 1, total: len(opts.ScanDirs)}:
			default:
			}
		}

		// Scan this directory
		optsForDir := opts
		optsForDir.ScanDirs = []string{scanDir}

		candidates, err := DiscoverDotfiles(ctx, fs, optsForDir, client, targetDir)
		if err != nil {
			return nil, err
		}

		allCandidates = append(allCandidates, candidates...)
	}

	return allCandidates, nil
}

// progressCallback is a function that reports scanning progress.
type progressCallback func(dir string, index, total int)

// ScanWithProgress runs the discovery with a progress spinner.
func ScanWithProgress(
	ctx context.Context,
	fs domain.FS,
	opts DiscoveryOptions,
	client *dot.Client,
	targetDir string,
) ([]DotfileCandidate, error) {
	m := newScannerModel(ctx, fs, opts, client, targetDir)

	p := tea.NewProgram(m)

	// Store program reference for sending progress updates
	progressChan = make(chan scanProgressMsg, 10)

	go func() {
		for msg := range progressChan {
			p.Send(msg)
		}
	}()

	finalModel, err := p.Run()
	close(progressChan)

	if err != nil {
		return nil, fmt.Errorf("failed to run scanner: %w", err)
	}

	final := finalModel.(scannerModel)
	if final.err != nil {
		return nil, final.err
	}

	return final.candidates, nil
}

// progressChan is used to send progress updates to the scanner.
var progressChan chan scanProgressMsg
