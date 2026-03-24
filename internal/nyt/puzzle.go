package nyt

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	dailyMetaURL  = "https://www.nytimes.com/svc/crosswords/v6/puzzle/daily.json"
	puzzlePDFURL  = "https://www.nytimes.com/svc/crosswords/v2/puzzle/%d.pdf"
	printPDFURL   = "https://www.nytimes.com/svc/crosswords/v2/puzzle/print/%s.pdf"
)

// FetchPDF returns the PDF bytes for the crossword on the given date.
// On Sundays, it fetches the print edition; other days fetch by puzzle ID.
func (c *Client) FetchPDF(date time.Time) ([]byte, error) {
	if date.Weekday() == time.Sunday {
		return c.fetchSundayPrintPDF(date)
	}
	return c.fetchWeekdayPDF(date)
}

func (c *Client) fetchSundayPrintPDF(date time.Time) ([]byte, error) {
	// Format: Mon0226 → "Mar2226" for March 22 2026
	dateStr := date.Format("Jan") + fmt.Sprintf("%02d%02d", date.Day(), date.Year()%100)
	pdfURL := fmt.Sprintf(printPDFURL, dateStr)
	data, err := c.get(pdfURL)
	if err != nil {
		return nil, fmt.Errorf("fetching Sunday print PDF: %w", err)
	}
	return data, nil
}

func (c *Client) fetchWeekdayPDF(date time.Time) ([]byte, error) {
	id, err := c.fetchPuzzleID(date)
	if err != nil {
		return nil, err
	}
	pdfURL := fmt.Sprintf(puzzlePDFURL, id)
	data, err := c.get(pdfURL)
	if err != nil {
		return nil, fmt.Errorf("fetching puzzle PDF (id=%d): %w", id, err)
	}
	return data, nil
}

func (c *Client) fetchPuzzleID(date time.Time) (int, error) {
	metaURL := dailyMetaURL
	if !isToday(date) {
		metaURL = fmt.Sprintf(
			"https://www.nytimes.com/svc/crosswords/v6/puzzle/daily-%s.json",
			date.Format("2006-01-02"),
		)
	}

	body, err := c.get(metaURL)
	if err != nil {
		return 0, fmt.Errorf("fetching puzzle metadata: %w", err)
	}

	var result struct {
		Results []struct {
			PuzzleID int `json:"puzzle_id"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parsing puzzle metadata: %w", err)
	}
	if len(result.Results) == 0 || result.Results[0].PuzzleID == 0 {
		return 0, fmt.Errorf("no puzzle ID found in metadata response")
	}
	return result.Results[0].PuzzleID, nil
}

func isToday(date time.Time) bool {
	now := time.Now()
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
