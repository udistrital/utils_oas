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
		logs.Info("usando credenciales locales para la conexión a la base de datos")
		conn := formatPostgresConnectionString(beego.AppConfig.String("PGuser"), beego.AppConfig.String("PGpass"))
		return conn, nil
	}

	parameterBasePath := "/" + baseParameterStore + "/" + beego.AppConfig.String("appname") + "/db/"

	username, err := ssm.GetParameterFromParameterStore(parameterBasePath + "username")
	if err != nil {
		logs.Critical("error consultando username: %v", err)
		return "", errors.New("error consultando credenciales: " + err.Error())
	}

	password, err := ssm.GetParameterFromParameterStore(parameterBasePath + "password")
	if err != nil {
		logs.Critical("error consultando credenciales: %v", err)
		return "", errors.New("error consultando credenciales: " + err.Error())
	}

	conn := formatPostgresConnectionString(username, password)

	return conn, nil
}

func formatPostgresConnectionString(username, password string) string {
	return "postgres://" + username + ":" + url.QueryEscape(password) + "@" + beego.AppConfig.String("PGhost") + ":" + beego.AppConfig.String("PGport") + "/" + beego.AppConfig.String("PGdb") + "?sslmode=disable&search_path=" + beego.AppConfig.String("PGschema")
}
