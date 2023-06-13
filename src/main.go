package main

import (
	"DAQDriver/src/web"
	"net/http"
	"time"
)

const addr string = ":3000"

func main() {

	routers := web.InitRouter()

	server := &http.Server{
		Addr:         addr,
		Handler:      routers,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server.ListenAndServe()
	/*
		var coreController CoreController = *newCoreController()

		SampleNumber, _ := strconv.ParseInt(os.Args[1], 10, 64)
		socketCommand := os.Args[2]

		if socketCommand == "socket" {
			coreController.socket.start()
		}

		coreController.avgData(SampleNumber)
		time.Sleep(time.Duration(2000) * time.Millisecond)
		coreController.avgData(SampleNumber)
		time.Sleep(time.Duration(2000) * time.Millisecond)
	*/
}
