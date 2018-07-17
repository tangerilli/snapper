package snapper

type Options struct {
	Timeout *int `json:"timeout"`
}

func SetDefaultOptions(options *Options) {
	if options.Timeout == nil {
		options.Timeout = new(int)
		*options.Timeout = 15
	}
}

type Request struct {
	Html    *string  `json:"html"`
	Url     *string  `json:"url"`
	Options *Options `json:"options"`
}

type Response struct {
	PdfData string `json:"pdfData"`
}
