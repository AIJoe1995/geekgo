package main

import "geekgo/week9/webook/ioc"

func main() {
	//server := InitServer()
	//server.Run(":8080")
	ioc.InitPrometheus()
	ioc.InitGinPrometheus()
	ioc.InitKafkaPromethues()

	app := InitServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	app.web.Run(":8080")

}
