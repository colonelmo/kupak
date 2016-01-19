package kupak

import (
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

func fetchUrl(url string) ([]byte, error) {
	if strings.HasPrefix(strings.ToLower(url), "http://") ||
		strings.HasPrefix(strings.ToLower(url), "https://") {
		c := &http.Client{}
		resp, err := c.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return ioutil.ReadFile(url)
	}
}

func joinUrl(baseUrl string, url string) string {
	return path.Join(path.Dir(baseUrl), url)
}
