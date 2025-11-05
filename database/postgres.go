package database

import (
	"errors"
	"net/url"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/ssm"
)

func BuildPostgresConnectionString() (string, error) {

	baseParameterStore := beego.AppConfig.String("parameterStore")
	if baseParameterStore == "" {
		err := errors.New("parameterStore no configurado")
		logs.Critical(err)
		return "", err
	}

	parameterStore := "/" + baseParameterStore + "/" + beego.AppConfig.String("appname") + "/db/"

	username, err := ssm.GetParameterFromParameterStore(parameterStore + "username")
	if err != nil {
		logs.Critical("error consultando credenciales: %v", err)
		return "", errors.New("error consultando credenciales: " + err.Error())
	}

	password, err := ssm.GetParameterFromParameterStore(parameterStore + "password")
	if err != nil {
		logs.Critical("error consultando credenciales: %v", err)
		return "", errors.New("error consultando credenciales: " + err.Error())
	}

	conn := "postgres://" + username + ":" + url.QueryEscape(password) + "@" + beego.AppConfig.String("PGhost") + ":" + beego.AppConfig.String("PGport") + "/" + beego.AppConfig.String("PGdb") + "?sslmode=disable&search_path=" + beego.AppConfig.String("PGschema")

	return conn, nil
}
