package requestresponse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/xray"
)

var inspectRequest bool

// ? Request manager al estilo REST para transacciones json unicamente
func init() { // ? Init para cargar variable inspectRequest desde app.conf (Solo para pruebas en local y seguir las peticiones)
	inspectRequest, _ = beego.AppConfig.Bool("inspectRequest")
}

// PostRequest con manejo de status code y verificación de data existente mediante parser
//   - url: api + endpoint a consultar
//   - body: data como interface{}
//   - parser: función que formatea y valida la data; ver ~/utils_oas/requestResponse/responses_format
//
// Retorna:
//   - data como interface{}
//   - error si existe validando status y data no vacia o invalida
func Post(url string, body interface{}, parser func(interface{}) (interface{}, error)) (interface{}, error) {
	// ? se prepara el body de la petición
	bodyBytes := new(bytes.Buffer)
	if body == nil {
		return nil, fmt.Errorf("body is empty")
	}
	json.NewEncoder(bodyBytes).Encode(body)

	jsonResponse, err := _doReq("POST", url, bodyBytes)
	if err != nil {
		return jsonResponse, err
	}
	// ? Se pasa la respuesta por un parser que convaida la estructura de la respuesta
	data, err := parser(jsonResponse)
	if err != nil {
		return data, err
	}
	return data, nil
}

// GetRequest con manejo de status code y verificación de data existente mediante parser
//   - url: api + endpoint a consultar
//   - parser: función que formatea y valida la data; ver ~/utils_oas/requestResponse/responses_format
//
// Retorna:
//   - data como interface{}
//   - error si existe validando status y data no vacia o invalida
func Get(url string, parser func(interface{}) (interface{}, error)) (interface{}, error) {
	jsonResponse, err := _doReq("GET", url, nil)
	if err != nil {
		return jsonResponse, err
	}
	// ? Se pasa la respuesta por un parser que convaida la estructura de la respuesta
	data, err := parser(jsonResponse)
	if err != nil {
		return data, err
	}
	return data, nil
}

// PutRequest con manejo de status code y verificación de data existente mediante parser
//   - url: api + endpoint a consultar
//   - body: data como interface{}
//   - parser: función que formatea y valida la data; ver ~/utils_oas/requestResponse/responses_format
//
// Retorna:
//   - data como interface{}
//   - error si existe validando status y data no vacia o invalida
func Put(url string, body interface{}, parser func(interface{}) (interface{}, error)) (interface{}, error) {
	// ? se prepara el body de la petición
	bodyBytes := new(bytes.Buffer)
	if body == nil {
		return nil, fmt.Errorf("body is empty")
	}
	json.NewEncoder(bodyBytes).Encode(body)

	jsonResponse, err := _doReq("PUT", url, bodyBytes)
	if err != nil {
		return jsonResponse, err
	}
	// ? Se pasa la respuesta por un parser que convaida la estructura de la respuesta
	data, err := parser(jsonResponse)
	if err != nil {
		return data, err
	}
	return data, nil
}

// DeleteRequest con manejo de status code y verificación de data existente mediante parser
//   - url: api + endpoint a consultar
//   - parser: función que formatea y valida la data; ver ~/utils_oas/requestResponse/responses_format
//
// Retorna:
//   - data como interface{}
//   - error si existe validando status y data no vacia o invalida
func Delete(url string, parser func(interface{}) (interface{}, error)) (interface{}, error) {
	jsonResponse, err := _doReq("DELETE", url, nil)
	if err != nil {
		return jsonResponse, err
	}
	// ? Se pasa la respuesta por un parser que convaida la estructura de la respuesta
	data, err := parser(jsonResponse)
	if err != nil {
		return data, err
	}
	return data, nil
}

// Función general que realmente realiza las peticiones, añade headers, verifica si la respuesta es json y si status ok
func _doReq(method string, url string, body io.Reader) (interface{}, error) {
	if inspectRequest { // ? Print para debugging
		fmt.Println(method, url)
		fmt.Println(body)
	}
	// ? Preparar petición
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if (method == "POST") || (method == "PUT") {
		req.Header.Set("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
	}
	// ? Realizar la petición
	client := &http.Client{}
	seg := xray.BeginSegmentSec(req)
	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // ? Terminar la petición en caso de fallo o no
	// ? Verifica si body de la respuesta es un json
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf("not a JSON response")
	}
	// ? Decodifica el body de la respuesta a interface{}
	var jsonResponse interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return nil, err
	}
	// ? Se checkea el Status para saber si la petición es exitosa o no
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return jsonResponse, fmt.Errorf("not successful response")
	}

	if inspectRequest { // ? Print para debugging
		bytes, _ := json.MarshalIndent(jsonResponse, "", "  ")
		fmt.Println(string(bytes))
	}
	return jsonResponse, nil
}
