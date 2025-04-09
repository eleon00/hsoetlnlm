package benthos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ExecuteBenthosPipeline runs a Benthos pipeline using the provided YAML configuration.
// It executes the 'rpk connect run' command, passing the config via standard input.
// It returns the combined stdout/stderr output and any execution error.
func ExecuteBenthosPipeline(ctx context.Context, configYAML string) (string, error) {
	// Check if rpk executable exists in PATH
	rpkPath, err := exec.LookPath("rpk")
	if err != nil {
		return "", fmt.Errorf("rpk command not found in PATH: %w", err)
	}

	// Create the command with context for potential cancellation/timeout
	// We pass the config via stdin using the '-c -' flags with 'rpk connect run'.
	cmd := exec.CommandContext(ctx, rpkPath, "connect", "run", "-c", "-")

	// Set the standard input to the configuration string
	cmd.Stdin = bytes.NewBufferString(configYAML)

	// Capture combined stdout and stderr
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	// Execute the command
	startTime := time.Now()
	err = cmd.Run()
	duration := time.Since(startTime)

	outputStr := output.String()

	if err != nil {
		// If the command failed, include the output in the error message
		// as Benthos often prints useful error details to stderr.
		return outputStr, fmt.Errorf("rpk connect run execution failed after %v: %w\nOutput:\n%s", duration, err, outputStr)
	}

	// Optionally, log the successful execution duration and output snippet
	// logger.Info().Dur("duration", duration).Msg("rpk connect run execution successful")

	return outputStr, nil
}

// Helper function to truncate strings for logging (optional)
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
