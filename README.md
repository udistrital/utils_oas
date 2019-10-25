# utils_oas

Este es un paquete de librerías y utilidades generales para las aplicaciones desarrolladas en el framework beego que hacen parte de la arquitectura API RES de la OAS:

- **apiStatusLib**

- **customerror**

  Funcionalidades para establecer el [código de estado HTTP](https://es.wikipedia.org/wiki/Anexo:C%C3%B3digos_de_estado_HTTP) de cada una de las respuestas del API  y establecer la estructura de respuesta en formato JSON.

- **formatdata**

  Funcionalidades para la conversión y trabajo de estructuras JSON

- **optimize**

  Funcionalidades de optimización con procesamiento de datos en golang

- **request**

  Funcionalidades para el consumo de servicios JSON desde una API

- **ruler**

  Funcionalidades para las reglas de negocio

- **security**

  Funcionalidades de seguridad para aplicaciones legadas, híbridas y nuevas en go
- <details>
    <summary><b>time_bogota</b></summary>

    importar:

    ```go
    "github.com/udistrital/utils_oas/time_bogota"
    ```

    3 funcinalidades: 

    - Tiempo_bogota : 
     Da la hora de Bogota sin importar la zona horaria de la maquina o contenedor

        ***usar en codigo (remplarar)***

        ```go
        VariableDeTiempo = tiem.Now()
        ```
        por

        ```go
        VariableDeTiempo = time_bogota.Tiempo_bogota()
        ```

    - TiempoBogotaFormato()
        ***(Nota : esta funcion funciona perfectamente en peticiones POST, para los put puede mandar lio asi que se recomienda usar para los PUT la tercera funcion aqui nombrada)***
        
        Esta funcion da el formato para la hora y que esta sea aceptada por la base de datos.

        ya que esta funcion retorna un string, se debe cambiar en los modelos del api donde se quiera usar la funcion, esto evitara problemas con la hora y que genere una hora con UTC 0

        ***en codigo***

        ```go
        type ResolucionEstado struct {
            Id            int
            FechaRegistro time.Time
            Usuario       string
            Estado        *EstadoResolucion
            Resolucion    *Resolucion
        }
        ```

        por

        ```go
        type ResolucionEstado struct {
            Id            int
            FechaRegistro string
            Usuario       string
            Estado        *EstadoResolucion
            Resolucion    *Resolucion
        }
        ```
        ---
        ```go
        VariableDeTiempo = tiem.Now()
        ```
        por

        ```go
        VariableDeTiempo = time_bogota.TiempoBogotaFormato()
        ```
    - TiempoCorreccionFormato(inputDate string) : esta funcion recibe un string y devuelve otro transformado, esta funcion surge como solucion al problema de que las fechas al traerlas de la base de datos pueden llegar en el siguiente formato `2019-10-08 18:26:45.58 +0000 +0000`, este formato al hacer un update en la base de datos provoca errores, por ende esta funcion realiza la correccion
    
	    para usarla se usara el siguiente ejemplo, suponga que de la base de datos trae una fecha y se llama `FechaFin` y su valor al imprimirlo es el siguiente : ``2019-10-08 18:26:45.58 +0000 +0000` para corregirlo realice lo siguiente:
  
	    ```go
	    FechaFin = time_bogota.TiempoCorreccionFormato(FechaFin)
	    ```
  
	    esto le devolvera la fecha en el siguiente formato : `2019-10-08T18:26:45.58Z` el cual la ase de datos recibira.



</details>

## Licencia

This file is part of utils_oas.

utils_oas is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Foobar is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Foobar.  If not, see <https://www.gnu.org/licenses/>.
