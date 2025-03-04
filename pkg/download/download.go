package download

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var ErrEmptyURL = errors.New("empty url")

const UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36"

type HttpClient struct {
	*http.Client
}

var httpClient *HttpClient

func init() {
	httpClient = &HttpClient{http.DefaultClient}
	httpClient.Timeout = time.Second * 60
	httpClient.Transport = &http.Transport{
		TLSHandshakeTimeout:   time.Second * 5,
		IdleConnTimeout:       time.Second * 10,
		ResponseHeaderTimeout: time.Second * 10,
		ExpectContinueTimeout: time.Second * 20,
		Proxy:                 http.ProxyFromEnvironment,
	}
}

func (c *HttpClient) Get(urls ...string) (body []byte, err error) {
	var req *http.Request
	var resp *http.Response

	for _, url := range urls {
		req, err = http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println(err)
			continue
		}
		req.Header.Set("User-Agent", UserAgent)
		resp, err = c.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			return
		}
	}

	return nil, err
}

func Download(filePath string, url string) (err error) {
	if len(url) == 0 {
		return ErrEmptyURL
	}

	data, err := httpClient.Get(url)
	if err != nil {
		return err
	}

	return SaveFile(filePath, data)
}

func SaveFile(path string, data []byte) (err error) {
	// Remove file if exist
	if _, err = os.Stat(path); err == nil {
		if err = os.Remove(path); err != nil {
			return err
		}
	}

	d := filepath.Dir(path)
	if len(d) != 0 {
		if !FileExist(d) {
			if err := os.MkdirAll(d, os.ModePerm); err != nil {
				return err
			}
		}
	}

	// save file
	return os.WriteFile(path, data, 0644)
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
