package driver

import (
	spictl "DAQDriver/src/driver/spi"
	"math"
	"time"
)

type DACInfo struct {
	line          uint32
	devnum        int64
	baseAddress   uint32
	voltageRange  float64
	step          uint32
	offsetVoltage float64
	voltageNow    float64
}

type DACReg struct {
	name  string
	index int64
}

type DACController interface {
	initialize()
	regDACPort(string, string, int64)
	setDACVoltage(string, float64) float64
	getDACVoltage(string) float64
	setDACOffset(string, float64)
}

type DACControllerInstance struct {
	spi         spictl.SPIController
	DeviceInfo  []DACInfo
	DACRegister map[string]DACReg
}

func (dacctl *DACControllerInstance) initialize() {

	LVDACline := dacctl.spi.RegisterSlave(spictl.SlaveInfo{Name: "LVDAC", Offset: 0, CPHA: 1})
	HVDACline := dacctl.spi.RegisterSlave(spictl.SlaveInfo{Name: "HVDAC", Offset: 1, CPHA: 0})
	dacctl.DACRegister = make(map[string]DACReg)
	dacctl.DeviceInfo = []DACInfo{
		{0, LVDACline, 0x08, 5e9, 76972, 0, 0},
		{1, LVDACline, 0x08, 5e9, 76972, 0, 0},
		{2, LVDACline, 0x08, 5e9, 76972, 0, 0},
		{3, LVDACline, 0x08, 5e9, 76972, 0, 0},
		{0, HVDACline, 0x30, 2.5e9, 76100, 0, 0},
		{1, HVDACline, 0x30, 10e9, 304400, 0, 0},
		{2, HVDACline, 0x30, 10e9, 304400, 0, 0},
		{3, HVDACline, 0x30, 10e9, 304400, 0, 0}}

	dacctl.spi.Send(1, ((0x60+0x00)<<16 + 0x04))
	time.Sleep(time.Duration(1) * time.Millisecond)
	dacctl.spi.Send(1, ((0x60+0x01)<<16 + 0x03))
	time.Sleep(time.Duration(1) * time.Millisecond)
	dacctl.spi.Send(1, ((0x60+0x02)<<16 + 0x03))
	time.Sleep(time.Duration(1) * time.Millisecond)
	dacctl.spi.Send(1, ((0x60+0x03)<<16 + 0x03))

}

func (dacctl *DACControllerInstance) regDACPort(name string, device string, line int64) {
	index := int64(0)
	if device == "HVDAC" {
		index = index + 4
	}
	index = index + line

	dacctl.DACRegister[name] = DACReg{name, index}
}

func (dacctl *DACControllerInstance) setDACVoltage(name string, voltage float64) float64 {

	dev, ok := dacctl.DACRegister[name]
	if !ok {
		return -1
	}

	channelInfo := dacctl.DeviceInfo[dev.index]
	if math.Abs(voltage) > channelInfo.voltageRange {
		return -1
	}
	number := uint32(0)
	numberTmp := int32(((voltage - channelInfo.offsetVoltage) / float64(channelInfo.step)) + float64(32768))

	if numberTmp < 0 {
		numberTmp = 0
	} else if numberTmp > 65535 {
		numberTmp = 65535
	}
	number = uint32(numberTmp)

	dacctl.spi.Send(channelInfo.devnum, (((channelInfo.baseAddress + channelInfo.line) << 16) + number))

	channelInfo.voltageNow = float64(number * channelInfo.step)

	return channelInfo.voltageNow

}

func (dacctl *DACControllerInstance) getDACVoltage(name string) float64 {

	dev, ok := dacctl.DACRegister[name]
	if !ok {
		return -1
	}
	return dacctl.DeviceInfo[dev.index].voltageNow
}

func (dacctl *DACControllerInstance) setDACOffset(name string, offsetVoltage float64) {

	dev, ok := dacctl.DACRegister[name]
	if !ok {
		return
	}
	dacctl.DeviceInfo[dev.index].offsetVoltage = offsetVoltage
}

func newDACController(SPIAddress int64) *DACController {

	var DACSPI spictl.SPIController = *spictl.NewSPIController(SPIAddress)
	var DACCTL DACController = &DACControllerInstance{spi: DACSPI}

	DACCTL.initialize()

	return &DACCTL
}
