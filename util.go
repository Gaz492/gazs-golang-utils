package gazs_golang_utils

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func Cleanup(tmpFile *os.File) {
	if err := tmpFile.Close(); err != nil {
		log.Fatal("Unable to close temp file: ", err)
	}
	if err := os.Remove(tmpFile.Name()); err != nil {
		log.Fatal(err)
	}
}

func CleanupFolder(tmpFolder string) {
	if err := os.RemoveAll(tmpFolder); err != nil {
		log.Fatal(err)
	}
}

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func HandleGetRequest(url string, headers map[string][]string) (*http.Response, error) {
	resp, err := makeRequest("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, err
	}
	return resp, nil
}

func HandlePostRequest(url string, body []byte, headers map[string][]string) (*http.Response, error) {
	resp, err := makeRequest("POST", url, body, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Error: %d", resp.StatusCode))
	}
	return resp, nil
}

func makeRequest(method, url string, b []byte, requestHeaders map[string][]string) (*http.Response, error) {
	headers := map[string][]string{
		"Accept": []string{"application/json"},
	}
	if method == "POST" {
		headers["Content-Type"] = []string{"application/json"}
	}
	/*if authKey != "" {
		if strings.HasPrefix(authKey, "$") {
			headers["x-api-key"] = []string{authKey}
		} else if strings.HasPrefix(authKey, "Bearer") {
			headers["Authorization"] = []string{authKey}
		} else {
			headers["Authorization"] = []string{"Basic " + authKey}
		}
	}*/
	for k, v := range requestHeaders {
		headers[k] = v
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header = headers

	return client.Do(req)
}
