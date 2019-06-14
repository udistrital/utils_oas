package auditoria

import (

	"encoding/json"
	"net/http"
  "log"
  "github.com/astaxie/beego/context"
  
)

func GetJsonWithHeader(urlp string, target interface{}, ctx *context.Context) error {
  req, err := http.NewRequest("GET",urlp, nil)
  if err != nil {
    log.Fatal("Error reading request. ", err)
  }

  req.Header.Set("Authorization", ctx.Request.Header["Authorization"][0])
  client := &http.Client{}

  resp, err := client.Do(req)
  if err != nil {
    log.Fatal("Error reading response. ", err)
  }

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
