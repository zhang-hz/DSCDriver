package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HeatProcess struct {
	ProcessType      string  `json:"processtype"`
	StartTemperature float64 `json:"starttemperature"`
	StopTemperature  float64 `json:"stoptemperature"`
	ProcessSpeed     float64 `json:"speed"`
	ProcessTime      float64 `json:"processtime"`
}

func connectPC(c *gin.Context) {
	c.Status(http.StatusOK)
}

func connectSocket(c *gin.Context) {
	result := device.ConnectSocket()
	if result {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusBadRequest)
	}
}

func disconnectSocket(c *gin.Context) {
	device.DisconnectSocket()
	c.Status(http.StatusOK)
}

//向PL上传bitfile
func setPLBitFile(c *gin.Context) {

	filename := c.Param("bitfilename")
	result := setBitfile(filename)
	if result {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusBadRequest)
	}

}

func avgVoltage(c *gin.Context) {

	SampleNumber, _ := strconv.ParseInt(c.Param("samplenumber"), 10, 64)
	result := getAvgVoltage(SampleNumber)

	c.JSON(http.StatusOK, result)

}

func setADCVos(c *gin.Context) {

	ADCCH, _ := strconv.ParseInt(c.Param("adcch"), 10, 64)
	vosnumber, _ := strconv.ParseFloat(c.Param("vosnumber"), 64)
	fmt.Println("Web API: setADCVos: ADCCH:", ADCCH, "Vosnumber:", vosnumber)
	setADCVosLocal(ADCCH, vosnumber)

	c.Status(http.StatusOK)

}

func setDACVoltage(c *gin.Context) {

	DACPort := c.Param("dacport")
	Voltage, _ := strconv.ParseFloat(c.Param("voltage"), 64)
	fmt.Println("Web API: setDACVoltage: DACPort:", DACPort, "Voltage:", Voltage)
	result := setDACVoltageLocal(DACPort, Voltage)
	c.JSON(200, gin.H{"voltage": result})

}

func getDACVoltage(c *gin.Context) {

	DACPort := c.Param("dacport")
	result := getDACVoltageLocal(DACPort)
	c.JSON(200, gin.H{"voltage": result})

}

func setDACOffset(c *gin.Context) {

	DACPort := c.Param("dacport")
	Offset, _ := strconv.ParseFloat(c.Param("offset"), 64)
	fmt.Println("Web API: setDACOffset: DACPort:", DACPort, "Offset:", Offset)
	setDACOffsetLocal(DACPort, Offset)

	c.Status(http.StatusOK)

}

func startHeaterProgramPID(c *gin.Context) {

	baseVoltage, _ := strconv.ParseFloat(c.Param("basevoltage"), 64)
	heatingSpeed, _ := strconv.ParseFloat(c.Param("heatingspeed"), 64)
	coolSpeed, _ := strconv.ParseFloat(c.Param("coolspeed"), 64)
	maxTemperature, _ := strconv.ParseFloat(c.Param("maxtemperature"), 64)
	baseTemperature, _ := strconv.ParseFloat(c.Param("basetemperature"), 64)
	fmt.Println("Web API: startHeaterProgramPID: HeatingSpeed:", heatingSpeed, "basevoltage:", baseVoltage, "basetemperature:", baseTemperature)
	startHeaterProgramPIDLocal(baseVoltage, heatingSpeed, coolSpeed, maxTemperature, baseTemperature)
	fmt.Println("Web API: startHeaterProgramPID: Done")
	c.Status(http.StatusOK)
}

func startHeaterProgramAdvPID(c *gin.Context) {

	count, _ := strconv.ParseFloat(c.Param("count"), 64)
	baseVoltage, _ := strconv.ParseFloat(c.Param("basevoltage"), 64)
	baseTemperature, _ := strconv.ParseFloat(c.Param("basetemperature"), 64)

	json := []HeatProcess{}
	c.BindJSON(&json)

	fmt.Println("Web API: startHeaterProgramAdvanced: basevoltage:", baseVoltage, "basetemperature:", baseTemperature)
	startHeaterProgramAdvPIDLocal(int64(count), baseVoltage, baseTemperature, json)

	fmt.Println("Web API: startHeaterProgramAdvanced: Done")
	c.Status(http.StatusOK)
}

