package driver

import (
	"fmt"
	"net"
)

//const port = int64(4567)

type socketController struct {
	listen       net.Listener
	connect      net.Conn
	socketDataCH chan socketCH
}

func (sckctl *socketController) addListener() {

	listen, err := net.Listen("tcp", "0.0.0.0:4567")

	sckctl.listen = listen

	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return
	}

}

func (sckctl *socketController) start(result chan<- bool) {

	if sckctl.listen == nil {
		sckctl.addListener()
	}

	for {

		conn, err := sckctl.listen.Accept()
		fmt.Print("Remote socket address: " + conn.RemoteAddr().String() + "\n")

		sckctl.connect = conn

		if err != nil {
			fmt.Printf("accept failed, err:%v\n", err)
			result <- false
			return
		} else {
			go sendDataSocket(conn, sckctl.socketDataCH)
			result <- true
			return
		}

	}
}

func (sckctl *socketController) stop() {

	socketCHSign = 0
	if sckctl.connect != nil {
		sckctl.connect.Close()
	}

}

func (sckctl *socketController) setCH(socketDataCH chan socketCH) {
	sckctl.socketDataCH = socketDataCH
}

func newSocketController() *socketController {
	var SKTCTL = &socketController{}
	SKTCTL.addListener()
	return SKTCTL
}
