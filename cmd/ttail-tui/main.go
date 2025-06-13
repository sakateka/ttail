package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakateka/ttail"
	"github.com/sakateka/ttail/internal/config"
)

var (
	flagDuration         time.Duration
	flagLogTypes         []string
	flagConfigFile       string
	flagTimeFromLastLine bool
	flagMaxLines         int
	flagAutoDetect       bool
)

// Custom flag type for multiple log types
type logTypeFlag []string

func (ltf *logTypeFlag) String() string {
	return strings.Join(*ltf, ",")
}

func (ltf *logTypeFlag) Set(value string) error {
	*ltf = append(*ltf, value)
	return nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] file [file ...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCompact TUI log tailer (ripgrep-style)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s app1.log app2.log                    # Auto-detect log types\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t tskv -t java app1.log app2.log    # Specify types per file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t java *.log                       # Same type for all files\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nSupported log types: tskv, java, apache, nginx, docker, json, syslog, etc.\n\n")
		flag.PrintDefaults()
	}

	flag.DurationVar(&flagDuration, "n", 10*time.Second, "time duration to tail")
	flag.StringVar(&flagConfigFile, "c", "", "config file path")
	flag.BoolVar(&flagTimeFromLastLine, "l", false, "tail from last line timestamp")
	flag.IntVar(&flagMaxLines, "max-lines", 10, "maximum lines to show when expanded")
	flag.BoolVar(&flagAutoDetect, "auto", false, "auto-detect log types (default behavior when no -t specified)")
}

func main() {
	// Parse flags manually to handle the custom flag
	var logTypesFlag logTypeFlag
	flag.Var(&logTypesFlag, "t", "log type (can be specified multiple times, one per file)")

	flag.Parse()
	flagLogTypes = []string(logTypesFlag)

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	files := flag.Args()

	// Initialize the TUI model
	model := newModel(files)

	// Create the Bubble Tea program without alt screen (preserves terminal history)
	p := tea.NewProgram(model)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}

// LogFile represents a single log file being tailed
type LogFile struct {
	Path        string
	DisplayName string
	LogType     string
	LastLine    string
	LastUpdate  time.Time
	TFile       *ttail.TFile
	File        *os.File
	Lines       []string
	IsExpanded  bool
	Error       string
}

// Model represents the TUI application state
type Model struct {
	logFiles     []*LogFile
	selectedFile int
	maxLines     int
	ctx          context.Context
	cancel       context.CancelFunc
}

// Message types for Bubble Tea
type tickMsg time.Time
type logUpdateMsg struct {
	fileIndex int
	lines     []string
	lastLine  string
	err       error
}

// Ripgrep-like styles - very minimal
var (
	// File name style (like ripgrep's green filename)
	filenameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87")).
			Bold(true)

	// Expanded file name style (more prominent)
	expandedFilenameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF87")).
				Background(lipgloss.Color("#2D3748")).
				Bold(true).
				Padding(0, 1)

	// Selected file indicator
	selectedIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B")).
				Bold(true)

	// Expanded file indicator (more prominent)
	expandedIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	// Log type style (muted, like ripgrep's line numbers)
	logTypeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))

	// Expanded log type style (more prominent)
	expandedLogTypeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				Bold(true)

	// Content style (normal text - plain white)
	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Expanded content style (plain white, no bold)
	expandedContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8"))

	// Expanded line style (like ripgrep's context lines - plain white, no italic)
	expandedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))
)

func newModel(filePaths []string) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	model := &Model{
		logFiles:     make([]*LogFile, 0, len(filePaths)),
		selectedFile: 0,
		maxLines:     flagMaxLines,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Determine log types for each file
	logTypes := determineLogTypes(filePaths)

	// Initialize log files
	for i, path := range filePaths {
		logFile := &LogFile{
			Path:        path,
			DisplayName: filepath.Base(path),
			LogType:     logTypes[i],
			Lines:       make([]string, 0, flagMaxLines),
			LastUpdate:  time.Now(),
		}

		if err := model.initLogFile(logFile); err != nil {
			logFile.Error = err.Error()
		}

		model.logFiles = append(model.logFiles, logFile)
	}

	return model
}

