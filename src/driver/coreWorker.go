package driver

import (
	"fmt"
	"time"
)

type DAQDataCH struct {
	directv [4]float64
}

//ADC数据采集，ADC->fetchData()->DAQDataCH
func (corectl *CoreController) fetchData(dout chan<- DAQDataCH) {

	fmt.Print("Start Fetch Data\n")

	var tmp [4]float64

	for {
		if !runningDAQ {
			fmt.Print("Stop Fetch Data\n")
			return
		}

		tmp = corectl.adc.read()
		dout <- DAQDataCH{directv: [4]float64{-1 * tmp[0], -1 * tmp[1], tmp[2], tmp[3]}}

	}
}

//非阻塞数据交换总线，
//DAQDataCH -> HeaterCH, socketCH, helperCH
//HeaterCH -> HeaterInfo -> socketCH, helperCH
func interconnectHub(din <-chan DAQDataCH, heater *HeaterInfo, socket chan<- socketCH, dout1 chan<- DAQDataCH, dout2 chan<- DAQDataCH) {

	var socketdata socketCH
	var num = int64(0)
	var downsample = int64(0)
	var heaterDownSample = int64(0)
	var data = DAQDataCH{}

	for {

		//接收ADC数据，DAQDataCH->din
		data = <-din

		//向上位机发送数据通道写入数据
		//din -> socketCH
		//首先进行降采样。降采样率为socketDownSampleRate，在coreController.go中定义
		if downsample >= socketDownSampleRate-1 {

			downsample = 0

			//判断socketCH是否开启或socketCH是否已经满了，如果关闭或满了，就不再写入数据
			if socketCHSign == 1 || (socketCHSign == 0 && (int64(len(socket)) < socketChDepth-1)) {

				//判断数据量是否达到发送要求，如果达到要求，就向上位机控制器发送数据
				if num >= socketDataLength {
					socket <- socketdata
					num = 0
					socketdata.time = time.UnixMicro(time.Now().UnixMicro())
					socketdata.interval = (1e9 * float64(socketDownSampleRate)) / 50e3
					socketdata.length = socketDataLength
				}

				socketdata.directv[0][num] = data.directv[0]                   //敏感热堆电压
				socketdata.directv[1][num] = data.directv[0] - data.directv[1] //参考热堆电压
				socketdata.directv[2][num] = data.directv[2]                   //敏感热堆加热器的采样电阻的电压
				socketdata.directv[3][num] = data.directv[3]                   //参考热堆加热器的采样电阻的电压
				socketdata.diffv[num] = data.directv[1]                        //差分热堆电压

				socketdata.heaterv[0][num] = heater.voltage[0] - data.directv[2] //敏感热堆加热器的加热电压
				socketdata.heaterv[1][num] = heater.voltage[1] - data.directv[3] //参考热堆加热器的加热电压
				socketdata.heaterv[2][num] = heater.target[0]                    //程序控温的目标电压

				socketdata.heaterp[0][num] = 0.85 * (heater.voltage[0] - data.directv[2]*1.4) * (data.directv[2] / 10) / 1e9 //敏感热堆加热器的加热功率
				socketdata.heaterp[1][num] = 0.85 * (heater.voltage[1] - data.directv[3]*1.4) * (data.directv[3] / 10) / 1e9 //参考热堆加热器的加热功率
				socketdata.heaterp[2][num] = socketdata.heaterp[0][num] - socketdata.heaterp[1][num]                         //波换补偿功率

				num++
			}
		} else {
			downsample++
		}

		if helperCHSign&0x1 != 0 && int64(len(dout1)) < helperChDepth-1 {
			dout1 <- data
		}

		if (helperCHSign&0x2 != 0 && int64(len(dout2)) < helperChDepth-1) || (helperCHSign&0x2 == 0 && (int64(len(dout2)) < helperChDepth-1)) {
			if heaterDownSample > heaterDownSampleRate-1 {
				dout2 <- data
				heaterDownSample = 0
			} else {
				heaterDownSample++
			}

		}

	}

}
