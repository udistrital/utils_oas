# Estrategia de implementación librería Auditoría

Con el fin de mantener un registro completo de las transacciones y peticiones realizadas por un sistema, y saber qué se ha hecho, cuándo se ha hecho y quién lo ha hecho, se debe realizar la implementación de la librería de auditoría en APIs mid y crud.

Requerimientos:

* Implementar la última versión de la librería utils_oas y utilizar los métodos allí dispuestos para realizar peticiones REST

## Implementación Auditoria en APIs CRUD

Para la implementación de la librería en APIs CRUD se deben seguir los siguientes pasos:

1. En la sección de código del archivo main.go correspondiente al import, realizar la importación de la librería de auditoría

```go
  import (
    _ "github.com/udistrital/titan_api_crud/routers"
    "github.com/udistrital/utils_oas/apiStatusLib"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
    _ "github.com/lib/pq"
    "github.com/astaxie/beego/plugins/cors"
    "github.com/udistrital/utils_oas/customerror"
    "github.com/udistrital/auditoria"
  )
```

2. En el mismo archivo main.go, realizar el llamado a la librería por medio del código auditoria.InitMiddleware():

```go
  func main() {
    orm.RegisterDataBase("default", "postgres", "postgres://"+beego.AppConfig.String("PGuser")+":"+beego.AppConfig.String("PGpass")+"@"+beego.AppConfig.String("PGurls")+"/"+beego.AppConfig.String("PGdb")+"?sslmode=disable&search_path="+beego.AppConfig.String("PGschemas")+"")
    apistatus.Init()
    //Prueba de auditoria
    auditoria.InitMiddleware()
    beego.ErrorController(&customerror.CustomErrorController{})
    beego.Run()
  }
```

3. Para entornos locales, basta con ejecutar nuevamente el API para que la librería de auditoría genere los logs; esto puede ser revisado en la consola. Para entornos de desarrollo (dev), preproducción(test) y producción(prod), se debe realizar el respectivo push a Github, que permite la construcción del API en estos entornos y por ende la ejecución de la última versión de la librería.

## Implementación Auditoria en APIs MID

Para la implementación de la librería en APIs CRUD se deben seguir los siguientes pasos:

1. En la sección de código del archivo main.go correspondiente al import, realizar la importación de la librería de auditoría

```go
  import (
    _ "github.com/udistrital/titan_api_mid/routers"
    "github.com/udistrital/utils_oas/apiStatusLib"
    "github.com/astaxie/beego/plugins/cors"
    "github.com/astaxie/beego"
    "github.com/udistrital/auditoria"

  )
```

2. En el mismo archivo main.go, realizar el llamado a la librería por medio del código auditoria.InitMiddleware():

```go
  func main() {
    orm.RegisterDataBase("default", "postgres", "postgres://"+beego.AppConfig.String("PGuser")+":"+beego.AppConfig.String("PGpass")+"@"+beego.AppConfig.String("PGurls")+"/"+beego.AppConfig.String("PGdb")+"?sslmode=disable&search_path="+beego.AppConfig.String("PGschemas")+"")
    apistatus.Init()
    //Prueba de auditoria
    auditoria.InitMiddleware()
    beego.ErrorController(&customerror.CustomErrorController{})
    beego.Run()
  }
```

3. Crear un archivo llamado interceptor.go en la carpeta raíz del proyecto:

```
myproject
├── conf
│   └── app.conf
├── controllers
│   └── default.go
├── main.go
├── models
├── routers
│   └── router.go
├── static
│   ├── css
│   ├── img
│   └── js
├── tests
│   └── default_test.go
└── views
|   └── index.tpl│
└── interceptor.go
```

4. Copiar en él el mismo contenido disponible en
[Interceptor](https://github.com/udistrital/auditoria/blob/dev/interceptor.go). Este archivo debe lucir así:

```go
  package main
  import (
    "github.com/astaxie/beego"
   "github.com/astaxie/beego/context"
    "github.com/udistrital/utils_oas/request"
  )

  func InterceptMidRequest(ctx *context.Context) {
    end_point := ctx.Request.URL.String()
     if end_point != "/" {
       defer func () {
        //Catch
        if r := recover(); r != nil {
  }
       }()
      // try
      request.SetHeader(ctx.Request.Header["Authorization"][0])
    }

 }

  func InitInterceptor() {
    beego.InsertFilter("*", beego.BeforeExec, InterceptMidRequest, false)
  }
```

5. Para entornos locales, basta con ejecutar nuevamente el API para que la librería de auditoría genere los logs; esto puede ser revisado en la consola. Para entornos de desarrollo (dev), preproducción(test) y producción(prod), se debe realizar el respectivo push a Github, que permite la construcción del API en estos entornos y por ende la ejecución de la última versión de la librería.
