package request

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

var global *context.Context

func SendJson(urlp string, trequest string, target interface{}, datajson interface{}) error {
	b := new(bytes.Buffer)
	if datajson != nil {
		json.NewEncoder(b).Encode(datajson)
	}
	//proxyUrl, err := url.Parse("http://10.20.4.15:3128")
	//http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}

	client := &http.Client{}
	req, err := http.NewRequest(trequest, urlp, b)

	//Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	defer func () {
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
	header := GetHeader().Request.Header
	fmt.Println("header send:",header)
	req.Header.Set("Authorization", header["Authorization"][0])

	resp, err := client.Do(req)
	if err != nil {
		beego.Error("Error reading response. ", err)
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func GetJsonWSO2(urlp string, target interface{}) error {
	b := new(bytes.Buffer)
	//proxyUrl, err := url.Parse("http://10.20.4.15:3128")
	//http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlp, b)
	req.Header.Set("Accept", "application/json")
	r, err := client.Do(req)
	//r, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		fmt.Println("error", err)
		//beego.Error("error", err)
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func SetHeader(ctx *context.Context){
	global = ctx

}

func GetHeader()(ctx *context.Context){
	return global
}

func GetJson(urlp string, target interface{}) error {
	
	
	req, err := http.NewRequest("GET",urlp, nil)
	  if err != nil {
		beego.Error("Error reading request. ", err)
	  }

	  //Se intenta acceder a cabecera, si no existe, se realiza peticion normal.
	
		defer func () {
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
	  header := GetHeader().Request.Header
	  fmt.Println("header get:",header)
	  req.Header.Set("Authorization", header["Authorization"][0])
	  client := &http.Client{}
	
	  resp, err := client.Do(req)
	  if err != nil {
		beego.Error("Error reading response. ", err)
	  }
	
		defer resp.Body.Close()
		return json.NewDecoder(resp.Body).Decode(target)
	}

func GetJsonTest(url string, target interface{}) (response *http.Response, err error) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	response, err = myClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return response, json.NewDecoder(response.Body).Decode(target)
}

func diff(a, b time.Time) (year, month, day int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)

	// Normalize negative values

	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}
