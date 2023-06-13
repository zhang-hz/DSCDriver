package web

import (
	"DAQDriver/src/driver"
	"fmt"
	"os/exec"
)

func setBitfile(filename string) bool {

	command := exec.Command("python3", "./setuppl.py", filename)
	err := command.Run()

	return err == nil
}

type avgVoltageWeb struct {
	Voltage [4]float64
}

func getAvgVoltage(SampleNumber int64) avgVoltageWeb {

	av := device.AvgVoltage(SampleNumber)

	var result = &avgVoltageWeb{
		Voltage: av.Voltage,
	}

	fmt.Print("---------------CH0-----------------\n")
	fmt.Print(av.Voltage[0]/float64(1e6), " mV\n")
	fmt.Print(SampleNumber, "\n")
	fmt.Print("---------------CH1-----------------\n")
	fmt.Print(av.Voltage[1]/float64(1e6), " mV\n")
	fmt.Print(SampleNumber, "\n")
	fmt.Print("---------------CH2-----------------\n")
	fmt.Print(av.Voltage[2]/float64(1e6), " mV\n")
	fmt.Print(SampleNumber, "\n")
	fmt.Print("---------------CH3-----------------\n")
	fmt.Print(av.Voltage[3]/float64(1e6), " mV\n")
	fmt.Print(SampleNumber, "\n")
	fmt.Print("-------------Time------------------\n")
	fmt.Print(float64(av.Time/1e6), " ms\n")

	return *result

}

func setADCVosLocal(ADCCH int64, VosNumber float64) {
	device.SetADCVos(ADCCH, VosNumber)
}

func setDACVoltageLocal(DACport string, Voltage float64) float64 {
	return device.SetDACVoltage(DACport, Voltage)
}

func getDACVoltageLocal(DACport string) float64 {
	return device.GetDACVoltage(DACport)
}

func setDACOffsetLocal(DACport string, offsetVoltage float64) {
	device.SetDACOffset(DACport, offsetVoltage)
}

func startHeaterProgramPIDLocal(baseVoltage float64, heatingSpeed float64, coolSpeed float64, maxTemperature float64, baseTemperature float64) {

	step := make([]driver.ProgHeatingStep, 3)
	step = append(step, driver.ProgHeatingStep{
		ProcessType:      "heat",
		StartTemperature: baseTemperature,
		StopTemperature:  maxTemperature,
		ProcessSpeed:     heatingSpeed,
		ProcessTime:      0,
	}, driver.ProgHeatingStep{
		ProcessType:      "hold",
		StartTemperature: maxTemperature,
		StopTemperature:  maxTemperature,
		ProcessSpeed:     0,
		ProcessTime:      10,
	}, driver.ProgHeatingStep{
		ProcessType:      "cool",
		StartTemperature: maxTemperature,
		StopTemperature:  baseTemperature,
		ProcessSpeed:     coolSpeed,
		ProcessTime:      0,
	})

	device.StartProgramHeater(int64(3), baseVoltage, baseTemperature, step)
}

func startHeaterProgramAdvPIDLocal(count int64, baseVoltage float64, baseTemperature float64, process []HeatProcess) {

	step := make([]driver.ProgHeatingStep, len(process))
	for i := 0; i < len(process); i++ {
		step[i] = driver.ProgHeatingStep{
			ProcessType:      process[i].ProcessType,
			StartTemperature: process[i].StartTemperature,
			StopTemperature:  process[i].StopTemperature,
			ProcessSpeed:     process[i].ProcessSpeed / 1e6,
			ProcessTime:      process[i].ProcessTime * 1e6,
		}
	}

	device.StartProgramHeater(count, baseVoltage, baseTemperature, step)
}

func stopHeaterProgramPIDLocal() {
	device.StopProgramHeater()
}

func startHeaterStaticPIDLocal(temperature float64, baseVoltage float64) {
	device.StartStaticHeater(baseVoltage, temperature)
}

func stopHeaterStaticPIDLocal() {
	device.StopStaticHeater()
}

func setupHeaterTemperaturePIDLocal(temperature float64) {
	device.SetupStaticTemperature(temperature)
}

func setupHeaterPIDParameterLocal(kp float64, ki float64, kd float64, tolerance float64, errorTolerance float64, initialValue float64) {
	device.SetupHeaterPIDParameter(kp, ki, kd, tolerance, errorTolerance, initialValue)
}

func setupCompensatorPIDParameterLocal(kp float64, ki float64, kd float64, tolerance float64, errorTolerance float64, initialValue float64) {
	device.SetupCompensatorPIDParameter(kp, ki, kd, tolerance, errorTolerance, initialValue)
}

func setupAutoCompensatorLocal(isStart float64) {
	device.SetupCompensator(isStart)
}

func startManualCompensatorLocal() {
	device.StartManualCompensator()
}

func stopManualCompensatorLocal() {
	device.StopManualCompensator()
}
