# TTail TUI - Interactive Log Viewer

A beautiful Terminal User Interface (TUI) for tailing multiple log files simultaneously with live preview.

## Features

- **Multi-file Support**: Monitor multiple log files at once
- **Interactive Selection**: Navigate between files with arrow keys
- **Live Preview**: Real-time log updates in a dedicated preview window
- **No Terminal Scroll**: Maintains terminal history while updating preview
- **Built-in Log Types**: Supports 25+ log formats out of the box
- **Customizable**: Configurable preview window size and update intervals

## Usage

```bash
# Basic usage with multiple files
ttail-tui test-logs/app1.log test-logs/app2.log

# Specify log types for different files
ttail-tui -t tskv test-logs/app1.log -t java test-logs/app2.log

# Custom preview window size
ttail-tui -lines 30 test-logs/*.log

# Tail from last line timestamp
ttail-tui -l -n 5m test-logs/*.log
```

## Controls

- **↑/↓** or **j/k**: Navigate between log files
- **ENTER** or **SPACE**: Toggle preview for selected file
- **q** or **Ctrl+C**: Quit application

## Interface Layout

```
  app1.log: timestamp=2025-06-14T12:34...
 ▼app2.log: 2025-06-14 12:34:55 [INFO]...
  | 2025-06-14T12:34:50 level=info
  | 2025-06-14T12:34:51 level=debug
  | 2025-06-14T12:34:52 level=info
  | ...
  | ...
  app3.log: {"timestamp": "2025-06-14T1...
```

## Options

- `-n duration`: Time duration to tail (default: 10s)
- `-t type`: Log type (tskv, java, apache, nginx, etc.)
- `-c config`: Custom configuration file
- `-l`: Tail from last line timestamp instead of current time
- `-lines N`: Number of lines in preview window (default: 50)

## Demo

To test the TUI with simulated live logs:

```bash
# Terminal 1: Start log simulation
./simulate-logs.sh

# Terminal 2: Start TUI
ttail-tui -t tskv test-logs/app1.log -t java test-logs/app2.log
```

## Log Format Support

The TUI supports all built-in log formats:

- **TSKV**: `timestamp=2025-06-14T12:34:56	level=info	msg=...`
- **Java**: `2025-06-14 12:34:56 [INFO] Application message`
- **Apache**: `[14/Jun/2025:12:34:56 +0000] "GET / HTTP/1.1"`
- **Docker**: `2025-06-14T12:34:56.123456789Z Container message`
- **JSON**: `{"timestamp":"2025-06-14T12:34:56","level":"info"}`
- And 20+ more formats...

## Performance

The TUI is optimized for:
- **Low CPU Usage**: Updates only when files change
- **Memory Efficient**: Keeps only preview window lines in memory
- **Responsive UI**: Non-blocking updates with smooth navigation
- **Large Files**: Efficient binary search for timestamp-based positioning

## Tips

1. **Multiple Log Types**: Use different `-t` flags for different file formats
2. **Large Files**: The tool efficiently handles large log files using binary search
3. **Time Ranges**: Use `-n` to control how far back to look for log entries
4. **Real-time Monitoring**: Files are checked every second for new content
5. **Terminal Size**: The preview window adapts to your terminal size

## Troubleshooting

**No logs showing?**
- Check if the log format matches the specified type
- Verify file permissions
- Ensure timestamps are within the specified duration

**Performance issues?**
- Reduce preview lines with `-lines`
- Use more specific time ranges with `-n`
- Check if log files are extremely large

**UI not updating?**
- Ensure terminal supports ANSI colors
- Try resizing the terminal window
- Check if files are being actively written to
