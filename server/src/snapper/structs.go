package snapper

type Request struct {
	Html *string `json:"html"`
}

type Response struct {
	PdfData string `json:"pdfData"`
}
