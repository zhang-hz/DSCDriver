package driver

import (
	pidctl "DAQDriver/src/driver/pid"
	"fmt"
	"time"
)

//温度控制器接口
type TemperatureController interface {
	initialize()
	setupProgHeatingCurve(int64, float64, []ProgHeatingStep)
	startProgHeater(int64, float64, float64, []ProgHeatingStep)
	stopProgHeater()
	startStaticHeater(float64, float64)
	stopStaticHeater()
	setupStaticTemperature(float64)
	setupHeaterPID(pidctl.PIDSetting)
	setupCompensatorPID(pidctl.PIDSetting)
	setupCompensator(float64)
	startManualCompensator()
	stopManualCompensator()
}

//温度控制器对象
type TemperatureControllerInstance struct {
	vtRelation       []vtCoefficient      //温度-电压映射关系(数组切片)
	datain           chan DAQDataCH       //数据输入通道（接收ADC数据）
	dac              DACController        //DAC控制器对象
	heater           pidctl.PIDController //程序升温PID控制器对象
	compensator      pidctl.PIDController //温度补偿PID控制器对象
	isCompensating   bool                 //是否正在进行温度补偿
	info             *HeaterInfo          //加热器工作实时信息
	baseVoltage      float64              //没有加热时的输出电压，用于补偿电压偏置
	progHeatingInfo  progHeatingSetting   //程序控温的状态参数
	progHeatingCurve []progHeatingProcess //程序控温温度曲线
}

//温度-电压映射函数的单段数据结构
type vtCoefficient struct {
	start float64 //DSC芯片温度范围的起始值，单位：摄氏度
	end   float64 //DSC芯片温度范围的终止值，单位：摄氏度
	a     float64 //DSC芯片的温度灵敏度，单位：nV/K
	b     float64 //DSC芯片的零点修正，单位：nV
}

//程序控温控制器状态参数
type progHeatingSetting struct {
	startTime       int64   //程序控温的起始时间，单位：us，UTC时间
	stepNumber      int64   //当前控温过程的序号
	baseTemperature float64 //程序控温的起始温度，即环境温度
	endTemperature  float64 //程序控温的终止温度，即最后一步控温的终止温度
	stepLen         int64   //程序控温的总步数
}

//程序控温的与上位机通讯的单步骤数据结构
type ProgHeatingStep struct {
	ProcessType      string  //单步控温过程类型：heat,hold,cool
	StartTemperature float64 //单步控温过程起始温度，单位：摄氏度
	StopTemperature  float64 //单步控温过程终止温度，单位：摄氏度
	ProcessSpeed     float64 //单步控温过程升温速率，单位：摄氏度/s
	ProcessTime      float64 //单步控温过程的持续时间，单位：s
}

//程序升温曲线的单步骤数据结构
type progHeatingProcess struct {
	processType      int64   //单步控温过程类型：1-升温，0-保温，-1-降温
	startTemperature float64 //单步控温过程起始温度，单位：摄氏度
	stopTemperature  float64 //单步控温过程终止温度，单位：摄氏度
	startTime        int64   //单步控温过程开始时间，单位：us，相对于程序控温的起始时间
	processSpeed     float64 //单步控温过程升温速率，单位：摄氏度/us
	processTime      int64   //单步控温过程的持续时间，单位：us，相对于本步骤控温的起始时间
}

//初始化温度控制器对象
func (tmpctl *TemperatureControllerInstance) initialize() {

	//初始化温度-电压映射关系
	//tmpctl.vtRelation为一个切片，用于存储温度-电压映射函数的多段线性函数
	tmpctl.vtRelation = make([]vtCoefficient, 1)
	//针对Zhang DSC芯片的温度灵敏度和零点修正参数，需要根据实际情况进行修改
	//Zhang DSC在工作温度范围内保持线性，因此仅需要一段线性函数即可
	//如果其他DSC芯片的温度灵敏度和零点修正参数不是线性的，则需要添加更多的vtCoefficient结构体，使用多段线性函数进行拟合
	tmpctl.vtRelation[0] = vtCoefficient{
		start: DSCChipMinTemperture,
		end:   DSCChipMaxTemperture,
		a:     DSCChipAlpha,
		b:     DSCChipBeta,
	}
	//初始化温度控制器的状态参数
	tmpctl.baseVoltage = 0
	tmpctl.isCompensating = false
	tmpctl.progHeatingInfo = progHeatingSetting{startTime: 0, stepNumber: 0, baseTemperature: 0, endTemperature: 0, stepLen: 0}
	tmpctl.progHeatingCurve = []progHeatingProcess{}

}