// determineLogTypes assigns log types to files based on flags or auto-detection
func determineLogTypes(filePaths []string) []string {
	logTypes := make([]string, len(filePaths))

	// If no log types specified, auto-detect
	if len(flagLogTypes) == 0 || flagAutoDetect {
		for i, path := range filePaths {
			logTypes[i] = autoDetectLogType(path)
		}
		return logTypes
	}

	// If one log type specified, use it for all files
	if len(flagLogTypes) == 1 {
		for i := range filePaths {
			logTypes[i] = flagLogTypes[0]
		}
		return logTypes
	}

	// If multiple log types specified, match them to files
	for i := range filePaths {
		if i < len(flagLogTypes) {
			logTypes[i] = flagLogTypes[i]
		} else {
			// Fall back to auto-detection for extra files
			logTypes[i] = autoDetectLogType(filePaths[i])
		}
	}

	return logTypes
}

// autoDetectLogType attempts to detect the log type by reading a few lines
func autoDetectLogType(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return "tskv" // default fallback
	}
	defer file.Close()

	// Read first few lines to detect format
	buf := make([]byte, 4096)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "tskv" // default fallback
	}

	content := string(buf[:n])
	lines := strings.Split(content, "\n")

	// Load config to get all available log types
	conf, err := config.LoadConfig(flagConfigFile)
	if err != nil {
		return "tskv" // default fallback
	}

	// Test each log type against the sample lines
	for logTypeName, logType := range conf {
		if logType.TimeReStr == "" {
			continue
		}

		re, err := regexp.Compile(logType.TimeReStr)
		if err != nil {
			continue
		}

		// Check if this regex matches any of the first few lines
		matchCount := 0
		for i, line := range lines {
			if i >= 5 { // Only check first 5 lines
				break
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			if re.MatchString(line) {
				matchCount++
			}
		}

		// If we found matches, this is likely the correct type
		if matchCount > 0 {
			return logTypeName
		}
	}

	return "tskv" // default fallback
}

