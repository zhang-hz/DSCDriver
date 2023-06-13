package driver

import (
	pidctl "DAQDriver/src/driver/pid"
	"fmt"
	"time"
)

/**
*
* 获取ADC平均电压
*
**/

type avgVoltageData struct {
	Voltage [4]float64 //平均电压
	Time    float64    //耗时
}

func (corectl *CoreController) AvgVoltage(SampleNumber int64) avgVoltageData {

	var result avgVoltageData
	rchan := make(chan avgVoltageData, 1)
	go avgVoltageExec(rchan, SampleNumber, corectl.dchhelper1)
	result = <-rchan
	close(rchan)

	return result
}

func avgVoltageExec(dout chan<- avgVoltageData, SampleNumber int64, din <-chan DAQDataCH) {

	fmt.Print("Start avg Data\n")
	var avg = [4]float64{0, 0, 0, 0}

	var i int64
	var n int64
	var data DAQDataCH

	for len(din) > 0 {
		<-din
	}

	helperCHSign = helperCHSign | 0x1

	for i = 0; i < dataChDepth+100; i++ {
		data = <-din
	}

	startTime := time.Now().UnixNano()

	for i = 0; i < SampleNumber; i++ {
		data = <-din
		for n = 0; n < 4; n++ {
			avg[n] = avg[n] + data.directv[n]
		}
	}
	helperCHSign = helperCHSign & 0xFE
	endTime := time.Now().UnixNano()

	for i = 0; i < 4; i++ {
		avg[i] = avg[i] / float64(SampleNumber)
	}

	result := avgVoltageData{
		Voltage: [4]float64{avg[0], avg[0] - avg[1], avg[2], avg[3]},
		Time:    float64(endTime - startTime),
	}

	dout <- result

	//wg.Done()
}

/**
*
*  设定ADC电压偏置
*
**/

func (corectl *CoreController) SetADCVos(ADCCH int64, VosNumber float64) {

	corectl.adc.setVos(ADCCH, VosNumber)

}

/**
*
*  设定DAC电压
*
**/

func (corectl *CoreController) SetDACVoltage(DACPort string, Voltage float64) float64 {

	voltageNow := corectl.dac.setDACVoltage(DACPort, Voltage)
	if DACPort == "TP1" {
		corectl.heater.voltage[0] = voltageNow
	} else if DACPort == "TP2" {
		corectl.heater.voltage[1] = voltageNow
	}

	return voltageNow

}
func (corectl *CoreController) GetDACVoltage(DACPort string) float64 {

	return corectl.dac.getDACVoltage(DACPort)

}

/**
*
*  设定DAC电压补偿
*
**/
func (corectl *CoreController) SetDACOffset(DACPort string, offsetVoltage float64) {

	corectl.dac.setDACOffset(DACPort, offsetVoltage)
}

/**
*
*  启动程序温度控制器
*
**/

func (corectl *CoreController) StartProgramHeater(count int64, basevoltage float64, baseTemperature float64, processStep []ProgHeatingStep) {

	corectl.tmp.startProgHeater(count, basevoltage, baseTemperature, processStep)

}

/**
*
*  停止程序温度控制器
*
**/

func (corectl *CoreController) StopProgramHeater() {

	corectl.tmp.stopProgHeater()
}

/**
*
*  启动静态温度控制器
*
**/

func (corectl *CoreController) StartStaticHeater(basevoltage float64, targetTemperature float64) {

	corectl.tmp.startStaticHeater(basevoltage, targetTemperature)

}

/**
*
*  停止静态温度控制器
*
**/

func (corectl *CoreController) StopStaticHeater() {

	corectl.tmp.stopStaticHeater()
}

/**
*
*  设定静态温度控制器
*
**/

func (corectl *CoreController) SetupStaticTemperature(temperature float64) {

	corectl.tmp.setupStaticTemperature(temperature)
}

/**
*
*  设定温度控制器参数
*
**/

func (corectl *CoreController) SetupHeaterPIDParameter(kp float64, ki float64, kd float64, tolerance float64, errorTolerance float64, initialValue float64) {

	pidsetting := pidctl.PIDSetting{
		Kp:             kp,
		Ki:             ki,
		Kd:             kd,
		Interval:       20000 * float64(heaterDownSampleRate),
		Tau:            20000 * float64(heaterDownSampleRate),
		Tolerance:      tolerance,
		ErrorTolerance: errorTolerance,
		InitialValue:   initialValue,
		LimitMin:       0,
		LimitMax:       DSCChipMaxHeatingVoltage,
		LimitMinIntg:   0,
		LimitMaxIntg:   0,
	}
	corectl.tmp.setupHeaterPID(pidsetting)
}
func (corectl *CoreController) SetupCompensatorPIDParameter(kp float64, ki float64, kd float64, tolerance float64, errorTolerance float64, initialValue float64) {

	pidsetting := pidctl.PIDSetting{
		Kp:             kp,
		Ki:             ki,
		Kd:             kd,
		Interval:       20000 * float64(heaterDownSampleRate),
		Tau:            20000 * float64(heaterDownSampleRate),
		Tolerance:      tolerance,
		ErrorTolerance: errorTolerance,
		InitialValue:   initialValue,
		LimitMin:       -2.5e9,
		LimitMax:       2.5e9,
		LimitMinIntg:   0,
		LimitMaxIntg:   0,
	}
	corectl.tmp.setupCompensatorPID(pidsetting)
}

//启动/停止在程序升温/控温中的功率补偿

func (corectl *CoreController) SetupCompensator(isStart float64) {

	corectl.tmp.setupCompensator(isStart)
}

//开启手动控温下的手动功率补偿

func (corectl *CoreController) StartManualCompensator() {
	corectl.tmp.startManualCompensator()
}

//停止手动控温下的手动功率补偿

func (corectl *CoreController) StopManualCompensator() {
	corectl.tmp.stopManualCompensator()
}
