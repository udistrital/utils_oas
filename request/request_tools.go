package request

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/xray"
)

var global string

func SendJson(urlp string, trequest string, target interface{}, datajson interface{}) error {
	b := new(bytes.Buffer)
	if datajson != nil {
		if err := json.NewEncoder(b).Encode(datajson); err != nil {
			beego.Error(err)
		}
	}

	client := &http.Client{}
	req, err := http.NewRequest(trequest, urlp, b)
	seg := xray.BeginSegmentSec(req)
	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.

	//try
	header := GetHeader()
	req.Header.Set("Authorization", header)
	req.Header.Set("Accept", AppJson)
	req.Header.Add("Content-Type", AppJson)
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		beego.Error("Error reading response. ", err)
		return err
	}
	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	defer func() {
		//Catch
		if r := recover(); r != nil {
			client := &http.Client{}
			resp, err := client.Do(req)
			xray.UpdateSegment(resp, err, seg)
			if err != nil {
				beego.Error("Error reading response. ", err)
			}

			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(target)
		}
	}()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()
	return json.NewDecoder(resp.Body).Decode(target)
}

func SendJson2(url string, trequest string, target interface{}, datajson interface{}) error {
	b := new(bytes.Buffer)
	if datajson != nil {
		if err := json.NewEncoder(b).Encode(datajson); err != nil {
			beego.Error(err)
		}
	}

	client := &http.Client{}
	req, _ := http.NewRequest(trequest, url, b)
	seg := xray.BeginSegmentSec(req)
	defer func() {
		//Catch
		if r := recover(); r != nil {
			client := &http.Client{}
			resp, err := client.Do(req)
			xray.UpdateSegment(resp, err, seg)
			if err != nil {
				beego.Error("Error reading response. ", err)
			}

			defer resp.Body.Close()
			mensaje, err := io.ReadAll(resp.Body)
			if err != nil {
				beego.Error("Error converting response. ", err)
			}
			bodyreq, err := io.ReadAll(req.Body)
			if err != nil {
				beego.Error("Error converting response. ", err)
			}
			respuesta := map[string]interface{}{"request": map[string]interface{}{"url": req.URL.String(), "header": req.Header, "body": bodyreq}, "body": mensaje, "statusCode": resp.StatusCode, "status": resp.Status}
			e, err := json.Marshal(respuesta)
			if err != nil {
				logs.Error(err)
			}
			json.Unmarshal(e, &target)
		}
	}()

	req.Header.Set("Authorization", "")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("accept", "*/*")
	r, err := client.Do(req)
	xray.UpdateSegment(r, err, seg)
	if err != nil {
		beego.Error("error", err)
		return err
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			beego.Error(err)
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
	//proxyUrl, err := url.Parse("http://10.20.4.15:3128")
	//http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}

	client := &http.Client{}
	req, err := http.NewRequest(trequest, urlp, b)
	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	defer func() {
		//Catch
		if r := recover(); r != nil {

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				beego.Error("Error reading response. ", err)
			}

			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(target)
		}
	}()

	//try
	header := GetHeader()
	req.Header.Set("Authorization", header)
	seg := xray.BeginSegmentSec(req)
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		beego.Error("Error reading response. ", err)
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonWSO2(urlp string, target interface{}) error {
	b := new(bytes.Buffer)
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlp, b)
	req.Header.Set("Accept", AppJson)
	seg := xray.BeginSegmentSec(req)
	r, err := client.Do(req)
	xray.UpdateSegment(r, err, seg)
	if err != nil {
		beego.Error("error", err)
		return err
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()

	return json.NewDecoder(r.Body).Decode(target)
}

func GetJsonWSO2Test(urlp string, target interface{}) (status int, err error) {
	b := new(bytes.Buffer)
	req, _ := http.NewRequest("GET", urlp, b)
	req.Header.Set("Accept", AppJson)
	seg := xray.BeginSegmentSec(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		beego.Error("error", err)
		return resp.StatusCode, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			beego.Error(nil, err)
		}
	}()
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}
func GetJson(urlp string, target interface{}) error {
	req, err := http.NewRequest("GET", urlp, nil)
	if err != nil {
		beego.Error("Error reading request. ", err)
	}
	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	seg := xray.BeginSegmentSec(req)
	header := GetHeader()
	req.Header.Set("Authorization", header)
	client := &http.Client{}
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		beego.Error("Error reading response. ", err)
		return err
	}
	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	defer func() {
		//Catch
		if r := recover(); r != nil {
			client := &http.Client{}
			resp, err := client.Do(req)
			xray.UpdateSegment(resp, err, seg)
			if err != nil {
				beego.Error("Error reading response. ", err)
			}
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(target)
		}
	}()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()
	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonTest(url string, target interface{}) (response *http.Response, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	seg := xray.BeginSegmentSec(req)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()
	return resp, json.NewDecoder(response.Body).Decode(target)
}

func GetJsonTest2(url string, target interface{}) (status int, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	seg := xray.BeginSegmentSec(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return resp.StatusCode, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(target)
}

func GetXml(url string, target interface{}) error {
	req, _ := http.NewRequest("GET", url, nil)
	seg := xray.BeginSegmentSec(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		beego.Error("Error reading response. ", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()
	return xml.NewDecoder(resp.Body).Decode(target)
}

func GetXML2String(url string, target interface{}) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		beego.Error("Error reading request. ", err)
	}

	client := &http.Client{}
	seg := xray.BeginSegmentSec(req)
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		beego.Error("Error reading response. ", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		beego.Error("Error reading response. ", err)
	}

	print(body)
	print(string(body))
	s := strings.TrimSpace(string(body))
	return s
}

func SetHeader(h string) {
	global = h
}

func GetHeader() (h string) {
	return global
}
