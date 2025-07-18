package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chromedp/chromedp"
)

var browser = NewBrowser()

func TestBrowser_Navigate(t *testing.T) {

	ctx  := context.Background()

	type args struct {
		url string
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<head><title>%s</title></head>
			<body><a id="foo" href="/foo">foo</a></body>
		`, r.URL.Path)
	}))
	defer ts.Close()

	tests := []struct {
		name    string
		browser *Browser
		url     string
	}{
		{"Must visit " + ts.URL, browser, ts.URL},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.browser.Visit(ctx, tt.url)
			if err != nil {
				t.Errorf("Browser.Navigate(%s) error = %v", tt.url, err)
				return
			}
		})
	}
}

func TestBrowser_GetDocument(t *testing.T) {

	ctx, cancel  := chromedp.NewContext(context.Background())
	defer cancel()

	tests := []struct {
		name   		string
		browser		*Browser
		selector 	string
		want   		string
	}{
		{"Get foo", browser, `#name`, "placa de video"},// TODO: Add test cases.
	}

	server := serve("placa-video.html")
	defer server.Close()
	
	for _, tt := range tests {
		
		t.Run(tt.name, func(t *testing.T) {	
			err := tt.browser.Visit(ctx, server.URL)
			if err != nil {
				t.Errorf("Browser.Visit( %s ) = %v", server.URL, err)
			}

			err = tt.browser.WaitForElement(ctx, tt.selector)
			if err != nil {
				t.Errorf("Browser.WaitForElement( %s ) = %v", tt.selector, err)
			}		
			
			if got := tt.browser.GetDocument(); got != tt.want {
				t.Errorf("Browser.GetDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBrowser_GetNodes(t *testing.T) {

	ctx, cancel  := chromedp.NewContext(context.Background())
	defer cancel()

	tests := []struct {
		name   		string
		browser		*Browser
		selector 	string
		want   		string
	}{
		{"Get foo", browser, `a.product-detail`, "placa de video"},// TODO: Add test cases.
	}

	server := serve("placa-video.html")
	defer server.Close()
	
	for _, tt := range tests {
		
		t.Run(tt.name, func(t *testing.T) {	
			err := tt.browser.Visit(ctx, server.URL)
			if err != nil {
				t.Errorf("Browser.Visit( %s ) = %v", server.URL, err)
			}

			nodes, err := tt.browser.GetNodes(ctx, tt.selector)
			if err != nil {
				t.Errorf("Browser.WaitForElement( %s ) = %v", tt.selector, err)
			}

			if len(nodes) == 0 {
				t.Errorf("Browser.GetNodes() = %v, want at least one node", nodes)
			}
			fmt.Printf("First node text: %s", nodes[0].Children[0].NodeValue)
		})
	}
}

func serve(url string) *httptest.Server{

	output, err := getHTMLFromFile(url)
	if err != nil {
		fmt.Printf("Error reading HTML file: %v\n", err)
		return nil
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, output, r.URL.Path)
		fmt.Println("server inicializado")
	}))	
	return ts
}

func getHTMLFromFile(file string) (string, error) {
	if file == "" {
		return file, fmt.Errorf("Nome do arquivo não pode ser vazio")
	}

	fileContent, err := os.ReadFile("testdata/" + file)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}
