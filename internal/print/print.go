package print

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	ipp "github.com/phin1x/go-ipp"
)

// ToPrinter sends pdfPath to the named printer (empty = system default).
// If printer is an IPP URI, submits directly via the IPP protocol.
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
	u, err := url.Parse(printerURI)
	if err != nil {
		return fmt.Errorf("parsing printer URI: %w", err)
	}

	port := 631
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	client := ipp.NewIPPClient(u.Hostname(), port, "", "", u.Scheme == "ipps")

	f, err := os.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("opening PDF: %w", err)
	}
	defer f.Close()

	doc := ipp.Document{
		Document: f,
		Name:     "crossword.pdf",
		MimeType: "application/pdf",
		Size:     -1,
	}

	jobID, err := client.PrintJob(doc, u.Path, nil)
	if err != nil {
		return fmt.Errorf("IPP print job: %w", err)
	}
	fmt.Printf("Print job submitted (id=%d)\n", jobID)
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
