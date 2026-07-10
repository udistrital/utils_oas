package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/udistrital/utils_oas/v2/ssm"
)

func BuildPostgresConnectionString() (string, error) {
	baseParameterStore, _ := beego.AppConfig.String("parameterStore")
	if baseParameterStore == "" {
		logs.Info("usando credenciales locales para la conexión a la base de datos")
		user, _ := beego.AppConfig.String("PGuser")
		pass, _ := beego.AppConfig.String("PGpass")
		conn := formatPostgresConnectionString(user, pass)
		return conn, nil
	}

	appname, _ := beego.AppConfig.String("appname")
	parameterBasePath := fmt.Sprintf("/%s/%s/db/", baseParameterStore, appname)

	ctx := context.Background()

	username, err := ssm.GetValueFromParameterStore(ctx, parameterBasePath+"username")
	if err != nil {
		logs.Critical("error consultando username: %v", err)
		return "", errors.New("error consultando credenciales: " + err.Error())
	}

	password, err := ssm.GetValueFromParameterStore(ctx, parameterBasePath+"password")
	if err != nil {
		logs.Critical("error consultando credenciales: %v", err)
		return "", errors.New("error consultando credenciales: " + err.Error())
	}

	conn := formatPostgresConnectionString(username, password)

	return conn, nil
}

func formatPostgresConnectionString(username, password string) string {
	host, _ := beego.AppConfig.String("PGhost")
	port, _ := beego.AppConfig.String("PGport")
	db, _ := beego.AppConfig.String("PGdb")
	schema, _ := beego.AppConfig.String("PGschema")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?search_path=%s&sslmode=disable",
		username,
		url.QueryEscape(password),
		host,
		port,
		db,
		schema)
}
