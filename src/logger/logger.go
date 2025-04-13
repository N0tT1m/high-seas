package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"high-seas/src/utils"
)

type LogSeverity int

const (
	DEBUG   LogSeverity = iota
	ERROR   LogSeverity = iota
	WARNING LogSeverity = iota
	INFO    LogSeverity = iota
)

// Configuration for remote logging
var (
	remoteEnabled = utils.EnvVar("REMOTE_LOGGING_ENABLED", "false") == "true"
	sumoURL       = utils.EnvVar("SUMO_COLLECTOR_URL", "")
	logBuffer     = make([]string, 0, 100)
	bufferMutex   sync.Mutex
	flushInterval = 5 * time.Second
)

// SumoLogicHook implements logrus.Hook interface
type SumoLogicHook struct {
	levels []logrus.Level
}

// Ensure SumoLogicHook implements the logrus.Hook interface
var _ logrus.Hook = &SumoLogicHook{}

// Initialize logger with remote logging capability
func init() {
	// Set up log format
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Set output to both stdout and file
	f, err := os.OpenFile("high-seas.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, f)
		logrus.SetOutput(mw)
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Warnf("Failed to open log file: %v", err)
	}

	// Set up remote logging if enabled
	if remoteEnabled {
		if sumoURL == "" {
			logrus.Warn("Remote logging enabled but SUMO_COLLECTOR_URL not set")
		} else {
			// Add SumoLogic hook
			hook := &SumoLogicHook{
				levels: []logrus.Level{
					logrus.PanicLevel,
					logrus.FatalLevel,
					logrus.ErrorLevel,
					logrus.WarnLevel,
					logrus.InfoLevel,
					logrus.DebugLevel,
				},
			}
			logrus.AddHook(hook)

			// Start background goroutine to flush logs
			go flushLogsRoutine()
		}
	}
}

// GetSeverityString returns the string representation of log severity
func GetSeverityString(logType LogSeverity) string {
	switch logType {
	case DEBUG:
		return "DEBUG"
	case ERROR:
		return "ERROR"
	case WARNING:
		return "WARNING"
	case INFO:
		return "INFO"
	default:
		return "INFO"
	}
}

// Maps our severity to logrus levels
func getSeverityLevel(logType LogSeverity) logrus.Level {
	switch logType {
	case DEBUG:
		return logrus.DebugLevel
	case ERROR:
		return logrus.ErrorLevel
	case WARNING:
		return logrus.WarnLevel
	case INFO:
		return logrus.InfoLevel
	default:
		return logrus.InfoLevel
	}
}

// write logs a message with the specified severity
func write(logType LogSeverity, message interface{}) {
	entry := logrus.WithField("component", "high-seas")

	switch getSeverityLevel(logType) {
	case logrus.ErrorLevel:
		entry.Error(message)
	case logrus.WarnLevel:
		entry.Warn(message)
	case logrus.DebugLevel:
		entry.Debug(message)
	default:
		entry.Info(message)
	}
}

// WriteError logs an error message
func WriteError(message string, err error) {
	write(ERROR, fmt.Sprintf("%s %v", message, err))
}

// WriteWarning logs a warning message
func WriteWarning(message string) {
	write(WARNING, message)
}

// WriteInfo logs an info message
func WriteInfo(message interface{}) {
	write(INFO, message)
}

// WriteFatal logs a fatal error and terminates the program
func WriteFatal(errMsg string, err error) {
	logrus.WithFields(logrus.Fields{
		"reason_for_error": errMsg,
		"error":            err,
	}).Fatal("Application terminating due to fatal error")
}

// WriteCMDInfo logs command execution information
func WriteCMDInfo(cmd string, output string) {
	logrus.WithFields(logrus.Fields{
		"command":        cmd,
		"command_output": output,
	}).Info("Command executed")
}

// Implementation of logrus.Hook interface methods for SumoLogicHook

// Levels returns the logging levels this hook is enabled for
func (hook *SumoLogicHook) Levels() []logrus.Level {
	return hook.levels
}

// Fire is called when a log event occurs
func (hook *SumoLogicHook) Fire(entry *logrus.Entry) error {
	if !remoteEnabled || sumoURL == "" {
		return nil
	}

	// Format the log entry as JSON
	line, err := entry.String()
	if err != nil {
		return err
	}

	// Add to buffer
	bufferMutex.Lock()
	logBuffer = append(logBuffer, line)
	bufferMutex.Unlock()

	// If buffer gets too large, flush immediately
	if len(logBuffer) >= 100 {
		go flushLogs()
	}

	return nil
}

// flushLogs sends buffered logs to SumoLogic
func flushLogs() {
	bufferMutex.Lock()
	if len(logBuffer) == 0 {
		bufferMutex.Unlock()
		return
	}

	// Get logs and clear buffer
	logsToSend := make([]string, len(logBuffer))
	copy(logsToSend, logBuffer)
	logBuffer = logBuffer[:0]
	bufferMutex.Unlock()

	// Send logs to SumoLogic
	payload := strings.Join(logsToSend, "")
	err := sendToSumoLogic(payload)
	if err != nil {
		logrus.Errorf("Failed to send logs to SumoLogic: %v", err)
	}
}

// flushLogsRoutine periodically flushes logs to SumoLogic
func flushLogsRoutine() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		flushLogs()
	}
}

// sendToSumoLogic sends logs to SumoLogic HTTP source
func sendToSumoLogic(payload string) error {
	if sumoURL == "" {
		return fmt.Errorf("SumoLogic URL not configured")
	}

	// Create an HTTP client with a timeout
	client := utils.CreateHTTPClient()

	// Create request
	req, err := utils.CreatePostRequest(sumoURL, "application/json", strings.NewReader(payload))
	if err != nil {
		return err
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 300 {
		return fmt.Errorf("SumoLogic responded with status: %d", resp.StatusCode)
	}

	return nil
}
