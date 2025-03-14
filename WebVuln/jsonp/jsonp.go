package jsonp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/BreakOnCrash/opendast/WebVuln/jsonp/js"
)

var (
	jsonpQueryRegex = regexp.MustCompile(`(?m)(?i)(callback)|(jsonp)|(^cb$)|(function)`)
	jsonpValueRegex = regexp.MustCompile(`(?m)(?i)(uid)|(userid)|(user_id)|(nin)|(name)|(username)|(nick)|(nickname)|(memberid)|(loginid)|(mobilephone)|(passportid)|(profile)|(profile)|(c)|(loginid)|(email)|(realname)|(birthday)|(sex)|(ip)`)

	// https://github.com/chromium/chromium/blob/fc262dcd403c74cf3e22896f32d9723ba463f0b6/third_party/blink/common/mime_util/mime_util.cc#L42
	javascriptTypes = []string{
		"application/ecmascript",
		"application/javascript",
		"application/x-ecmascript",
		"application/x-javascript",
		"text/ecmascript",
		"text/javascript",
		"text/javascript1.0",
		"text/javascript1.1",
		"text/javascript1.2",
		"text/javascript1.3",
		"text/javascript1.4",
		"text/javascript1.5",
		"text/jscript",
		"text/livescript",
		"text/x-ecmascript",
		"text/x-javascript",
	}
)

func AuditJSONPHijacking(urlStr string) error {
	URL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// find callback function name
	var callName string
	for k, v := range URL.Query() {
		if len(v) == 0 {
			continue
		}
		if jsonpQueryRegex.MatchString(k) {
			callName = v[0]
		}
	}

	if callName == "" {
		return errors.New("not found callback query")
	}

	referer := URL.Scheme + "://" + URL.Host
	content, err := fetchWithReferer(urlStr, referer)
	if err != nil {
		return err
	}

	params, err := js.ParseJSCode(content, callName)
	if err != nil {
		return err
	}

	// 如果包含敏感字段,将 referer 置空再请求一次
	if jsonpValueRegex.MatchString(params) {
		noRefererContent, err := fetchWithReferer(urlStr, "")
		if err != nil {
			return err
		}
		params, err := js.ParseJSCode(noRefererContent, callName)
		if err != nil {
			return err
		}
		if jsonpValueRegex.MatchString(params) {
			fmt.Println("Exist JSONP Hijacking!!!")
		}
	}

	return nil
}

func fetchWithReferer(URL, referer string) (string, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	if referer == "" {
		req.Header.Del("Referer")
	} else {
		req.Header.Set("Referer", referer)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("status code error")
	}

	// check response content-type
	if !containsContentType(resp.Header, javascriptTypes...) {
		return "", errors.New("not javascript content")
	}

	return string(body), nil
}

func containsContentType(h http.Header, contentTypes ...string) bool {
	v := h.Get("Content-Type")
	if v == "" {
		return false
	}

	if parts := strings.SplitN(v, ";", 2); len(parts) >= 1 {
		mainType := strings.TrimSpace(parts[0])
		for _, ct := range contentTypes {
			if strings.EqualFold(mainType, ct) {
				return true
			}
		}
	}
	return false
}
