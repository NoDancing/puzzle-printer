package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/seancohan/puzzle-printer/internal/config"
	"github.com/seancohan/puzzle-printer/internal/nyt"
	"github.com/seancohan/puzzle-printer/internal/print"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dateStr  = flag.String("date", "", "Fetch puzzle for this date (YYYY-MM-DD); defaults to today")
		output   = flag.String("output", "", "Save PDF to this path instead of printing")
		noprint  = flag.Bool("no-print", false, "Open PDF in Preview instead of sending to printer")
	)
	flag.Parse()

	// Resolve date
	date := time.Now()
	if *dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", *dateStr)
		if err != nil {
			return fmt.Errorf("invalid date %q (want YYYY-MM-DD)", *dateStr)
		}
	}

	// Load config / credentials
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Authenticate and fetch
	fmt.Printf("Fetching NYT crossword for %s...\n", date.Format("Monday, January 2 2006"))
	client, err := nyt.NewClient(cfg.NYT.Email, cfg.NYT.Password, cfg.NYT.Token)
	if err != nil {
		return err
	}

	pdf, err := client.FetchPDF(date)
	if err != nil {
		return err
	}

	// Determine output path
	pdfPath := *output
	if pdfPath == "" {
		f, err := os.CreateTemp("", "crossword-*.pdf")
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		pdfPath = f.Name()
		f.Close()
	}

	if err := os.WriteFile(pdfPath, pdf, 0644); err != nil {
		return fmt.Errorf("writing PDF: %w", err)
	}

	// Upload to remote server if configured
	if cfg.Upload.Host != "" && cfg.Upload.KeyPath != "" {
		remotePath := cfg.Upload.RemotePath
		if remotePath == "" {
			remotePath = "~/puzzle-site/pdfs/"
		}
		remoteFile := fmt.Sprintf("crossword-%s.pdf", date.Format("2006-01-02"))
		dest := cfg.Upload.Host + ":" + remotePath + remoteFile

		// Copy key to a temp file with strict permissions (0600) so scp accepts it,
		// since the mounted file may have overly broad permissions.
		keyData, err := os.ReadFile(cfg.Upload.KeyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: reading upload key: %v\n", err)
		} else {
			tmpKey, err := os.CreateTemp("", "scp-key-*")
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: creating temp key: %v\n", err)
			} else {
				tmpKeyPath := tmpKey.Name()
				defer os.Remove(tmpKeyPath)
				if err := tmpKey.Chmod(0600); err == nil {
					if _, err := tmpKey.Write(keyData); err == nil {
						tmpKey.Close()
						cmd := exec.Command("scp",
							"-i", tmpKeyPath,
							"-o", "StrictHostKeyChecking=no",
							pdfPath, dest,
						)
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						fmt.Printf("Uploading to %s...\n", dest)
						if err := cmd.Run(); err != nil {
							fmt.Fprintf(os.Stderr, "warning: upload failed: %v\n", err)
						} else {
							fmt.Println("Upload complete.")
						}
					}
				}
				tmpKey.Close()
			}
		}
	}

	// Print or open
	if *output != "" {
		fmt.Printf("Saved to %s\n", pdfPath)
		return nil
	}
	if *noprint {
		fmt.Printf("Opening in Preview: %s\n", pdfPath)
		return print.ToPreview(pdfPath)
	}

	fmt.Println("Sending to printer...")
	return print.ToPrinter(pdfPath, cfg.Print.Printer)
}
