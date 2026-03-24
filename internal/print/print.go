package print

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ToPrinter sends pdfPath to the named printer (empty = system default).
// If printer is an IPP URI, converts PDF to URF and submits via ipptool.
// Otherwise passes it as a CUPS destination name to lp.
func ToPrinter(pdfPath, printer string) error {
	if strings.HasPrefix(printer, "ipp://") || strings.HasPrefix(printer, "ipps://") {
		return printIPP(pdfPath, printer)
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

func printIPP(pdfPath, printerURI string) error {
	// Convert PDF to URF (AirPrint raster format) — Brother printers don't
	// accept PDF directly over IPP; they need URF or PWG Raster.
	urfPath := pdfPath + ".urf"
	defer os.Remove(urfPath)

	gs := exec.Command("gs",
		"-dBATCH", "-dNOPAUSE", "-dQUIET",
		"-sDEVICE=urfgray", "-r600",
		"-sOutputFile="+urfPath,
		pdfPath,
	)
	gs.Stdout = os.Stdout
	gs.Stderr = os.Stderr
	if err := gs.Run(); err != nil {
		return fmt.Errorf("converting PDF to URF: %w", err)
	}

	// Write a temporary ipptool test file
	testFile, err := os.CreateTemp("", "ipp-*.test")
	if err != nil {
		return fmt.Errorf("creating ipptool test file: %w", err)
	}
	defer os.Remove(testFile.Name())

	_, err = fmt.Fprintf(testFile, `{
    NAME "Print-Job"
    OPERATION Print-Job
    GROUP operation-attributes-tag
    ATTR charset attributes-charset utf-8
    ATTR language attributes-natural-language en
    ATTR uri printer-uri "%s"
    ATTR name requesting-user-name "puzzle-printer"
    ATTR name job-name "crossword.pdf"
    ATTR mimeMediaType document-format "image/urf"
    FILE %s
    STATUS successful-ok
}
`, printerURI, urfPath)
	testFile.Close()
	if err != nil {
		return fmt.Errorf("writing ipptool test file: %w", err)
	}

	cmd := exec.Command("ipptool", "-t", printerURI, testFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ipptool: %w", err)
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
