package driver

import (
	"fmt"
	"time"
)

//恒温控制器
//输入：
//DAQDataCH，输入数据通道（接收ADC数据），定义见coreWorker.go
//HeaterInfo，输出数据通道（发送恒温控制器实时信息），定义见coreController.go
//输出：无
func (tmpctl *TemperatureControllerInstance) heating(din <-chan DAQDataCH, heaterInfo *HeaterInfo) {

	//初始化恒温控制器

	//清空输入通道，防止接收到过时数据
	for len(din) > 0 {
		<-din
	}
	//设置helperCHSign，开启输入通道
	helperCHSign = helperCHSign | 0x2

	//再次清空输入通道，防止接收到过时数据
	for i := int64(0); i < dataChDepth+100; i++ {
		<-din
	}

	//初始化数据接收变量
	var dtmp DAQDataCH

	//sleepTime := time.Duration(tmpctl.heater.Interval/2) * time.Nanosecond
	//设置DAC输出间隔
	dacInterval := time.Duration(5) * time.Microsecond

	//初始化缓存变量
	lastRefVoltage := float64(0)
	lastSenseVoltage := float64(0)
	lastCompensationVoltage := float64(0)

	//恒温控制器主循环，启动恒温控制器
	for {

		//判断helperCHSign，如果helperCHSign&0x2==0，说明恒温控制器已经被关闭，退出循环
		//通过数据通道判断恒温控制器开关可以避免在关闭恒温控制器时，由于数据通道未清空导致的阻塞
		if helperCHSign&0x2 == 0 {
			helperCHSign = helperCHSign & 0xFD
			heaterInfo.voltage[0] = 0
			heaterInfo.voltage[1] = 0
			fmt.Print("Core API: Stop heater temperature controller \n")
			return
		}

		//从数据通道接收数据
		dtmp = <-din

		//使用PID控制器计算共模加热电压
		//输入：参考热堆的输出电压，单位：nV。参考热堆的输出电压需通过敏感热堆的输出电压与差分电压计算得到。
		//输出：共模加热电压，单位：nV
		common := tmpctl.heater.BasicPID(dtmp.directv[0] - dtmp.directv[1])

		//使用PID控制器计算补偿电压
		//输入：差分输出电压，单位：nV
		//输出：补偿电压，单位：nV
		//如果补偿未开启，则补偿电压为0
		compensation := float64(0)
		if tmpctl.isCompensating {
			compensation = tmpctl.compensator.BasicPID(dtmp.directv[1])
		}

		//计算参考热堆的加热电压
		refVoltage := common
		//计算敏感热堆的加热电压
		senseVoltage := common - compensation

		//fmt.Println(heater)

		//如果共模加热电压发生变化，则通过DAC的TP2通道输出更新的的加热电压
		//同时更新输出数据通道的参考热堆加热电压数据
		if refVoltage != lastRefVoltage {
			tmpctl.dac.setDACVoltage("TP2", refVoltage)
			time.Sleep(dacInterval)
			heaterInfo.voltage[1] = refVoltage
			lastRefVoltage = refVoltage
		}
		//如果补偿电压发生变化，则通过DAC的TP1通道输出更新的的补偿电压
		if compensation != lastCompensationVoltage {
			tmpctl.dac.setDACVoltage("TP1", compensation)
			time.Sleep(dacInterval)
			lastCompensationVoltage = compensation
		}
		//如果敏感热堆的加热电压发生变化，则更新输出数据通道的敏感热堆加热电压数据
		if senseVoltage != lastSenseVoltage {
			heaterInfo.voltage[0] = senseVoltage
			lastSenseVoltage = senseVoltage

		}
		//time.Sleep(sleepTime)

		//循环

	}

}

