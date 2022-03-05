# how to use 

```go
package main

import (
	"log"

	gtbg "github.com/fy403/gotenberg-client-go"
	"github.com/google/uuid"
)

func main() {
	dir := "\\Users\\Administrator\\print"

	paths := []string{
		"https://www.rfc-editor.org/rfc/pdfrfc/rfc3510.txt.pdf",
		dir + "\\" + "Description.doc",
		dir + "\\" + "print_test.pdf",
		dir + "\\" + "file.xlsx",
	}

	url := "http://localhost:3000/forms/libreoffice/convert"
	formKey := "files"
	extraParams := map[string]string{
		"merge":     "true",
		"pdfFormat": "PDF/A-1a",
	}
	saveFileName := "merge_" + uuid.NewString()

	gotenberg := gtbg.NewGotenberg(url)
	request, err := gotenberg.NewRequest(extraParams, formKey, paths...)
	if err != nil {
		log.Fatalln(err)
	}
	savePdfPath, err := gotenberg.Send(request, dir, saveFileName)
	if err != nil {
		log.Fatalln(err)
	}
	pages, err := gotenberg.Pdfpages(savePdfPath)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Pages: %d\n", pages)
}

```