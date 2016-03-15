package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var random *rand.Rand

// githubize transforms the user/repo github address to a valid url
func githubize(url string) (string, error) {
	splitOnSlash := strings.Split(url, "/")
	if len(splitOnSlash) < 3 {
		return "", errors.New("invalid github address")
	}
	splitOnSlash = splitOnSlash[1:] // strip out "github.com""
	return fmt.Sprintf("%s%s",
		"https://",
		path.Join("raw.githubusercontent.com",
			splitOnSlash[0],
			splitOnSlash[1],
			"master",
			path.Join(splitOnSlash[2:]...))), nil
}

func FetchURL(url string) ([]byte, error) {
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
	} else if strings.HasPrefix(url, "github.com") {
		var err error
		if githubizedURL, err := githubize(url); err == nil {
			return FetchURL(githubizedURL)
		}
		return []byte{}, err
	}
	return ioutil.ReadFile(url)
}

func JoinURL(baseURL string, secondURL string) string {
	// only combine path part if baseURL is url not a local file and it make sense
	base, err := url.Parse(baseURL)
	if err == nil {
		base.Path = path.Join(base.Path, secondURL)
		return base.String()
	}
	return path.Join(baseURL, secondURL)
}

func GetMapChild(keys []string, m map[string]interface{}) (interface{}, error) {
	var innerMap map[string]interface{}
	var v interface{}
	var has, ok bool
	for i := range keys {
		if innerMap == nil {
			innerMap = m
		} else {
			innerMap, ok = v.(map[string]interface{})
			if !ok {
				return nil, errors.New("key not found " + keys[i])
			}
		}
		v, has = innerMap[keys[i]]
		if !has {
			return nil, errors.New("key not found " + keys[i])
		}
	}
	return v, nil
}

func MergeStringMaps(a map[string]string, b map[string]string) map[string]string {
	if a == nil {
		a = make(map[string]string)
	}
	out := make(map[string]string)
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func GenerateRandomString(chars string, length int) string {
	if random == nil {
		random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	randStr := make([]byte, length)
	charsLen := len(chars)
	for i := 0; i < length; i++ {
		randStr[i] = chars[random.Intn(charsLen)]
	}
	return string(randStr)
}

func GenerateRandomGroup() string {
	return GenerateRandomString("abcdefghijklmnopqrstuvwxyz", 1) + GenerateRandomString("abcdefghijklmnopqrstuvwxyz1234567890", 7)
}

func StringToBool(s string) (bool, error) {
	s = strings.ToLower(s)
	if s == "true" || s == "yes" || s == "y" || s == "ok" || s == "t" || s == "1" {
		return true, nil
	} else if s == "false" || s == "n" || s == "no" || s == "f" || s == "0" {
		return false, nil
	}
	return false, fmt.Errorf("can't parse \"%s\" as boolean", s)
}

// Relative returns true when p is a relative path
// returns false otherwise: absolute path, http(s)://url, or an address
// with github.com prefix
func Relative(p string) bool {
	return len(p) == 0 || p[0] == '.' ||
		(!strings.HasPrefix(p, "github.com") &&
			!strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://"))
}