//程序控温控制器
//输入：
//DAQDataCH，输入数据通道（接收ADC数据），定义见coreWorker.go
//HeaterInfo，输出数据通道（发送恒温控制器实时信息），定义见coreController.go
//输出：无
func (tmpctl *TemperatureControllerInstance) progHeating(din <-chan DAQDataCH, heaterInfo *HeaterInfo) {

	//初始化恒温控制器

	//清空输入通道，防止接收到过时数据
	for len(din) > 0 {
		<-din
	}
	//设置helperCHSign，开启输入通道
	helperCHSign = helperCHSign | 0x2

	//再次清空输入通道，防止接收到过时数据
	for i := int64(0); i < dataChDepth+100; i++ {
		<-din
	}

	//初始化数据接收变量
	var dtmp DAQDataCH

	//sleepTime := time.Duration(tmpctl.heater.Interval/4) * time.Nanosecond
	//设置DAC输出间隔
	dacInterval := time.Duration(5) * time.Microsecond

	//初始化缓存变量
	lastRefVoltage := float64(0)
	lastSenseVoltage := float64(0)
	lastCompensationVoltage := float64(0)

	//程序控温控制器主循环，启动程序控制器
	for {

		//判断helperCHSign，如果helperCHSign&0x2==0，说明程序控温控制器已经被关闭，退出循环
		//通过数据通道判断程序控温控制器开关可以避免在关闭程序控温控制器时，由于数据通道未清空导致的阻塞
		if helperCHSign&0x2 == 0 {
			helperCHSign = helperCHSign & 0xFD
			heaterInfo.voltage[0] = 0
			heaterInfo.voltage[1] = 0
			fmt.Print("Core API: Stop heater temperature controller \n")
			return
		}

		//从数据通道接收数据
		dtmp = <-din

		//根据程序控温曲线计算当前时刻的目标电压
		//输入：无
		//输出：目标电压，单位：nV
		tmpctl.heater.Target = tmpctl.progVTMap()

		//使用加热PID控制器计算共模加热电压
		//输入：参考热堆的输出电压，单位：nV。参考热堆的输出电压需通过敏感热堆的输出电压与差分电压计算得到。
		//输出：共模加热电压，单位：nV
		common := tmpctl.heater.BasicPID(dtmp.directv[0] - dtmp.directv[1])

		//使用补偿PID控制器计算补偿电压
		//如果补偿未开启，则补偿电压为0
		//输入：差分输出电压，单位：nV
		//输出：补偿电压，单位：nV
		compensation := float64(0)
		if tmpctl.isCompensating {
			compensation = tmpctl.compensator.BasicPID(dtmp.directv[1])
		}

		//计算参考热堆的加热电压
		refVoltage := common
		//计算敏感热堆的加热电压
		senseVoltage := common - compensation

		//更新输出数据通道的目标电压数据
		heaterInfo.target[0] = tmpctl.heater.Target

		//如果参考热堆的加热电压发生变化，则通过DAC的TP2通道输出更新的的参考热堆加热电压
		if refVoltage != lastRefVoltage {
			tmpctl.dac.setDACVoltage("TP2", refVoltage)
			time.Sleep(dacInterval)
			//更新输出数据通道的参考热堆加热电压数据
			heaterInfo.voltage[1] = refVoltage
			lastRefVoltage = refVoltage
		}
		//如果补偿电压发生变化，则通过DAC的TP1通道输出更新的的补偿电压
		if compensation != lastCompensationVoltage {
			tmpctl.dac.setDACVoltage("TP1", compensation)
			time.Sleep(dacInterval)
			lastCompensationVoltage = compensation
		}
		//如果敏感热堆的加热电压发生变化，则通过DAC的TP0通道输出更新的的敏感热堆加热电压
		if senseVoltage != lastSenseVoltage {
			heaterInfo.voltage[0] = senseVoltage
			lastSenseVoltage = senseVoltage

		}
		//time.Sleep(sleepTime)

	}

}

func (tmpctl *TemperatureControllerInstance) manualCompensator(din <-chan DAQDataCH, heaterInfo *HeaterInfo) {

	for len(din) > 0 {
		<-din
	}
	helperCHSign = helperCHSign | 0x2

	for i := int64(0); i < dataChDepth+100; i++ {
		<-din
	}

	var dtmp DAQDataCH

	//sleepTime := time.Duration(tmpctl.heater.Interval/2) * time.Nanosecond
	dacInterval := time.Duration(5) * time.Microsecond

	lastCompensation := float64(0)

	for {

		if helperCHSign&0x2 == 0 {
			helperCHSign = helperCHSign & 0xFD
			heaterInfo.voltage[0] = heaterInfo.voltage[1]
			fmt.Print("Core API: Stop heater temperature controller \n")
			return
		}

		dtmp = <-din

		tmpctl.compensator.Target = 0
		compensation := tmpctl.compensator.BasicPID(dtmp.directv[1])

		//fmt.Println(heater)

		if compensation != lastCompensation {
			tmpctl.dac.setDACVoltage("TP1", compensation)
			time.Sleep(dacInterval)
			heaterInfo.voltage[0] = heaterInfo.voltage[1] - compensation
			lastCompensation = compensation
		}

		//time.Sleep(sleepTime)

	}

}
