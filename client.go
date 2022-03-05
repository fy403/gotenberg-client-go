package client

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type Gotenberg struct {
	backendUrl         string
	contentTypeRegeExp string
	pathSeparator      string
	client             *http.Client
}

func NewGotenberg(backendUrl string) *Gotenberg {
	var sep string
	if runtime.GOOS == "windows" {
		sep = "\\"
	} else {
		sep = "/"
	}
	return &Gotenberg{
		backendUrl:         backendUrl,
		contentTypeRegeExp: `application\/[a-z]+`,
		pathSeparator:      sep,
		client:             &http.Client{},
	}
}

// Creates a new file upload http request with optional extra params
func (gtbg *Gotenberg) NewRequest(params map[string]string, paramName string, paths ...string) (request *http.Request, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, path := range paths {
		if gtbg.isNetworkPath(path) {
			// Download
			tmpPath, err := gtbg.downloadFile(path)
			if err != nil {
				return nil, err
			}
			path = tmpPath
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		part, err := writer.CreateFormFile(paramName, filepath.Base(path))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
		file.Close()
	}
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	// Must Close to add some boundary in body
	// Don`t use defer
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	request, err = http.NewRequest("POST", gtbg.backendUrl, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return
}

func (gtbg *Gotenberg) Send(request *http.Request, saveDirName, saveFileName string) (string, error) {
	resp, err := gtbg.client.Do(request)
	if err != nil {
		return "", err
	}
	reg, err := regexp.Compile(gtbg.contentTypeRegeExp)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Unknown StatusCode: ", resp.StatusCode)
	}
	var savePdfPath string
	if v := reg.FindAllString(resp.Header.Get("Content-Type"), -1); len(v) > 0 {
		_ = os.MkdirAll(saveDirName, 0755)
		savePdfPath = fmt.Sprintf("%s%s%s.%s", saveDirName, gtbg.pathSeparator, saveFileName, strings.Split(v[0], "/")[1])
		out, err := os.Create(savePdfPath)
		if err != nil {
			return "", err
		}
		_, err = out.ReadFrom(resp.Body)
		if err != nil {
			return "", err
		}
		out.Close()
	}
	return savePdfPath, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (gtbg *Gotenberg) downloadFile(url string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unknown StatusCode: %d", resp.StatusCode)
	}
	reg, err := regexp.Compile(gtbg.contentTypeRegeExp)
	if err != nil {
		return "", err
	}
	var suffix string
	if v := reg.FindAllString(resp.Header.Get("Content-Type"), -1); len(v) > 0 &&
		!strings.Contains(resp.Header.Get("Content-Type"), "octet-stream") {
		suffix = strings.Split(v[0], "/")[1]
	} else {
		split := strings.Split(url, ".")
		if v := len(split); v > 0 {
			suffix = split[v-1]
		}
	}

	if suffix == "" {
		return "", fmt.Errorf("Unknown Centent-Type: %s", resp.Header.Get("Content-Type"))
	}
	saveFilePath := fmt.Sprintf("%s%s%s.%s", os.TempDir(), gtbg.pathSeparator, uuid.NewString(), suffix)
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(saveFilePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return saveFilePath, err
}

func (gtbg *Gotenberg) isNetworkPath(path string) bool {
	reg, _ := regexp.Compile(`^((ht|f)tps?):\/\/`)
	return len(reg.FindAllString(path, -1)) > 0
}

func (gtbg *Gotenberg) Pdfpages(pdfPath string) (int, error) {
	f, r, err := pdf.Open(pdfPath)
	// remember close file
	defer f.Close()
	if err != nil {
		return math.MaxInt32, err
	}
	totalPage := r.NumPage()
	return totalPage, nil
}