//热堆温度-输出电压映射函数
//输入：temerature，含义：温度，单位：摄氏度
//输出：v，含义：电压，单位：nV
func (tmpctl *TemperatureControllerInstance) vtmap(temperature float64) float64 {

	//初始化电压值
	var v = float64(0)

	//根据温度-电压关系(tmpctl.vtRelation)计算电压值
	//温度-电压关系(tmpctl.vtRelation)为至少一段线性函数。可以根据实际情况添加更多的分段，以提高精度
	//需要注意的是，需要保持温度-电压关系(tmpctl.vtRelation)的温度范围和输出值连续，否则可能导致温度-电压映射函数不连续
	//每一段线性函数的公式为：y = a*temperature + b，其中a为当前分段的温度灵敏度，b为截距
	//单位：电压：nV，温度灵敏度：nV/K，温度：K
	for i := 0; i < len(tmpctl.vtRelation); i++ {
		//便利温度-电压关系(tmpctl.vtRelation)，找到温度所在的分段
		if temperature > tmpctl.vtRelation[i].start && temperature < tmpctl.vtRelation[i].end {
			//计算电压值
			v = tmpctl.vtRelation[i].a*temperature + tmpctl.vtRelation[i].b
			break
		}
	}

	return v

}

//将升温过程转换为升温曲线，并完成升温曲线的设置
//输入：
//count：升温过程数。超过count个升温过程的参数将被忽略
//baseTemperature：升温曲线的起始温度，即环境温度，单位：摄氏度
//step：升温过程数组切片
//输出：
//无
func (tmpctl *TemperatureControllerInstance) setupProgHeatingCurve(count int64, baseTemperature float64, step []ProgHeatingStep) {

	//初始化升温曲线
	tmpctl.progHeatingCurve = []progHeatingProcess{}

	//检查参数是否合法
	if count <= 0 {
		return
	}
	if len(step) == 0 {
		return
	} else if len(step) < int(count) {
		count = int64(len(step))
	}
	//检查初始温度是否合法
	if step[0].StartTemperature != baseTemperature {
		step[0].StartTemperature = baseTemperature
	}
	//初始化时间中间变量，温度中间变量
	timeTemp := int64(0)
	lastTemperature := step[0].StartTemperature

	//生成升温曲线
	for i := 0; i < int(count); i++ {

		fmt.Println(i)
		//检查单步起始温度是否合法，需与上一步终止温度相同
		if step[i].StartTemperature != lastTemperature {
			step[i].StartTemperature = lastTemperature
		}

		//根据不同的控温方式计算升温曲线
		//升温过程类型：1-升温，0-保温，-1-降温
		if step[i].ProcessType == "heat" {

			//检查单步终止温度是否合法，需大于起始温度
			if step[i].StopTemperature <= step[i].StartTemperature {
				continue
			}
			//检查单步升温速率是否合法，需大于0
			if step[i].ProcessSpeed <= float64(0) {
				continue
			}
			//向升温曲线中添加升温过程
			//温度单位：摄氏度，时间单位：us，速率单位：摄氏度/us（正值）
			tmpctl.progHeatingCurve = append(tmpctl.progHeatingCurve, progHeatingProcess{
				processType:      1,
				startTime:        timeTemp,
				startTemperature: step[i].StartTemperature,
				stopTemperature:  step[i].StopTemperature,
				processSpeed:     step[i].ProcessSpeed,
				processTime:      int64((step[i].StopTemperature - step[i].StartTemperature) / step[i].ProcessSpeed),
			})
			//更新时间中间变量
			lastTemperature = step[i].StopTemperature
			timeTemp += int64((step[i].StopTemperature - step[i].StartTemperature) / step[i].ProcessSpeed)

		} else if step[i].ProcessType == "cool" {

			//检查单步终止温度是否合法，需小于起始温度
			if step[i].StopTemperature >= step[i].StartTemperature {
				continue
			}
			//检查单步降温速率是否合法，需取绝对值
			if step[i].ProcessSpeed < float64(0) {
				step[i].ProcessSpeed = -1 * step[i].ProcessSpeed
			} else if step[i].ProcessSpeed == float64(0) {
				continue
			}

			//向升温曲线中添加降温过程
			//温度单位：摄氏度，时间单位：us，速率单位：摄氏度/us（负值）
			tmpctl.progHeatingCurve = append(tmpctl.progHeatingCurve, progHeatingProcess{
				processType:      -1,
				startTime:        timeTemp,
				startTemperature: step[i].StartTemperature,
				stopTemperature:  step[i].StopTemperature,
				processSpeed:     -1 * step[i].ProcessSpeed,
				processTime:      int64((step[i].StartTemperature - step[i].StopTemperature) / step[i].ProcessSpeed),
			})
			//更新时间中间变量
			lastTemperature = step[i].StopTemperature
			timeTemp += int64((step[i].StartTemperature - step[i].StopTemperature) / step[i].ProcessSpeed)

		} else if step[i].ProcessType == "hold" {

			//检查单步终止温度是否合法，需与起始温度相同
			if step[i].StartTemperature != step[i].StopTemperature {
				continue
			}
			//检查单步保温时间是否合法，需大于0
			if step[i].ProcessTime == float64(0) {
				continue
			}

			//向升温曲线中添加恒温过程
			//温度单位：摄氏度，时间单位：us，速率单位：摄氏度/us（0）
			tmpctl.progHeatingCurve = append(tmpctl.progHeatingCurve, progHeatingProcess{
				processType:      0,
				startTime:        timeTemp,
				startTemperature: step[i].StartTemperature,
				stopTemperature:  step[i].StartTemperature,
				processSpeed:     0,
				processTime:      int64(step[i].ProcessTime),
			})

			//更新时间中间变量
			lastTemperature = step[i].StopTemperature
			timeTemp += int64(step[i].ProcessTime)
		}
	}

	//更新升温曲线信息
	tmpctl.progHeatingInfo.baseTemperature = baseTemperature
	tmpctl.progHeatingInfo.endTemperature = tmpctl.progHeatingCurve[len(tmpctl.progHeatingCurve)-1].stopTemperature
	tmpctl.progHeatingInfo.stepLen = int64(len(tmpctl.progHeatingCurve))

}

