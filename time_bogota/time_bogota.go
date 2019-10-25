package time_bogota

import (
	"fmt"
	"time"
	"strings"

	"github.com/astaxie/beego/logs"
)

var tiempoBogota time.Time

func Tiempo_bogota() time.Time {
	fmt.Println("tiempo antes de correccion")
	tiempoBogota = time.Now()
	logs.Info(tiempoBogota)

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(loc)
	tiempoBogota = tiempoBogota.In(loc)
	fmt.Println("tiempo despues de correccion")
	logs.Info(tiempoBogota)
	return tiempoBogota
}

func TiempoBogotaFormato() string {
	fmt.Println("tiempo con formato")
	var tiempoFormato = Tiempo_bogota().Format(time.RFC3339Nano)
	logs.Info(tiempoFormato)
	return tiempoFormato
}

func TiempoCorreccionFormato(inputDate string) string {
	inputDate = strings.ToLower(inputDate)
	inputDate = strings.Replace(inputDate, " +0000 +0000", "", -1)
	inputDate = strings.Replace(inputDate, " ", "T", -1)
	inputDate = inputDate + "Z"
	return inputDate
}
