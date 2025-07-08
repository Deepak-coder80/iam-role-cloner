package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

type Logger struct {
	verbose bool
	logFile *os.File
}

// new logger instance
func New(verbose bool, logFileName string) (*Logger, error) {
	var logFile *os.File
	var err error

	if logFileName != "" {
		logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
	}

	return &Logger{
		verbose: verbose,
		logFile: logFile,
	}, nil
}

// Close the log file
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

// Info logs informational messages
func (l *Logger) Info(message string) {
	timestamp := time.Now().Format("15:04:05")
	coloredMessage := color.New(color.FgBlue).Sprintf("[INFO] %s", message)

	fmt.Printf("%s %s\n", color.New(color.FgCyan).Sprint(timestamp), coloredMessage)
	l.writeToFile("INFO", message)
}

// Success logs success messages
func (l *Logger) Success(message string) {
	timestamp := time.Now().Format("15:04:05")
	coloredMessage := color.New(color.FgGreen).Sprintf("[SUCCESS] %s", message)

	fmt.Printf("%s %s\n", color.New(color.FgCyan).Sprint(timestamp), coloredMessage)
	l.writeToFile("SUCCESS", message)
}

// Warning logs warning messages
func (l *Logger) Warning(message string) {
	timestamp := time.Now().Format("15:04:05")
	coloredMessage := color.New(color.FgYellow).Sprintf("[WARNING] %s", message)

	fmt.Printf("%s %s\n", color.New(color.FgCyan).Sprint(timestamp), coloredMessage)
	l.writeToFile("WARNING", message)
}

// Error logs error messages
func (l *Logger) Error(message string) {
	timestamp := time.Now().Format("15:04:05")
	coloredMessage := color.New(color.FgRed).Sprintf("[ERROR] %s", message)

	fmt.Printf("%s %s\n", color.New(color.FgCyan).Sprint(timestamp), coloredMessage)
	l.writeToFile("ERROR", message)
}

// Debug logs debug messages (only if verbose is enabled)
func (l *Logger) Debug(message string) {
	if !l.verbose {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	coloredMessage := color.New(color.FgMagenta).Sprintf("[DEBUG] %s", message)

	fmt.Printf("%s %s\n", color.New(color.FgCyan).Sprint(timestamp), coloredMessage)
	l.writeToFile("DEBUG", message)
}

// Progress shows a progress message with emoji
func (l *Logger) Progress(step int, total int, message string) {
	timestamp := time.Now().Format("15:04:05")
	progressMsg := fmt.Sprintf("[%d/%d] %s", step, total, message)
	coloredMessage := color.New(color.FgWhite).Sprint(progressMsg)

	fmt.Printf("%s %s\n", color.New(color.FgCyan).Sprint(timestamp), coloredMessage)
	l.writeToFile("PROGRESS", progressMsg)
}

// WriteToFile writes to log file if available
func (l *Logger) writeToFile(level, message string) {
	if l.logFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logEntry := fmt.Sprintf("%s [%s] %s\n", timestamp, level, message)
		l.logFile.WriteString(logEntry)
	}
}

// Header prints a formatted header
func (l *Logger) Header(title string) {
	fmt.Println()
	fmt.Println(color.New(color.FgWhite, color.Bold).Sprint("================================"))
	fmt.Println(color.New(color.FgWhite, color.Bold).Sprint(title))
	fmt.Println(color.New(color.FgWhite, color.Bold).Sprint("================================"))
	l.writeToFile("HEADER", title)
}

// Separator prints a visual separator
func (l *Logger) Separator() {
	fmt.Println(color.New(color.FgWhite).Sprint("--------------------------------"))
}
