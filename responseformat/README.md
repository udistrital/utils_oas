# Response Format 

##### FINALIDAD DE LA LIBRERIA:


*Esta librería tiene como Objetivo interceptar todas los errores a escala de ejecución en  las funciones de los API's y retornar apropiadamente el JSON de respuesta. Además, se añade funciones de formateo de respuesta para que se pueda controlar el estándar de esta dentro de los API construidos en la organización.*

#### FUNCIONES:
 ##### SetResponseFormat: 
*Esta función permite retornar datos dentro de la ejecución normal de un proceso en especifico con un formato preestablecido. 
	 Como esta función realiza el envio implicito de la respuesta en formato JSON, al llamarla se ignora el metodo c.ServeJSON(), el siguiente es un ejemplo de llamado y respuesta del API usando este metodo:*
```go
import "github.com/udistrital/utils_oas/responseformat"

func (c *Controller)testService() {
	data := make(map[string]interface{}) // Can be any Go Data Type or Primitive.
	// Do some actions with data elmnt...
	responseformat.SetResponseFormat(&c.Controller, data, "Some response code", 200)
	
}
```
*La respuesta del API para el anterior ejemplo seria la siguiente:*
```json
// Status=200
{
	"Type": "success",
	"Code": "Some response code",
	"Body": {
		"Some":"Data"
	} 
}
```

*Los Parametros que se deben ingresar a la función son:*
*  ***beego.Controller:** Apuntador a la instancia padre de un controlador de Beego. Se debe pasar en el formato mostrada en el ejemplo siempre, debido a que la librería debe acceder directamente al contexto del llamado al API.
* **data:** *Son los datos que retorna el proceso llamado en el servicio. Puede ser cualquier tipo de dato Go o sus primitivas, esta data será la que reciba el cliente como resultado de una operación.*
*  **code:** *Este elemento puede ser enviado vacío (""). Representa un código único de operación para que el cliente pueda traducirlo y mostrar el resultado de la operación en el formato que lo pidan los requerimientos.*
*  **status:** Representa el código http resultado de la operación, este también determina el elemento Type dentro del formato de la respuesta.

 ##### GlobalResponseHandler: 
*Esta función permite interceptar las respuestas de los controladores y dar formato a estas con el estándar definido. Esto permite que se puedan controlar los formatos de respuesta de los controladores auto generados por los frameworks sin hacer grandes modificaciones y también permite generar respuestas a los errores internos del servidor sin capturarlas de forma local en cada función. El siguiente es un ejemplo de implementación de este método:*
```go
package main
import (
"github.com/astaxie/beego"
"github.com/astaxie/beego/context"
"github.com/udistrital/utils_oas/responseformat"
)
func  main() {
beego.BConfig.RecoverFunc = responseformat.GlobalResponseHandler
// Some Code in main.go ...
}
```
*La respuesta de un error interno (panic) seria la siguiente:*
```json
// Status=500
{
	"Type": "error",
	"Code": "some string code or message",
	"Body": {
		"Some":"Data"
	} 
}
```

**Para que los controladores auto generados y en general todos los controladores usen de forma implícita este formato, es necesario quitar la expresión  c.ServeJSON() de la siguiente manera:**

```go

func (j *Controller) FullArbolRubro() {
	params := j.GetString(":someParams")
	data := dataHelper.ProcessData(params)
	j.Data["json"] = data

}
```
*El anterior código generaría  la siguiente repuesta:*

```json
// Status=200
{
	"Type": "success",
	"Code": "",
	"Body": {
		"Some":"Data"
	} 
}
```

*Por definir:*
- [ ] Uso de los códigos de error
