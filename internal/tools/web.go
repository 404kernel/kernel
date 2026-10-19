package tools

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

var (
	titleRegex   = regexp.MustCompile(`<a[^>]+class="result__url"[^>]+href="([^"]+)"[^>]*>[\s\S]*?<a[^>]+class="result__snippet"[^>]*>([\s\S]*?)</a>`)
	snippetRegex = regexp.MustCompile(`<a[^>]+class="result__snippet"[^>]*>([\s\S]*?)</a>`)
	linkRegex    = regexp.MustCompile(`<a[^>]+class="result__url"[^>]+href="([^"]+)"`)
	stripTags    = regexp.MustCompile(`<[^>]+>`)
)

// SearchWeb queries DuckDuckGo HTML without requiring any API keys
func SearchWeb(ctx context.Context, query string) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := string(body)
	var results []SearchResult

	// Extract snippets and links
	snippets := snippetRegex.FindAllStringSubmatch(bodyStr, 5)
	links := linkRegex.FindAllStringSubmatch(bodyStr, 5)

	maxItems := len(snippets)
	if len(links) < maxItems {
		maxItems = len(links)
	}

	for i := 0; i < maxItems; i++ {
		cleanSnippet := html.UnescapeString(stripTags.ReplaceAllString(snippets[i][1], ""))
		cleanSnippet = strings.TrimSpace(cleanSnippet)

		rawLink := links[i][1]
		if strings.Contains(rawLink, "uddg=") {
			// Extract destination URL from duckduckgo redirect link
			if u, err := url.Parse(rawLink); err == nil {
				if dest := u.Query().Get("uddg"); dest != "" {
					rawLink = dest
				}
			}
		}

		if cleanSnippet != "" {
			results = append(results, SearchResult{
				Title:   fmt.Sprintf("Result #%d", i+1),
				URL:     rawLink,
				Snippet: cleanSnippet,
			})
		}
	}

	return results, nil
}

// FormatSearchResults renders search results in clean Markdown
func FormatSearchResults(query string, results []SearchResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No web results found for: `%s`", query)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("### 🌐 Web Search Results for `%s`\n\n", query))
	for i, r := range results {
		b.WriteString(fmt.Sprintf("**%d. %s**\n%s\n> *%s*\n\n", i+1, r.URL, r.Snippet, r.URL))
	}
	return strings.TrimSpace(b.String())
}
