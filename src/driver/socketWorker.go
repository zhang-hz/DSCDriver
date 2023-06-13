package driver

import (
	"DAQDriver/src/driver/colfer"
	"fmt"
	"net"
	"time"
)

type socketCH struct {
	time     time.Time
	interval float64
	length   int64
	directv  [4][socketDataLength]float64
	diffv    [socketDataLength]float64
	heaterv  [3][socketDataLength]float64
	heaterp  [3][socketDataLength]float64
}

func sendDataSocket(conn net.Conn, din <-chan socketCH) {

	var dtmp socketCH

	fmt.Print("Socket Worker: Start sending data \n")

	socketCHSign = 2

	for len(din) > 0 {
		<-din
	}
	socketCHSign = 1

	for i := 0; i < 5; i++ {
		<-din
	}

	for {

		if socketCHSign == 0 {
			fmt.Print("Socket Worker: Stop sending data \n")
			return
		}

		dtmp = <-din

		sendData := &colfer.Colfbuf{
			Time:        dtmp.time,
			Interval:    dtmp.interval,
			Datalength:  dtmp.length,
			Directvch0:  dtmp.directv[0][:socketDataLength-1],
			Directvch1:  dtmp.directv[1][:socketDataLength-1],
			Directvch2:  dtmp.directv[2][:socketDataLength-1],
			Directvch3:  dtmp.directv[3][:socketDataLength-1],
			Diffv:       dtmp.diffv[:socketDataLength-1],
			Heatervch0:  dtmp.heaterv[0][:socketDataLength-1],
			Heatervch1:  dtmp.heaterv[1][:socketDataLength-1],
			Diffheaterv: dtmp.heaterv[2][:socketDataLength-1],
			Heaterpch0:  dtmp.heaterp[0][:socketDataLength-1],
			Heaterpch1:  dtmp.heaterp[1][:socketDataLength-1],
			Diffheaterp: dtmp.heaterp[2][:socketDataLength-1],
		}
		sendBuf, _ := sendData.MarshalBinary()
		//fmt.Print(len(sendBuf))
		_, err := conn.Write(sendBuf)
		time.Sleep(time.Duration(45) * time.Millisecond)

		if err != nil {
			fmt.Print("Socket Worker: ERROR: Stop send data  ")
			fmt.Print(err)
			fmt.Print("\n")
			socketCHSign = 0
			return
		}
	}
}
