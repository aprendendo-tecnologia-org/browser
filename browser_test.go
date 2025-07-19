package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func TestBrowser_GetNodes(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	tests := []struct {
		name        string
		html        string
		selector    string
		wantCount   int
		wantContent string
	}{
		{
			name: "Simple single node",
			html: `
                <!DOCTYPE html>
                <html>
                <body>
                    <div id="unique">Hello Node</div>
                </body>
                </html>
            `,
			selector:    "#unique",
			wantCount:   1,
			wantContent: "Hello Node",
		},
		{
			name: "Multiple product items",
			html: `
                <!DOCTYPE html>
                <html>
                <body>
                    <ul>
                        <li class="product">Produto A</li>
                        <li class="product">Produto B</li>
                        <li class="product">Produto C</li>
                    </ul>
                </body>
                </html>
            `,
			selector:    ".product",
			wantCount:   3,
			wantContent: "Produto B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, tt.html)
			}))
			defer server.Close()

			b := NewBrowser()
			_, err := b.Visit(ctx, server.URL)
			if err != nil {
				t.Fatalf("Visit() error = %v", err)
			}

			nodes, err := b.GetNodes(ctx, tt.selector)
			if err != nil {
				t.Fatalf("GetNodes() error = %v", err)
			}
			if len(nodes) != tt.wantCount {
				t.Errorf("GetNodes() count = %d, want %d", len(nodes), tt.wantCount)
			}
			// Check if any node contains the expected content
			found := false
			for _, node := range nodes {
				if node.NodeValue == tt.wantContent {
					found = true
					break
				}
				// Sometimes NodeValue is empty and text is in children
				for _, child := range node.Children {
					if child.NodeValue == tt.wantContent {
						found = true
						break
					}
				}
			}
			if tt.wantContent != "" && !found {
				t.Errorf("GetNodes() did not find node with content %q", tt.wantContent)
			}
		})
	}
}

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

func TestBrowser_Click_LongOperationCreatesElement(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// HTML with a button and a container div. Clicking the button waits 1s and adds a new element.
	const html = `
        <!DOCTYPE html>
        <html>
        <body>
            <button id="createBtn" onclick="setTimeout(function() {
                var el = document.createElement('span');
                el.id = 'created';
                el.textContent = 'Elemento criado!';
                document.getElementById('container').appendChild(el);
            }, 1200);">Criar elemento</button>
            <div id="container"></div>
        </body>
        </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}))
	defer server.Close()

	b := NewBrowser()
	_, err := b.Visit(ctx, server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	// Click the button that triggers the long operation
	err = b.Click(ctx, "#createBtn")
	if err != nil {
		t.Fatalf("Click() error = %v", err)
	}

	// Wait for the new element to appear in the DOM
	err = b.WaitForElement(ctx, "#created")
	if err != nil {
		t.Fatalf("WaitForElement() error = %v", err)
	}

	doc := b.GetDocument()
	if !contains(doc, "Elemento criado!") {
		t.Errorf("Expected created element text not found in document")
	}
}

func TestBrowser_Exec_ExtractLinks(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	const html = `
        <!DOCTYPE html>
        <html>
        <body>
            <a href="https://site1.com" class="external">Site 1</a>
            <a href="https://site2.com" class="external">Site 2</a>
            <a href="/internal" class="internal">Internal</a>
        </body>
        </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}))
	defer server.Close()

	b := NewBrowser()
	_, err := b.Visit(ctx, server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	var foundLinks []string
	err = b.Exec(ctx, "a.external", func(nodes []*cdp.Node) {
		for _, node := range nodes {
			for i := 0; i < len(node.Attributes); i += 2 {
				if node.Attributes[i] == "href" {
					foundLinks = append(foundLinks, node.Attributes[i+1])
				}
			}
		}
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	expected := []string{"https://site1.com", "https://site2.com"}
	if len(foundLinks) != len(expected) {
		t.Errorf("Expected %d links, got %d", len(expected), len(foundLinks))
	}
	for _, want := range expected {
		found := false
		for _, got := range foundLinks {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected link %q not found in results: %v", want, foundLinks)
		}
	}
}
