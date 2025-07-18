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

// func TestBrowser_GetDocument(t *testing.T) {

// 	ctx, cancel := chromedp.NewContext(context.Background())
// 	defer cancel()

// 	tests := []struct {
// 		name     string
// 		browser  *Browser
// 		selector string
// 		want     string
// 	}{
// 		{"Get foo", browser, `#name`, "placa de video"}, // TODO: Add test cases.
// 	}

// 	server := serve("placa-video.html")
// 	defer server.Close()

// 	for _, tt := range tests {

// 		t.Run(tt.name, func(t *testing.T) {
// 			err := tt.browser.Visit(ctx, server.URL)
// 			if err != nil {
// 				t.Errorf("Browser.Visit( %s ) = %v", server.URL, err)
// 			}

// 			err = tt.browser.WaitForElement(ctx, tt.selector)
// 			if err != nil {
// 				t.Errorf("Browser.WaitForElement( %s ) = %v", tt.selector, err)
// 			}

// 			if got := tt.browser.GetDocument(); got != tt.want {
// 				t.Errorf("Browser.GetDocument() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestBrowser_GetNodes(t *testing.T) {

// 	ctx, cancel := chromedp.NewContext(context.Background())
// 	defer cancel()

// 	tests := []struct {
// 		name     string
// 		browser  *Browser
// 		selector string
// 		want     string
// 	}{
// 		{"Get foo", browser, `a.product-detail`, "placa de video"}, // TODO: Add test cases.
// 	}

// 	server := serve("placa-video.html")
// 	defer server.Close()

// 	for _, tt := range tests {

// 		t.Run(tt.name, func(t *testing.T) {
// 			err := tt.browser.Visit(ctx, server.URL)
// 			if err != nil {
// 				t.Errorf("Browser.Visit( %s ) = %v", server.URL, err)
// 			}

// 			nodes, err := tt.browser.GetNodes(ctx, tt.selector)
// 			if err != nil {
// 				t.Errorf("Browser.WaitForElement( %s ) = %v", tt.selector, err)
// 			}

// 			if len(nodes) == 0 {
// 				t.Errorf("Browser.GetNodes() = %v, want at least one node", nodes)
// 			}
// 			fmt.Printf("First node text: %s", nodes[0].Children[0].NodeValue)
// 		})
// 	}
// }
func TestBrowser_Visit(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	tests := []struct {
		name        string
		browser     *Browser
		htmlFile    string
		wantStatus  StatusCode
		wantContent string
	}{
		{
			name:        "Visit valid page",
			browser:     NewBrowser(),
			htmlFile:    "placa-video.html",
			wantStatus:  200,
			wantContent: "placa de video", // Should be present in the HTML
		},
		{
			name:        "Visit missing page",
			browser:     NewBrowser(),
			htmlFile:    "missing.html",
			wantStatus:  404,
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.htmlFile == "missing.html" {
				// Simulate a server that returns 404
				server = httptest.NewServer(http.NotFoundHandler())
			} else {
				server = serve(tt.htmlFile)
			}
			defer server.Close()

			status, err := tt.browser.Visit(ctx, server.URL)
			if err != nil {
				t.Fatalf("Visit() error = %v, want nil", err)
			}
			if status != tt.wantStatus {
				t.Errorf("Visit() status = %v, want %v", status, tt.wantStatus)
			}
			doc := tt.browser.GetDocument()
			if tt.wantStatus == 200 && !contains(doc, tt.wantContent) {
				t.Errorf("Visit() document does not contain %q", tt.wantContent)
			}			
		})
	}
}

// contains checks if substr is in s.
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) > 0 && (s == substr || (len(s) > len(substr) && (s[0:len(substr)] == substr || contains(s[1:], substr)))))
}
func serve(url string) *httptest.Server {

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
