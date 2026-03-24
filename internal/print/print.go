package print

import (
	"fmt"
	"os"
	"os/exec"
)

// ToPrinter sends pdfPath to the named printer (empty = system default).
func ToPrinter(pdfPath, printer string) error {
	args := []string{}
	if printer != "" {
		args = append(args, "-d", printer)
	}
	args = append(args, pdfPath)

	cmd := exec.Command("lp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("lp: %w", err)
	}
	return nil
}

// ToPreview opens pdfPath in macOS Preview.
func ToPreview(pdfPath string) error {
	cmd := exec.Command("open", "-a", "Preview", pdfPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open Preview: %w", err)
	}
	return nil
}
