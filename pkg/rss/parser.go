package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"GoNews/pkg/model"
)

// RSS структура для разбора XML
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
}

// Parse обрабатывает RSS-ленту
func Parse(url string) ([]model.Post, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("XML unmarshal failed: %w", err)
	}

	var posts []model.Post
	for _, item := range rss.Channel.Items {
		pubTime, err := parseDate(item.PubDate)
		if err != nil {
			pubTime = time.Now()
		}

		posts = append(posts, model.Post{
			Title:   strings.TrimSpace(item.Title),
			Content: strings.TrimSpace(item.Description),
			PubTime: pubTime.Unix(),
			Link:    strings.TrimSpace(item.Link),
		})
	}
	return posts, nil
}

// parseDate обрабатывает различные форматы дат в RSS
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC3339,
		"Mon, 2 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", dateStr)
}
