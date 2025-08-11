package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Downloader struct {
	BaseURL    *url.URL
	Domain     string
	Visited    map[string]bool
	VisitedMux sync.Mutex
	HttpClient *http.Client
	OutputDir  string
	MaxDepth   int
	Jobs       chan Job
	Wg         sync.WaitGroup
}

type Job struct {
	URL   string
	Depth int
}

func (d *Downloader) saveFile(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (d *Downloader) alreadyVisited(u string) bool {
	d.VisitedMux.Lock()
	defer d.VisitedMux.Unlock()
	if d.Visited[u] {
		return true
	}
	d.Visited[u] = true
	return false
}

func (d *Downloader) worker() {
	for job := range d.Jobs {
		d.download(job.URL, job.Depth)
		d.Wg.Done()
	}
}

func (d *Downloader) enqueue(url string, depth int) {
	ext := strings.ToLower(filepath.Ext(url))
	mediaExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".svg": true,
		".css": true, ".js": true, ".woff": true, ".woff2": true, ".ttf": true,
	}

	if depth > d.MaxDepth && !mediaExtensions[ext] {
		return
	}
	if d.alreadyVisited(url) {
		return
	}

	d.Wg.Add(1)

	select {
	case d.Jobs <- Job{URL: url, Depth: depth}:
	default:
	}
}

func (d *Downloader) download(rawurl string, depth int) {
	u, err := url.Parse(rawurl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "⚠️ Ошибка разбора URL:", rawurl, "-", err)
		return
	}

	if u.Host != "" && u.Host != d.Domain {
		return
	}

	if !u.IsAbs() {
		u = d.BaseURL.ResolveReference(u)
	}

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := d.HttpClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "⚠️ Ошибка при запросе:", u.String(), "-", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "⚠️ Пропущено (код %d): %s\n", resp.StatusCode, u.String())
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "⚠️ Ошибка чтения ответа:", u.String(), "-", err)
		return
	}

	path := u.Path
	if strings.HasSuffix(path, "/") || path == "" {
		path = filepath.Join(path, "index.html")
	}
	localPath := filepath.Join(d.OutputDir, u.Host, path)

	if err := d.saveFile(localPath, body); err != nil {
		fmt.Fprintln(os.Stderr, "⚠️ Ошибка сохранения:", localPath, "-", err)
		return
	}

	fmt.Println("✅ Скачано:", u.String())

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		d.processHTML(body, u, depth)
	}
}

func (d *Downloader) processHTML(body []byte, base *url.URL, depth int) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		fmt.Fprintln(os.Stderr, "⚠️ Ошибка парсинга HTML:", base.String(), "-", err)
		return
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range []string{"href", "src"} {
				for _, a := range n.Attr {
					if a.Key == attr {
						link := strings.TrimSpace(a.Val)
						if link == "" ||
							strings.HasPrefix(link, "data:") ||
							strings.HasPrefix(link, "javascript:") {
							continue
						}
						absLink, err := base.Parse(link)
						if err != nil {
							continue
						}
						d.enqueue(absLink.String(), depth+1)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func main() {
	startURL := flag.String("url", "", "Начальный URL для загрузки")
	outputDir := flag.String("out", "output", "Папка для сохранения")
	depth := flag.Int("depth", 1, "Глубина рекурсии")
	workers := flag.Int("workers", 5, "Количество одновременных загрузок")
	flag.Parse()

	if *startURL == "" {
		fmt.Println("❌ Укажите URL с помощью параметра -url")
		return
	}

	parsedURL, err := url.Parse(*startURL)
	if err != nil {
		fmt.Println("❌ Ошибка разбора URL:", err)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: tr,
	}

	d := &Downloader{
		BaseURL:    parsedURL,
		Domain:     parsedURL.Host,
		Visited:    make(map[string]bool),
		HttpClient: client,
		OutputDir:  *outputDir,
		MaxDepth:   *depth,
		Jobs:       make(chan Job, 100),
	}

	var workersWg sync.WaitGroup
	workersWg.Add(*workers)

	for i := 0; i < *workers; i++ {
		go func() {
			defer workersWg.Done()
			d.worker()
		}()
	}

	d.enqueue(parsedURL.String(), 0)

	go func() {
		d.Wg.Wait()
		close(d.Jobs)
	}()

	workersWg.Wait()
}
