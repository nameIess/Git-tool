package clipboard

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/nameIess/git-tool/internal/logger"
)

// Copy copies the given text to the system clipboard.
func Copy(text string) error {
	logger.Debug("Attempting to copy %d bytes to clipboard", len(text))
	err := clipboard.WriteAll(text)
	if err != nil {
		logger.Error("Clipboard copy failed: %v", err)
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	logger.Info("Successfully copied text to clipboard")
	return nil
}