func stopHeaterProgramPID(c *gin.Context) {
	fmt.Println("Web API: stopHeaterProgramPID")
	stopHeaterProgramPIDLocal()
	c.Status(http.StatusOK)
}

func startHeaterStaticPID(c *gin.Context) {
	temperature, _ := strconv.ParseFloat(c.Param("temperature"), 64)
	baseVoltage, _ := strconv.ParseFloat(c.Param("basevoltage"), 64)
	fmt.Println("Web API: startHeaterStaticPID: temperature:", temperature, "basevoltage:", baseVoltage)
	startHeaterStaticPIDLocal(temperature, baseVoltage)
	fmt.Println("Web API: startHeaterStaticPID: Done")
	//c.Status(http.StatusOK)
}

func stopHeaterStaticPID(c *gin.Context) {
	fmt.Println("Web API: stopHeaterStaticPID")
	stopHeaterStaticPIDLocal()
	c.Status(http.StatusOK)
}

func setupTemperature(c *gin.Context) {
	temperature, _ := strconv.ParseFloat(c.Param("temperature"), 64)
	fmt.Println("Web API: setupTemperature: temperature:", temperature)
	setupHeaterTemperaturePIDLocal(temperature)
	c.Status(http.StatusOK)
}

func setupHeaterPIDParameter(c *gin.Context) {
	kp, _ := strconv.ParseFloat(c.Param("kp"), 64)
	ki, _ := strconv.ParseFloat(c.Param("ki"), 64)
	kd, _ := strconv.ParseFloat(c.Param("kd"), 64)
	tolerance, _ := strconv.ParseFloat(c.Param("tolerance"), 64)
	errorTolerance, _ := strconv.ParseFloat(c.Param("errorTolerance"), 64)
	initialValue, _ := strconv.ParseFloat(c.Param("initialValue"), 64)
	fmt.Println("Web API: setupHeaterPIDParameter: kp:", kp, "ki:", ki, "kd:", kd, "tolerance:", tolerance, "errorTolerance:", errorTolerance, "initialValue", initialValue)
	setupHeaterPIDParameterLocal(kp, ki, kd, tolerance, errorTolerance, initialValue)
	c.Status(http.StatusOK)
}

func setupCompensatorPIDParameter(c *gin.Context) {
	kp, _ := strconv.ParseFloat(c.Param("kp"), 64)
	ki, _ := strconv.ParseFloat(c.Param("ki"), 64)
	kd, _ := strconv.ParseFloat(c.Param("kd"), 64)
	tolerance, _ := strconv.ParseFloat(c.Param("tolerance"), 64)
	errorTolerance, _ := strconv.ParseFloat(c.Param("errorTolerance"), 64)
	initialValue, _ := strconv.ParseFloat(c.Param("initialValue"), 64)
	fmt.Println("Web API: setupCompensatorPIDParameter: kp:", kp, "ki:", ki, "kd:", kd, "tolerance:", tolerance, "errorTolerance:", errorTolerance, "initialValue", initialValue)
	setupCompensatorPIDParameterLocal(kp, ki, kd, tolerance, errorTolerance, initialValue)
	c.Status(http.StatusOK)
}

func startAutoCompensator(c *gin.Context) {

	fmt.Println("Web API: Start Auto Compensator")
	setupAutoCompensatorLocal(float64(1))
	c.Status(http.StatusOK)

}

func stopAutoCompensator(c *gin.Context) {

	fmt.Println("Web API: Stop Auto Compensator")
	setupAutoCompensatorLocal(float64(0))
	c.Status(http.StatusOK)

}

func startManualCompensator(c *gin.Context) {

	fmt.Println("Web API: Start Manual Compensator")
	startManualCompensatorLocal()
	c.Status(http.StatusOK)

}

func stopManualCompensator(c *gin.Context) {

	fmt.Println("Web API: Stop Manual Compensator")
	stopManualCompensatorLocal()
	c.Status(http.StatusOK)

}
