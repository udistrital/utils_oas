package request

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/xray"
)

var global string
var defClient = &http.Client{}

func SendJson(urlp string, trequest string, target interface{}, datajson interface{}) error {
	b := new(bytes.Buffer)
	if datajson != nil {
		if err := json.NewEncoder(b).Encode(datajson); err != nil {
			logs.Error(err)
			return err
		}
	}

	req, err := http.NewRequest(trequest, urlp, b)
	if err != nil {
		logs.Error(err)
		return err
	}

	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	req.Header.Set(authorizationKey, GetHeader())
	req.Header.Set(acceptHeader, contentTypeJSON)
	req.Header.Set(contentTypeKey, contentTypeJSON)

	resp, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("Error reading response. ", err)
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()
	return json.NewDecoder(resp.Body).Decode(target)
}

func SendJson2(url string, trequest string, target interface{}, datajson interface{}) error {
	b := new(bytes.Buffer)
	if datajson != nil {
		if err := json.NewEncoder(b).Encode(datajson); err != nil {
			logs.Error(err)
			return err
		}
	}

	req, err := http.NewRequest(trequest, url, b)
	if err != nil {
		logs.Error(err)
		return err
	}

	req.Header.Set(authorizationKey, "")
	req.Header.Set(contentTypeKey, "application/json; charset=UTF-8")
	req.Header.Set(acceptHeader, "*/*")

	r, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("error", err)
		return err
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()

	return json.NewDecoder(r.Body).Decode(target)
}

func SendJsonEscapeUnicode(urlp string, trequest string, target interface{}, datajson interface{}) error {
	b := new(bytes.Buffer)
	if datajson != nil {
		e := json.NewEncoder(b)
		e.SetEscapeHTML(false)
		e.Encode(datajson)
	}

	req, err := http.NewRequest(trequest, urlp, b)
	if err != nil {
		logs.Error(err)
		return err
	}

	req.Header.Set(authorizationKey, GetHeader())

	resp, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("Error reading response. ", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()
	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonWSO2(urlp string, target interface{}) error {
	b := new(bytes.Buffer)
	req, err := http.NewRequest("GET", urlp, b)
	if err != nil {
		logs.Error(err)
		return err
	}
	req.Header.Set(acceptHeader, contentTypeJSON)
	r, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("error", err)
		return err
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()

	return json.NewDecoder(r.Body).Decode(target)
}

func GetJsonWSO2Test(urlp string, target interface{}) (status int, err error) {
	b := new(bytes.Buffer)
	req, err := http.NewRequest("GET", urlp, b)
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	req.Header.Set(acceptHeader, contentTypeJSON)
	resp, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("error", err)
		return 0, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}

func GetJson(urlp string, target interface{}) error {
	req, err := http.NewRequest("GET", urlp, nil)
	if err != nil {
		logs.Error("Error reading request. ", err)
		return err
	}

	req.Header.Set(authorizationKey, GetHeader())

	resp, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("Error reading response. ", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()
	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonTest(url string, target interface{}) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error(err)
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := execRequest(client, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()
	return resp, json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonTest2(url string, target interface{}) (status int, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	resp, err := execRequest(defClient, req)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}

func GetXml(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error(err)
		return err
	}
	resp, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("Error reading response. ", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()
	return xml.NewDecoder(resp.Body).Decode(target)
}

func GetXML2String(url string, target interface{}) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error("Error reading request. ", err)
		return ""
	}

	resp, err := execRequest(defClient, req)
	if err != nil {
		logs.Error("Error reading response. ", err)
		return ""
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logs.Error(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.Error("Error reading response. ", err)
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
	seg := xray.BeginSegmentSec(req)
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	return resp, err
}