//progVTMap 根据升温曲线计算当前时间点的目标温度
//输入：无  函数自动获取当前时间，无需输入
//输出：目标电压，单位：nV
func (tmpctl *TemperatureControllerInstance) progVTMap() float64 {

	//计算当前时间，单位：us
	timeNow := time.Now().UnixMicro() - tmpctl.progHeatingInfo.startTime
	//初始化目标温度
	progTemp := float64(0)
	//获取当前升温步骤的序号
	stepNumber := tmpctl.progHeatingInfo.stepNumber
	//获取升温曲线的总步数
	stepLen := tmpctl.progHeatingInfo.stepLen
	//fmt.Println(stepNumber)

	//计算目标温度
	if stepNumber < stepLen-1 {
		//如果当前升温步骤不是最后一步，则计算当前步骤的目标温度
		if timeNow > tmpctl.progHeatingCurve[stepNumber+1].startTime {
			//如果当前时间已经超过了下一步升温的开始时间，则进入下一步升温
			tmpctl.progHeatingInfo.stepNumber++
			stepNumber = tmpctl.progHeatingInfo.stepNumber
		}
		//根据本步骤的初始温度，本步骤的开始时间，本步骤的控温速率计算当前时间的目标温度
		//公式：目标温度 = 当前步骤初始温度 + (当前时间-当前步骤开始时间)*当前步骤控温速率，其中控温速率为正值，负值，或0
		//温度单位：摄氏度，时间单位：us，速率单位：K/us
		progTemp = tmpctl.progHeatingCurve[stepNumber].startTemperature + float64((timeNow-tmpctl.progHeatingCurve[stepNumber].startTime))*tmpctl.progHeatingCurve[stepNumber].processSpeed

	} else if stepNumber == stepLen-1 {
		//如果当前升温步骤是最后一步，则计算当前步骤的目标温度
		if timeNow > tmpctl.progHeatingCurve[stepNumber].startTime+tmpctl.progHeatingCurve[stepNumber].processTime {
			//如果当前时间已经超过了最后一步升温的结束时间，则完成程序控温，进入恒温状态
			tmpctl.progHeatingInfo.stepNumber++
			stepNumber = tmpctl.progHeatingInfo.stepNumber
			//恒温温度为程序升温曲线的终止温度
			progTemp = tmpctl.progHeatingInfo.endTemperature

		} else {
			//如果当前时间还未超过最后一步升温的结束时间，则计算当前时间的目标温度
			//温度单位：摄氏度，时间单位：us，速率单位：摄氏度/us
			progTemp = tmpctl.progHeatingCurve[stepNumber].startTemperature + float64(timeNow-tmpctl.progHeatingCurve[stepNumber].startTime)*tmpctl.progHeatingCurve[stepNumber].processSpeed
		}
	} else {

		//如果当前升温步骤已经结束，则恒温温度为程序升温曲线的终止温度
		progTemp = tmpctl.progHeatingInfo.endTemperature
	}

	//完成当前时刻的目标温度计算
	//使用vtmap函数计算的目标温度对应的电压
	//返回目标电压，单位：nV
	return tmpctl.vtmap(progTemp)

}

