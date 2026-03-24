package print

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ToPrinter sends pdfPath to the named printer (empty = system default).
// If printer is an IPP URI (ipp://...), submits directly via curl.
// Otherwise passes it as a CUPS destination name to lp.
func ToPrinter(pdfPath, printer string) error {
	if strings.HasPrefix(printer, "ipp://") || strings.HasPrefix(printer, "ipps://") {
		cmd := exec.Command("curl", "-s", "-S",
			"-X", "POST", printer,
			"-H", "Content-Type: application/pdf",
			"--data-binary", "@"+pdfPath,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("curl print: %w", err)
		}
		return nil
	}

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
