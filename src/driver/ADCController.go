package driver

import (
	axictl "DAQDriver/src/driver/axi"
	spictl "DAQDriver/src/driver/spi"
	"time"
)

const activeCommand uint32 = 0x0F01

type commandInfo struct {
	name    string
	address uint32
	rawcode uint32
	opcode  uint32
}

var setupCommand = []commandInfo{
	{"reset", 0x00, 0x18, 0x99},
	{"powermode", 0x02, 0xD0, 0xD1},
	{"refgain", 0x10, 0x00, 0x02},
	{"datach", 0x12, 0x00, 0x02},
	{"ODRintlsb", 0x16, 0x40, 0xF0},
	{"filter", 0x1E, 0x00, 0x55},
	{"dataio", 0x11, 0x00, 0x33}}

var vosSetupCommand = []commandInfo{
	{"vosch0LSB", 0x2A, 0x00, 0x57},
	{"vosch0MID", 0x2B, 0x00, 0xFF},
	{"vosch0MSB", 0x2C, 0x00, 0xFF},
	{"vosch1LSB", 0x30, 0x00, 0xCA},
	{"vosch1MID", 0x31, 0x00, 0xFF},
	{"vosch1MSB", 0x32, 0x00, 0xFF}}

type ADCController interface {
	initialize()
	exeCommand([]commandInfo)
	read() [4]float64
	readADC() [2]uint64
	setVos(int64, float64)
}

type ADCControllerInstance struct {
	spi    spictl.SPIController
	axi    axictl.AXIBusController
	devnum int64
	vos    [4]float64
	gain   [4]float64
}

func (adcctl *ADCControllerInstance) exeCommand(command []commandInfo) {

	for i := 0; i < len(command); i++ {
		adcctl.spi.Send(adcctl.devnum, ((command[i].address)<<8)+command[i].opcode)
		time.Sleep(time.Duration(100) * time.Microsecond)
		adcctl.spi.Send(adcctl.devnum, activeCommand)
		time.Sleep(time.Duration(10) * time.Millisecond)
	}

}

func (adcctl *ADCControllerInstance) initialize() {

	ADCDeviceInfo := spictl.SlaveInfo{Name: "ADC", Offset: 0, CPHA: 0}
	adcctl.devnum = adcctl.spi.RegisterSlave(ADCDeviceInfo)

	adcctl.exeCommand(setupCommand)
	adcctl.exeCommand(vosSetupCommand)

	adcctl.vos = [4]float64{0, 0, 0, 0}
	adcctl.gain = [4]float64{-0.268*1.997, 4.96*1.997, 10*1.997, -10*1.997}

}

//var tmp = [2]uint64{0, 0}
var dataTempList = [4]uint32{0, 0, 0, 0}
var readResult = [4]float64{0, 0, 0, 0}
var orderFlagReg = uint32(0)

//var orderFlag uint32

//var n uint32
//var dataTemp uint32

func (adcctl *ADCControllerInstance) read() [4]float64 {

	for {

		tmp := adcctl.axi.Data8B

		dataTempList[0] = uint32(tmp[1] >> 32)
		dataTempList[1] = uint32(tmp[1] & 0xFFFFFFFF)
		dataTempList[2] = uint32(tmp[0] >> 32)
		dataTempList[3] = uint32(tmp[0] & 0xFFFFFFFF)

		orderFlag := (dataTempList[0] >> 25) & 0x1F

		if orderFlag != orderFlagReg {
			orderFlagReg = orderFlag

			for n := 0; n <= 3; n++ {

				dataTemp := dataTempList[n]
				if ((dataTemp >> 24) & 0x01) == 1 {
					dataTemp = dataTemp | uint32(0xFF000000)
				} else {
					dataTemp = dataTemp & uint32(0x00FFFFFF)
				}

				readResult[n] = float64(int32(dataTemp))*488/adcctl.gain[n] - adcctl.vos[n]
			}

			break

		}

	}
	return readResult

}

func (adcctl *ADCControllerInstance) readADC() [2]uint64 {

	return *adcctl.axi.Data8B
}

func (adcctl *ADCControllerInstance) setVos(chnum int64, vosnum float64) {
	adcctl.vos[chnum] = vosnum
}

func newADCController(SPIAddress int64, AXIAddress int64) *ADCController {

	var ADCSPI spictl.SPIController = *spictl.NewSPIController(SPIAddress)
	var ADCAXI axictl.AXIBusController = *axictl.NewAXIController(AXIAddress)
	var ADCCTL ADCController = &ADCControllerInstance{spi: ADCSPI, axi: ADCAXI}

	ADCCTL.initialize()

	return &ADCCTL
}