//启动程序控温
//输入：count：程序控温步骤总数，basevoltage：基准电压nV，baseTemperature：基准温度K，step：升温曲线过程数组
//输出：无
func (tmpctl *TemperatureControllerInstance) startProgHeater(count int64, basevoltage float64, baseTemperature float64, step []ProgHeatingStep) {

	//重置DAC，两个通道电压电压设置为0
	fmt.Println("Core API: Starting program heating temperature controller: ")
	tmpctl.dac.setDACVoltage("TP2", 0)
	tmpctl.info.voltage[1] = 0
	time.Sleep(time.Duration(10) * time.Microsecond)
	tmpctl.dac.setDACVoltage("TP1", 0)
	tmpctl.info.voltage[0] = 0
	time.Sleep(time.Duration(10) * time.Microsecond)

	//重置温度控制器参数
	tmpctl.heater.Reset()
	tmpctl.compensator.Reset()
	tmpctl.compensator.Target = 0

	//设置本次升温信息
	tmpctl.baseVoltage = basevoltage
	tmpctl.setupProgHeatingCurve(count, baseTemperature, step)
	tmpctl.progHeatingInfo.startTime = time.Now().UnixMicro()
	tmpctl.progHeatingInfo.stepNumber = 0

	//启动程序控温控制器
	go tmpctl.progHeating(tmpctl.datain, tmpctl.info)
	fmt.Println("Core API: Started program heating temperature controller: ")

}

//停止程序控温
func (tmpctl *TemperatureControllerInstance) stopProgHeater() {

	fmt.Println("Core API: Stopping program heating temperature controller: ")
	//切断数据通道
	//程序控温控制器在数据通道中断时会自动停止，这样可以保证程序控温控制器和数据传输通道的同步停止
	helperCHSign = helperCHSign & 0xFD
	time.Sleep(time.Duration(10) * time.Microsecond)

	//重置DAC
	tmpctl.dac.setDACVoltage("TP1", 0)
	time.Sleep(time.Duration(10) * time.Microsecond)
	tmpctl.dac.setDACVoltage("TP2", 0)
	tmpctl.info.voltage[1] = 0
	tmpctl.info.voltage[0] = 0

	tmpctl.info.target[0] = 0 //目标值reference重置为0

	fmt.Println("Core API: Stopped program heating temperature controller: ")

}

//静态恒温控制
//输入：basevoltage：基准电压nV，targetTemperature：目标温度K
//输出：无
func (tmpctl *TemperatureControllerInstance) startStaticHeater(basevoltage float64, targetTemperature float64) {

	fmt.Println("Core API: Starting heater temperature controller: ")
	//重置DAC
	tmpctl.dac.setDACVoltage("TP2", 0)
	tmpctl.info.voltage[1] = 0
	time.Sleep(time.Duration(1) * time.Millisecond)
	tmpctl.dac.setDACVoltage("TP1", 0)
	tmpctl.info.voltage[0] = 0
	time.Sleep(time.Duration(1) * time.Millisecond)
	//重置控制器
	tmpctl.heater.Reset()
	tmpctl.compensator.Reset()
	tmpctl.compensator.Target = 0

	tmpctl.baseVoltage = basevoltage
	tmpctl.heater.Target = tmpctl.vtmap(targetTemperature) + tmpctl.baseVoltage

	fmt.Println("Core API: V-T mapping: ", targetTemperature, "->", tmpctl.heater.Target)
	fmt.Println("Core API: Base Voltage: ", tmpctl.baseVoltage)
	fmt.Println("Core API: Heating Target Voltage: ", tmpctl.heater.Target)
	fmt.Println("Core API: Compensator Target Voltage: ", tmpctl.compensator.Target)

	//启动恒温控制器
	go tmpctl.heating(tmpctl.datain, tmpctl.info)
	fmt.Println("Core API: Started heater temperature controller: ")
}

