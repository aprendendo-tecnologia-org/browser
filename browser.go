package browser

import (
	"context"
	"fmt"
	//"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
)

type StatusCode int

// Browser represents a high-level browser abstraction over chromedp
type Browser struct {
	document   string
	currentURL string
	statusCode StatusCode
	timeout    time.Duration
}

// NewBrowser creates a new Browser instance
func NewBrowser() *Browser {
	return &Browser{
		timeout: 3 * time.Second,
	}
}

// NewBrowserWithOptions creates a new Browser with custom options
func NewBrowserWithOptions(headless bool, timeout time.Duration) *Browser {
	// opts := append(chromedp.DefaultExecAllocatorOptions[:],
	// 	chromedp.Flag("headless", headless),
	// 	chromedp.Flag("disable-gpu", true),
	// 	chromedp.Flag("no-sandbox", true),
	// 	chromedp.Flag("disable-dev-shm-usage", true),
	// )
	
	//allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	//ctx, cancel := chromedp.NewContext(allocCtx)
	
	return &Browser{
		//ctx:     ctx,
		//cancel:  cancel,
		timeout: timeout,
	}
}

// Visit navigates to the given URL and extracts the HTML document
func (b *Browser) Visit(ctx context.Context, url string) (StatusCode, error) {
	
	b.listenForNetworkEvent(ctx)

	// Navigate to URL and get HTML
	var htmlContent string
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	
	if err != nil {
		return 500, fmt.Errorf("failed to visit %s: %w", url, err)
	}
	
	// Store the document and current URL
	b.document = htmlContent
	b.currentURL = url
	
	return b.statusCode, nil
}

// GetDocument returns the current HTML document
func (b *Browser) GetDocument() string {
	return b.document
}

// GetCurrentURL returns the current URL
func (b *Browser) GetCurrentURL() string {
	return b.currentURL
}

// WaitForElement waits for an element to be visible and updates the document
func (b *Browser) WaitForElement(ctx context.Context, selector string) error {
	
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	
	if err != nil {
		return fmt.Errorf("failed to wait for element %s: %w", selector, err)
	}
	
	b.document = htmlContent
	return nil
}

func (b *Browser) GetNodes(ctx context.Context, selector string) ([]*cdp.Node, error) {
	var nodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find elements with selector %s: %w", selector, err)
	}
	return nodes, nil
}

// Click clicks on an element and updates the document
func (b *Browser) Click(ctx context.Context, selector string) error {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()
	
	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Click(selector, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second), // Wait for potential changes
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	
	if err != nil {
		return fmt.Errorf("failed to click element %s: %w", selector, err)
	}
	
	b.document = htmlContent
	return nil
}

// Exec extracts DOM nodes matching the selector and passes them to the function
func (b *Browser) Exec(ctx context.Context, selector string, f func([]*cdp.Node)) error {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()
	
	var nodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll),
	)
	
	if err != nil {
		return fmt.Errorf("failed to find elements with selector %s: %w", selector, err)
	}
	
	// Call the provided function with the nodes
	f(nodes)
	
	return nil
}

// ExecFirst is a convenience method that executes the function with only the first matching node
func (b *Browser) ExecFirst(ctx context.Context, selector string, f func(*cdp.Node)) error {
	return b.Exec(ctx, selector, func(nodes []*cdp.Node) {
		if len(nodes) > 0 {
			f(nodes[0])
		}
	})
}

// SetTimeout sets the timeout for browser operations
func (b *Browser) SetTimeout(timeout time.Duration) {
	b.timeout = timeout
}

// Example usage
// func main() {
// 	// Create a new browser instance
// 	browser := NewBrowser()
// 	defer browser.Close()
	
// 	// Visit a website
// 	err := browser.Visit("https://example.com")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
	
// 	// Get the HTML document
// 	doc := browser.GetDocument()
// 	fmt.Printf("Document length: %d\n", len(doc))
// 	fmt.Printf("Current URL: %s\n", browser.GetCurrentURL())
	
// 	// Use Exec to extract and work with DOM nodes
// 	err = browser.Exec("a", func(nodes []*cdp.Node) {
// 		fmt.Printf("Found %d anchor elements:\n", len(nodes))
// 		for i, node := range nodes {
// 			fmt.Printf("  Link #%d:\n", i+1)
// 			fmt.Printf("    Node Name: %s\n", node.NodeName)
// 			fmt.Printf("    Node ID: %d\n", node.NodeID)
			
// 			// Access attributes
// 			if node.Attributes != nil {
// 				fmt.Printf("    Attributes:\n")
// 				for j := 0; j < len(node.Attributes); j += 2 {
// 					if j+1 < len(node.Attributes) {
// 						fmt.Printf("      %s: %s\n", node.Attributes[j], node.Attributes[j+1])
// 					}
// 				}
// 			}
// 		}
// 	})
// 	if err != nil {
// 		log.Printf("Exec error: %v", err)
// 	}
	
// 	// Use ExecFirst for single node
// 	err = browser.ExecFirst("title", func(node *cdp.Node) {
// 		fmt.Printf("Title node: %s (ID: %d)\n", node.NodeName, node.NodeID)
// 	})
// 	if err != nil {
// 		log.Printf("ExecFirst error: %v", err)
// 	}
	
// 	// Example: Process all h1 nodes
// 	err = browser.Exec("h1", func(nodes []*cdp.Node) {
// 		fmt.Printf("Found %d h1 elements:\n", len(nodes))
// 		for _, node := range nodes {
// 			fmt.Printf("  H1 Node ID: %d\n", node.NodeID)
// 			if node.NodeValue != "" {
// 				fmt.Printf("  H1 Value: %s\n", node.NodeValue)
// 			}
// 		}
// 	})
// 	if err != nil {
// 		log.Printf("H1 extraction error: %v", err)
// 	}
	
// 	// Wait for a specific element (if it exists)
// 	err = browser.WaitForElement("h1")
// 	if err != nil {
// 		log.Printf("Warning: %v", err)
// 	}
	
// 	// Take a screenshot
// 	screenshot, err := browser.Screenshot()
// 	if err != nil {
// 		log.Printf("Screenshot error: %v", err)
// 	} else {
// 		fmt.Printf("Screenshot size: %d bytes\n", len(screenshot))
// 	}
// }
func (b *Browser) listenForNetworkEvent(ctx context.Context) {
    chromedp.ListenTarget(ctx, func(ev interface{}) {
        switch ev := ev.(type) {
        case *network.EventResponseReceived:
            resp := ev.Response
            b.statusCode = StatusCode(resp.Status)
			fmt.Printf("Response received: URL=%s, Status=%d\n", resp.URL, resp.Status)
        }
    })
}