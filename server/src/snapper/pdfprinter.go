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
	devt := devtool.New(chromeUrl)
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return nil, err
		}
	}

	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL, rpcc.WithWriteBufferSize(maxDataSize+1000), rpcc.WithCompression())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func connectToChromeWithRetries(chromeUrl string, ctx context.Context, maxDataSize int) (*rpcc.Conn, error) {
	retries := 5
	sleepTimeInMs := 100

	var conn *rpcc.Conn
	for true {
		var err error
		conn, err = connectToChrome(chromeUrl, ctx, maxDataSize)

		if err == nil {
			break
		}

		if retries == 0 {
			log.Println("Error connecting to chrome")
			return nil, err
		}

		log.Printf("Error connecting to chrome, sleeping for %dms\n", sleepTimeInMs)
		time.Sleep(time.Duration(sleepTimeInMs) * time.Millisecond)

		retries--
		sleepTimeInMs *= 2
	}

	log.Println("Connected to Chrome")
	return conn, nil
}

func setupPdfArgs() *page.PrintToPDFArgs {
	// TODO: set print args properly
	pdfArgs := page.NewPrintToPDFArgs()
	pdfArgs.SetPrintBackground(true)
	pdfArgs.SetMarginTop(0)
	pdfArgs.SetMarginBottom(0)
	pdfArgs.SetMarginRight(0)
	pdfArgs.SetMarginLeft(0)
	return pdfArgs
}

type commandRunner func(client *cdp.Client, ctx context.Context, domContentEvent page.LoadEventFiredClient) (interface{}, error)

func runCommandsInChrome(chromeUrl string, maxDataSize int, options *Options, commandFunction commandRunner) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*options.Timeout)*time.Second)
	defer cancel()

	conn, err := connectToChromeWithRetries(chromeUrl, ctx, maxDataSize)
	if err != nil {
		return nil, err
	}

	defer conn.Close() // Leaving connections open will leak memory.
	c := cdp.NewClient(conn)

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

	return commandFunction(c, ctx, domContent)
}

func generatePdf(chromeUrl string, maxDataSize int, options *Options, pageSetupFunction commandRunner) ([]byte, error) {
	result, err := runCommandsInChrome(chromeUrl, maxDataSize, options, func(c *cdp.Client, ctx context.Context, domContent page.LoadEventFiredClient) (interface{}, error) {
		pageSetupFunction(c, ctx, domContent)

		pdfArgs := setupPdfArgs()
		result, err := c.Page.PrintToPDF(ctx, pdfArgs)
		if err != nil {
			log.Println("Error calling PrintToPDF: ", err)
			return nil, err
		}

		encodedData := base64.StdEncoding.EncodeToString([]byte(result.Data))
		return []byte(encodedData), nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}

// TODO:
// - Need to be able to set options like total timeout, network request timeouts, print args, etc..
// - handle context timeout errors (i.e. requests greater than the number set in WithTimeout)
func PrintPdfFromHtml(chromeUrl string, options *Options, html string) ([]byte, error) {
	return generatePdf(chromeUrl, len(html), options, func(c *cdp.Client, ctx context.Context, domContent page.LoadEventFiredClient) (interface{}, error) {
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
		return nil, nil
	})
}

func PrintPdfFromUrl(chromeUrl string, options *Options, url string) ([]byte, error) {
	return generatePdf(chromeUrl, 2048, options, func(c *cdp.Client, ctx context.Context, domContent page.LoadEventFiredClient) (interface{}, error) {
		navArgs := page.NewNavigateArgs(url)
		_, err := c.Page.Navigate(ctx, navArgs)
		if err != nil {
			log.Println("Error navigating to page")
			return nil, err
		}

		if _, err = domContent.Recv(); err != nil {
			log.Println("Error waiting for navigation")
			return nil, err
		}

		return nil, nil
	})
}
