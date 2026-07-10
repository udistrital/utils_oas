package timebogota

import (
	"time"
)

var bogota, _ = time.LoadLocation("America/Bogota")

func TiempoBogota() time.Time {
	tiempoBogota := time.Now()
	tiempoBogota = tiempoBogota.In(bogota)

	return tiempoBogota
}
