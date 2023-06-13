package web

import (
	"DAQDriver/src/driver"
	"time"
)

var device driver.CoreControllerInterface = &driver.CoreController{}

type deviceProgHeatingStep driver.ProgHeatingStep

func init() {

	setBitfile("")
	time.Sleep(time.Duration(200) * time.Millisecond)
	device.Initialize()
	time.Sleep(time.Duration(200) * time.Millisecond)
	device.ConnectADC()
	time.Sleep(time.Duration(200) * time.Millisecond)
	device.StartFetchData()
	time.Sleep(time.Duration(200) * time.Millisecond)
	getAvgVoltage(50000)

}
