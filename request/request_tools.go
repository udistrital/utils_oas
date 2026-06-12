package request

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/xray"
)

var global string
var defClient = &http.Client{}

func SendJson(url string, method string, target, body any) error {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return fmt.Errorf("could not encode request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	req.Header.Set(authorizationKey, GetHeader())
	req.Header.Set(acceptHeader, contentTypeJSON)
	req.Header.Set(contentTypeKey, contentTypeJSON)

	resp, err := execRequest(defClient, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func SendJson2(url string, method string, target, body any) error {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return fmt.Errorf("could not encode request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set(authorizationKey, "")
	req.Header.Set(contentTypeKey, "application/json; charset=UTF-8")
	req.Header.Set(acceptHeader, "*/*")

	resp, err := execRequest(defClient, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func SendJsonEscapeUnicode(url string, method string, target, body any) error {
	b := new(bytes.Buffer)
	if body != nil {
		e := json.NewEncoder(b)
		e.SetEscapeHTML(false)
		if err := e.Encode(body); err != nil {
			return fmt.Errorf("could not encode request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set(authorizationKey, GetHeader())

	resp, err := execRequest(defClient, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonWSO2(url string, target any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set(acceptHeader, contentTypeJSON)
	resp, err := execRequest(defClient, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonWSO2Test(url string, target any) (status int, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set(acceptHeader, contentTypeJSON)
	resp, err := execRequest(defClient, req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}

func GetJson(url string, target any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set(authorizationKey, GetHeader())

	resp, err := execRequest(defClient, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonTest(url string, target any) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := execRequest(client, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return resp, json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonTest2(url string, target any) (status int, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := execRequest(defClient, req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}

func GetXml(url string, target any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	resp, err := execRequest(defClient, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	return xml.NewDecoder(resp.Body).Decode(target)
}

func GetXML2String(url string, target any) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error(fmt.Errorf("could not create request: %w", err))
		return ""
	}

	resp, err := execRequest(defClient, req)
	if err != nil {
		logs.Error(fmt.Errorf("request failed: %w", err))
		return ""
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return ""
	}

	return strings.TrimSpace(string(body))
}

func SetHeader(h string) {
	global = h
}

func GetHeader() (h string) {
	return global
}

// execRequest executes req using the provided HTTP client, wrapping the call
// with an X-Ray subsegment via the local xray package.
// The caller is responsible for closing resp.Body on success.
func execRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	ctx, subseg := xray.BeginSegmentSec(req)
	resp, err := client.Do(req.WithContext(ctx))
	xray.CloseSubsegment(subseg, resp, err)
	return resp, err
}
