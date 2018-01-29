package auditoria

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/streadway/amqp"
)

//Función que envía el msj a la cola correspondiente
func sentToRabbit(msj string) {

	//p := beego.AppConfig.Strings("RABBIT_MQ_URI")

	//Conexión RabbitMQ Server
	con, err := amqp.Dial("amqp://guest:guest@10.20.0.175:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer con.Close()

	//connection = con

	chanel, err := con.Channel()
	failOnError(err, "Failed to open a channel")
	defer chanel.Close()

	//chl = chanel
	//Cola a la que se enviará el msj
	cha := beego.AppConfig.Strings("RABBIT_MQ_CHANNEL")

	fmt.Println(cha)

	q, err := chanel.QueueDeclare(
		cha[0], // name
		true,   // durable
		false,  // delete when usused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	beego.Info(q)
	failOnError(err, "Failed to declare a queue")

	body := msj
	err = chanel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
	failOnError(err, "Failed to publish a message")
}
