package time_bogota

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/logs"
)

func tiempo_bogota() time.Time {
	fmt.Println("tiempo antes de correccion")
	var tiempoBogota = time.Now()
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

func tiempoBogotaFormato() string {
	fmt.Println("tiempo con formato")
	var tiempoFormato = tiempo_bogota().Format(time.RFC3339Nano)
	logs.Info(tiempoFormato)
	return tiempoFormato
}
