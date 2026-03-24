package main

import (
	"flag"
	"fmt"
	"os"
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
