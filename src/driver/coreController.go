package driver

import "time"

//通道常量设置
const dataChDepth = int64(100)         //主数据通道队列深度（ADC->数据交换器）
const socketChDepth = int64(10)        //上位机发送通道队列深度（数据交换器->SocketWorker）
const helperChDepth = int64(100000)    //辅助数据通道队列深度（数据交换器->TemperatureWorker）
const socketDataLength = int64(500)    //上位机单次发送数据数量（上位机发送）
const socketDownSampleRate = int64(20) //上位机单次发送数据的降采样系数，即丢弃n-1个数据。n=1时采样率为50kHz
const heaterDownSampleRate = int64(2)  //加热控制器降采样系数，即丢弃n-1个数据。n=1时采样率为50kHz

//AXI地址映射
const ADCFilterAddress int64 = 0xA0000000     //经过617阶高精度滤波器的数据通路的地址
const ADCFastFilterAddress int64 = 0xA0020000 //经过20阶快速滤波器的数据通路的地址
const ADCRawAddress int64 = 0xA0010000        //未经过滤波器的数据通路的地址
const ADCSPIAddress int64 = 0xB0020000        //ADC的SPI控制器的地址
const DACSPIAddress int64 = 0xB0010000        //DAC的SPI控制器的地址

//DSC芯片信息
const DSCChipType string = "MIS DSC-1"         //DSC芯片型号
const DSCChipMaxTemperture float64 = 600       //DSC芯片最高温度,单位：摄氏度
const DSCChipMinTemperture float64 = 0         //DSC芯片最低温度,单位：摄氏度
const DSCChipAlpha float64 = 27.99268e6        //DSC芯片温度灵敏度,单位：nV/K
const DSCChipBeta float64 = -611.16518e6       //DSC芯片温度偏置,单位：nV
const DSCChipMaxHeatingVoltage float64 = 4.4e9 //DSC芯片最高加热电压,单位：nV

//数据通路常量设置
//vos: 电压偏置，单位：nV
type DAQInfo struct {
	vos [4]float64
}

//温度控制器实时信息变量，用于向上位机发送温度控制器实时信息
type HeaterInfo struct {
	voltage     [2]float64
	power       [2]float64
	temperature [2]float64
	target      [2]float64
}

type CoreControllerInterface interface {
	Initialize()
	ConnectSocket() bool
	DisconnectSocket()
	ConnectADC()
	StartFetchData()
	StopFetchData()
	AvgVoltage(int64) avgVoltageData
	SetADCVos(int64, float64)
	SetDACVoltage(string, float64) float64
	GetDACVoltage(string) float64
	SetDACOffset(string, float64)
	StartProgramHeater(int64, float64, float64, []ProgHeatingStep)
	StopProgramHeater()
	StartStaticHeater(float64, float64)
	StopStaticHeater()
	SetupStaticTemperature(float64)
	SetupHeaterPIDParameter(float64, float64, float64, float64, float64, float64)
	SetupCompensatorPIDParameter(float64, float64, float64, float64, float64, float64)
	SetupCompensator(float64)
	StartManualCompensator()
	StopManualCompensator()
}

//CoreController是整个程序的核心控制器，负责初始化所有控制器，管理所有控制器的工作状态，并提供对外接口
type CoreController struct {
	WORKFLAG   int32          //工作状态标志位
	datach     chan DAQDataCH //主数据通道
	socketch   chan socketCH  //上位机发送通道
	dchhelper1 chan DAQDataCH //辅助数据通道1
	dchhelper2 chan DAQDataCH //辅助数据通道2
	ctlch      chan string    //控制通道
	DAQSetting DAQInfo        //数据通路参数

	socket socketController      //上位机数据发送控制器
	adc    ADCController         //ADC控制器
	dac    DACController         //DAC控制器
	tmp    TemperatureController //温度控制器

	heater HeaterInfo //温度控制器实时信息变量
}

//在栈上初始化全局变量

//ADC数据采集标志位
var runningDAQ = bool(true)

//上位机发送通道标志位
var socketCHSign = uint8(0)

//辅助数据通道标志位
var helperCHSign = uint8(0)

func (corectl *CoreController) Initialize() {

	//初始化全局标志位

	runningDAQ = true
	socketCHSign = 0
	helperCHSign = 0

	//初始化温度控制器实时信息变量

	corectl.heater = HeaterInfo{[2]float64{0, 0}, [2]float64{0, 0}, [2]float64{0, 0}, [2]float64{0, 0}}

	//初始化信号通道

	corectl.datach = make(chan DAQDataCH, dataChDepth)
	corectl.socketch = make(chan socketCH, socketChDepth)
	corectl.dchhelper1 = make(chan DAQDataCH, helperChDepth)
	corectl.dchhelper2 = make(chan DAQDataCH, helperChDepth)
	corectl.ctlch = make(chan string, 10)

	//初始化上位机数据发送控制器（Socket网络设备）

	corectl.socket = *newSocketController()
	corectl.socket.setCH(corectl.socketch)

	//初始化ADC设置

	ADCvos := [4]float64{2300.9, 7023, 0, 0}
	corectl.DAQSetting = DAQInfo{vos: ADCvos}

	//初始化ADC控制器

	corectl.adc = *newADCController(ADCSPIAddress, ADCFastFilterAddress) /*选择是否经过数字滤波器*/
	corectl.adc.setVos(0, corectl.DAQSetting.vos[0])
	corectl.adc.setVos(1, corectl.DAQSetting.vos[1])

	//初始化DAC控制器

	corectl.dac = *newDACController(DACSPIAddress)
	corectl.dac.regDACPort("TP1", "HVDAC", 0)
	corectl.dac.regDACPort("TP2", "HVDAC", 1)

	//重置DAC电压为0

	corectl.dac.setDACVoltage("TP1", float64(0))
	time.Sleep(time.Duration(1) * time.Millisecond)
	corectl.dac.setDACVoltage("TP2", float64(0))

	//初始化温度控制器

	corectl.tmp = *newTemperatureController(corectl.dac, corectl.dchhelper2, &corectl.heater)

	//初始化数据交换器

	go interconnectHub(corectl.datach, &corectl.heater, corectl.socketch, corectl.dchhelper1, corectl.dchhelper2)
}

func (corectl *CoreController) ConnectSocket() bool {
	resultChan := make(chan bool, 1)
	go corectl.socket.start(resultChan)
	result := <-resultChan
	close(resultChan)
	return result
}

func (corectl *CoreController) DisconnectSocket() {
	corectl.socket.stop()
}

func (corectl *CoreController) ConnectADC() {
	corectl.adc.initialize()
}

func (corectl *CoreController) StartFetchData() {
	runningDAQ = true
	go corectl.fetchData(corectl.datach)
}

func (corectl *CoreController) StopFetchData() {
	runningDAQ = false
}
