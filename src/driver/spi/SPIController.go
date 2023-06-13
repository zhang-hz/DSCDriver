package spi

import (
	axictl "DAQDriver/src/driver/axi"
	"time"
)

//Register Default Vaule
const defaultSPICR uint32 = 0b0000000100

//Register Offset Address
const offsetSPICR uint32 = 0x60

//const offsetSPISR uint32 = 0x64
const offsetSPIDTR uint32 = 0x68
const offsetSPISSR uint32 = 0x70

type SlaveInfo struct {
	Name   string
	Offset uint32
	CPHA   uint32
}

type SPIController struct {
	SPE        uint32
	SLAVE      int64
	CPHA       uint32
	CR         uint32
	slaveList  [10]SlaveInfo
	slaveRange int64
	mmio       axictl.AXIBusController
}

func (spictl *SPIController) initialize() {

	spictl.SPE = 0
	spictl.SLAVE = 0
	spictl.CPHA = 0
	spictl.CR = defaultSPICR
	spictl.mmio.Write(offsetSPICR, defaultSPICR)
	spictl.mmio.Write(offsetSPISSR, uint32(0xFFFFFFFE))

}

func (spictl *SPIController) reset() {

	spictl.SPE = 0
	spictl.SLAVE = 0
	spictl.CPHA = 0
	spictl.CR = defaultSPICR
	spictl.mmio.Write(offsetSPICR, defaultSPICR)
	spictl.mmio.Write(offsetSPISSR, uint32(0xFFFFFFFE))

}

func (spictl *SPIController) selectSlave(slaveNum int64) {

	offset := spictl.slaveList[slaveNum].Offset
	CPHA := spictl.slaveList[slaveNum].CPHA
	spictl.SLAVE = slaveNum

	if spictl.CPHA != CPHA {
		if CPHA == 1 {
			spictl.CR = spictl.CR | 0x10
			spictl.mmio.Write(offsetSPICR, spictl.CR)
		} else {
			spictl.CR = spictl.CR & 0xFFFFFFEF
			spictl.mmio.Write(offsetSPICR, spictl.CR)
		}

	}

	spictl.mmio.Write(offsetSPISSR, ((uint32(0xFFFFFFEF)<<offset)>>4)&uint32(0xFFFFFFFF))

}

func (spictl *SPIController) enable() {

	spictl.CR = spictl.CR | uint32(0b0000000010)
	spictl.mmio.Write(offsetSPIDTR, uint32(0xFFEFFEFF))
	spictl.mmio.Write(offsetSPICR, spictl.CR)
	time.Sleep(time.Duration(1) * time.Millisecond)
	spictl.SPE = 1

}

func (spictl *SPIController) disable() {

	spictl.CR = spictl.CR & 0b1111111101
	spictl.mmio.Write(offsetSPICR, spictl.CR)
	time.Sleep(time.Duration(1) * time.Millisecond)
	spictl.SPE = 0

}

func (spictl *SPIController) checkCR() {

	CR := spictl.mmio.Read(offsetSPICR)
	if CR&0b010 == 2 {
		spictl.SPE = 1
	} else {
		spictl.SPE = 0
	}

}

func (spictl *SPIController) RegisterSlave(slave SlaveInfo) int64 {

	spictl.slaveList[spictl.slaveRange] = slave
	spictl.slaveRange++
	return spictl.slaveRange - 1

}

func (spictl *SPIController) Send(slave int64, data uint32) {

	if spictl.SLAVE != slave {
		spictl.selectSlave(slave)
		spictl.SLAVE = slave
	}

	if spictl.SPE != 1 {
		spictl.enable()
		//fmt.Print("Enable SPI \n")
	}

	//fmt.Print("Write to SPI: ", "Address ", offsetSPIDTR, ", Data ", data, "\n")
	spictl.mmio.Write(offsetSPIDTR, data)

}

func NewSPIController(deviceAddress int64) *SPIController {
	var newAXI axictl.AXIBusController = *axictl.NewAXIController(deviceAddress)
	var newSPI *SPIController = &SPIController{mmio: newAXI}
	newSPI.initialize()

	return newSPI

}