func (m *Model) initLogFile(lf *LogFile) error {
	file, err := os.Open(lf.Path)
	if err != nil {
		return err
	}

	// Build options
	opts := []ttail.TimeFileOptions{
		ttail.WithDuration(flagDuration),
		ttail.WithTimeFromLastLine(flagTimeFromLastLine),
	}

	if lf.LogType != "" {
		logOpts, err := ttail.OptionsFromConfig(lf.LogType, flagConfigFile)
		if err != nil {
			return err
		}
		opts = append(opts, logOpts...)
	}

	tfile := ttail.NewTimeFile(file, opts...)

	lf.File = file
	lf.TFile = tfile

	return nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.updateAllLogs(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel()
			return m, tea.Quit
		case "up", "k":
			if m.selectedFile > 0 {
				m.selectedFile--
			}
			return m, nil
		case "down", "j":
			if m.selectedFile < len(m.logFiles)-1 {
				m.selectedFile++
			}
			return m, nil
		case "enter", " ":
			// Toggle expanded state of selected file and close others
			if m.selectedFile < len(m.logFiles) {
				selectedFile := m.logFiles[m.selectedFile]
				newExpandedState := !selectedFile.IsExpanded

				// Close all files first
				for _, lf := range m.logFiles {
					lf.IsExpanded = false
				}

				// Then expand the selected one if it wasn't expanded
				selectedFile.IsExpanded = newExpandedState
			}
			return m, nil
		}

	case tickMsg:
		return m, tea.Batch(
			m.tickCmd(),
			m.updateAllLogs(),
		)

	case logUpdateMsg:
		if msg.fileIndex < len(m.logFiles) {
			lf := m.logFiles[msg.fileIndex]
			if msg.err != nil {
				lf.Error = msg.err.Error()
			} else {
				lf.Lines = msg.lines
				lf.LastLine = msg.lastLine
				lf.LastUpdate = time.Now()
				lf.Error = ""
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	var lines []string

	// Ensure we always have at least one line to prevent rendering issues
	if len(m.logFiles) == 0 {
		return "No log files loaded"
	}

	// Ripgrep-style output - very clean and minimal
	for i, lf := range m.logFiles {
		// Determine if this file is expanded
		isExpanded := lf.IsExpanded
		isSelected := i == m.selectedFile

		// Selection indicator with enhanced styling for expanded files
		indicator := " "
		if isSelected {
			if isExpanded {
				indicator = expandedIndicator.Render("▼")
			} else {
				indicator = selectedIndicator.Render("►")
			}
		}

		// Filename and log type with enhanced styling for expanded files
		var filename, logType, content string

		if isExpanded {
			filename = expandedFilenameStyle.Render(lf.DisplayName)
			logType = expandedLogTypeStyle.Render(fmt.Sprintf("[%s]", lf.LogType))
		} else {
			filename = filenameStyle.Render(lf.DisplayName)
			logType = logTypeStyle.Render(fmt.Sprintf("[%s]", lf.LogType))
		}

		// Content or error with enhanced styling for expanded files
		if lf.Error != "" {
			content = errorStyle.Render(fmt.Sprintf("ERROR: %s", lf.Error))
		} else if lf.LastLine != "" {
			if isExpanded {
				content = expandedContentStyle.Render(lf.LastLine)
			} else {
				content = contentStyle.Render(lf.LastLine)
			}
		} else {
			content = logTypeStyle.Render("(no data)")
		}

		// Main line: ► filename [type] content
		// Always construct and add the main line - this should never be empty
		mainLine := fmt.Sprintf("%s%s %s %s", indicator, filename, logType, content)
		lines = append(lines, mainLine)

		// Expanded content (like ripgrep's context lines)
		if isExpanded && len(lf.Lines) > 0 {
			for _, line := range lf.Lines {
				if strings.TrimSpace(line) != "" {
					expandedLine := expandedStyle.Render(fmt.Sprintf("  │ %s", line))
					lines = append(lines, expandedLine)
				}
			}
		}
	}

	// Ensure we always return something visible
	if len(lines) == 0 {
		return "No content to display"
	}

	return strings.Join(lines, "\n")
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) updateAllLogs() tea.Cmd {
	var cmds []tea.Cmd

	for i, lf := range m.logFiles {
		if lf.TFile != nil && lf.Error == "" {
			cmds = append(cmds, m.updateLogFile(i, lf))
		}
	}

	return tea.Batch(cmds...)
}

func (m Model) updateLogFile(index int, lf *LogFile) tea.Cmd {
	return func() tea.Msg {
		// Re-initialize the file position for fresh reads
		if err := lf.TFile.FindPosition(); err != nil {
			return logUpdateMsg{fileIndex: index, err: err}
		}

		// Read new content
		reader, err := lf.TFile.GetReader()
		if err != nil {
			return logUpdateMsg{fileIndex: index, err: err}
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			return logUpdateMsg{fileIndex: index, err: err}
		}

		if len(content) == 0 {
			return logUpdateMsg{fileIndex: index, lines: lf.Lines, lastLine: lf.LastLine}
		}

		// Split into lines and keep only the last N lines
		allLines := strings.Split(strings.TrimSpace(string(content)), "\n")

		// Filter out empty lines
		var validLines []string
		for _, line := range allLines {
			if strings.TrimSpace(line) != "" {
				validLines = append(validLines, line)
			}
		}

		// Keep only the last N lines for preview
		startIdx := 0
		if len(validLines) > m.maxLines {
			startIdx = len(validLines) - m.maxLines
		}

		previewLines := validLines[startIdx:]
		lastLine := ""
		if len(validLines) > 0 {
			lastLine = validLines[len(validLines)-1]
		}

		return logUpdateMsg{
			fileIndex: index,
			lines:     previewLines,
			lastLine:  lastLine,
		}
	}
}
