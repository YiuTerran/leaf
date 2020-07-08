package netutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

func Download(uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "ftp") {
		return ftpDownload(uri)
	}
	if strings.HasPrefix(uri, "http") {
		return httpDownload(uri)
	}
	return nil, errors.New("not support protocol")
}

func DownloadAsFile(uri string, filepath string) error {
	bs, err := Download(uri)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, bs, os.ModePerm)
}

func httpDownload(uri string) ([]byte, error) {
	// Get the data
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func ftpDownload(uri string) ([]byte, error) {
	link, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("fail to parse url:%w", err)
	}
	c, err := ftp.Dial(link.Host, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return nil, err
	}

	if err = c.Login("anonymous", "anonymous"); err != nil {
		return nil, err
	}
	defer c.Quit()
	if r, err := c.Retr(link.Path); err != nil {
		return nil, err
	} else {
		defer r.Close()
		return ioutil.ReadAll(r)
	}
}