//停止静态恒温控制
func (tmpctl *TemperatureControllerInstance) stopStaticHeater() {

	//切断数据通道
	//恒温控制器在数据通道中断时会自动停止，这样可以保证恒温控制器和数据传输通道的同步停止
	helperCHSign = helperCHSign & 0xFD
	//重置DAC
	tmpctl.dac.setDACVoltage("TP2", 0)
	tmpctl.info.voltage[1] = 0
	time.Sleep(time.Duration(1) * time.Millisecond)
	tmpctl.dac.setDACVoltage("TP1", 0)
	tmpctl.info.voltage[0] = 0
	time.Sleep(time.Duration(1) * time.Millisecond)

}

//改变静态恒温温度
func (tmpctl *TemperatureControllerInstance) setupStaticTemperature(temperature float64) {

	fmt.Println("Core API: Setup temperature: ", temperature)
	tmpctl.heater.Target = tmpctl.vtmap(temperature) + tmpctl.baseVoltage
	fmt.Println("Core API: Base Voltage: ", tmpctl.baseVoltage)
	fmt.Println("Core API: Target Voltage: ", tmpctl.heater.Target)

}

//PID参数设置

func (tmpctl *TemperatureControllerInstance) setupHeaterPID(pidsetting pidctl.PIDSetting) {

	tmpctl.heater.Setup(pidsetting)

}

func (tmpctl *TemperatureControllerInstance) setupCompensatorPID(pidsetting pidctl.PIDSetting) {

	tmpctl.compensator.Setup(pidsetting)

}

// 程序升温/恒温中的补偿设置

func (tmpctl *TemperatureControllerInstance) setupCompensator(isStart float64) {
	if isStart == 1 {
		tmpctl.isCompensating = true
	} else {
		tmpctl.isCompensating = false
	}
}

//在开环控温中的手动补偿
//不建议使用手动补偿，因为功率补偿是温度控制器自动控制的，手动开启补偿会产生逻辑冲突
func (tmpctl *TemperatureControllerInstance) startManualCompensator() {

	if !tmpctl.isCompensating {
		return
	}

	fmt.Println("Core API: Starting pure compesator: ")
	//重置DAC
	tmpctl.dac.setDACVoltage("TP1", 0)
	tmpctl.info.voltage[0] = 0
	time.Sleep(time.Duration(1) * time.Millisecond)
	//重置控制器
	tmpctl.compensator.Reset()
	tmpctl.compensator.Target = 0

	//手动启动补偿控制器
	go tmpctl.manualCompensator(tmpctl.datain, tmpctl.info)
	fmt.Println("Core API: Started pure compesator: ")

}

//停止在开环控温中的手动补偿
//仅用于停止手动补偿，不会影响到自动补偿
func (tmpctl *TemperatureControllerInstance) stopManualCompensator() {

	//切断数据通道
	//补偿控制器在数据通道中断时会自动停止，这样可以保证补偿控制器和数据传输通道的同步停止
	helperCHSign = helperCHSign & 0xFD
	//重置DAC
	tmpctl.dac.setDACVoltage("TP1", 0)
	tmpctl.info.voltage[0] = 0
	time.Sleep(time.Duration(1) * time.Millisecond)

}

//实例化温度控制器
func newTemperatureController(dacctl DACController, dchinput chan DAQDataCH, info *HeaterInfo) *TemperatureController {

	var HEATPID pidctl.PIDController = *pidctl.NewPIDController()
	var COMPPID pidctl.PIDController = *pidctl.NewPIDController()
	var TMPCTL TemperatureController = &TemperatureControllerInstance{dac: dacctl, heater: HEATPID, compensator: COMPPID, datain: dchinput, info: info}

	TMPCTL.initialize()

	return &TMPCTL

}
