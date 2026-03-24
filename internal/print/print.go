package print

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ToPrinter sends pdfPath to the named printer (empty = system default).
// If printer is an IPP URI, converts PDF to URF and submits via IPP over HTTP.
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
	// Convert PDF to URF (AirPrint raster) — Brother printers don't accept PDF over IPP.
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

	docData, err := os.ReadFile(urfPath)
	if err != nil {
		return fmt.Errorf("reading URF file: %w", err)
	}

	ippReq := buildPrintJobRequest(printerURI, "crossword.pdf", "image/urf")
	body := append(ippReq, docData...)

	u, err := url.Parse(printerURI)
	if err != nil {
		return fmt.Errorf("parsing printer URI: %w", err)
	}
	port := 631
	if p := u.Port(); p != "" {
		port, _ = strconv.Atoi(p)
	}
	proto := "http"
	if u.Scheme == "ipps" {
		proto = "https"
	}
	httpURL := fmt.Sprintf("%s://%s:%d%s", proto, u.Hostname(), port, u.Path)

	resp, err := http.Post(httpURL, "application/ipp", bytes.NewReader(body)) //nolint:gosec
	if err != nil {
		return fmt.Errorf("sending IPP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IPP HTTP error: %s", resp.Status)
	}

	respBody, _ := io.ReadAll(resp.Body)
	if len(respBody) >= 4 {
		statusCode := int(binary.BigEndian.Uint16(respBody[2:4]))
		if statusCode != 0 {
			return fmt.Errorf("IPP error status: 0x%04x", statusCode)
		}
	}

	fmt.Println("Print job submitted successfully")
	return nil
}

// buildPrintJobRequest encodes a minimal IPP/1.1 Print-Job request.
func buildPrintJobRequest(printerURI, jobName, docFormat string) []byte {
	var b bytes.Buffer

	// IPP/1.1 header
	b.Write([]byte{0x01, 0x01})             // version 1.1
	b.Write([]byte{0x00, 0x02})             // operation: Print-Job
	b.Write([]byte{0x00, 0x00, 0x00, 0x01}) // request-id: 1

	// operation-attributes-tag
	b.WriteByte(0x01)
	writeAttr(&b, 0x47, "attributes-charset", "utf-8")
	writeAttr(&b, 0x48, "attributes-natural-language", "en")
	writeAttr(&b, 0x45, "printer-uri", printerURI)
	writeAttr(&b, 0x42, "requesting-user-name", "puzzle-printer")
	writeAttr(&b, 0x42, "job-name", jobName)
	writeAttr(&b, 0x49, "document-format", docFormat)

	// end-of-attributes
	b.WriteByte(0x03)

	return b.Bytes()
}

func writeAttr(b *bytes.Buffer, tag byte, name, value string) {
	b.WriteByte(tag)
	binary.Write(b, binary.BigEndian, int16(len(name)))  //nolint:errcheck
	b.WriteString(name)
	binary.Write(b, binary.BigEndian, int16(len(value))) //nolint:errcheck
	b.WriteString(value)
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
