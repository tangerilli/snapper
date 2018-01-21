// TODO
// Communicate with chrome (handle passed in from whatever is running things) to print a PDF
// given a set of options
package snapper

import (
	"context"
	"encoding/base64"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
	"log"
	"time"
)

func connectToChrome(chromeUrl string, ctx context.Context, maxDataSize int) (*rpcc.Conn, error) {
	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New(chromeUrl)
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Initiate a new RPC connection to the Chrome Debugging Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL, rpcc.WithWriteBufferSize(maxDataSize+1000), rpcc.WithCompression())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// TODO:
// - Need to be able to set options like total timeout, network request timeouts, print args, etc..
// - handle context timeout errors (i.e. requests greater than the number set in WithTimeout)
func PrintPdfFromHtml(chromeUrl string, html string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := connectToChrome(chromeUrl, ctx, len(html))
	if err != nil {
		log.Println("Error connecting to chrome")
		return nil, err
	}

	defer conn.Close() // Leaving connections open will leak memory.
	c := cdp.NewClient(conn)
	log.Println("Connected")

	domContent, err := c.Page.LoadEventFired(ctx)
	if err != nil {
		log.Println("Error setting up DOM content event handler")
		return nil, err
	}
	defer domContent.Close()

	if err = c.Page.Enable(ctx); err != nil {
		log.Println("Error enabling page events")
		return nil, err
	}

	navArgs := page.NewNavigateArgs("about:blank")
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		log.Println("Error navigating to blank page")
		return nil, err
	}

	if _, err = domContent.Recv(); err != nil {
		log.Println("Error waiting for navigation")
		return nil, err
	}

	contentArgs := page.NewSetDocumentContentArgs(nav.FrameID, html)
	err = c.Page.SetDocumentContent(ctx, contentArgs)
	if err != nil {
		log.Println("Error setting document content: ", err)
		return nil, err
	}

	if _, err = domContent.Recv(); err != nil {
		log.Println("Error waiting for content to be set")
		return nil, err
	}

	// TODO: set print args
	pdfArgs := page.NewPrintToPDFArgs()
	result, err := c.Page.PrintToPDF(ctx, pdfArgs)

	encodedData := base64.StdEncoding.EncodeToString([]byte(result.Data))
	return []byte(encodedData), nil
}
