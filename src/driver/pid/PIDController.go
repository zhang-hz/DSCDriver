package pid

import (
	"math"
)

type PIDController struct {
	Kp               float64
	Ki               float64
	Kd               float64
	Interval         float64
	tolerance        float64
	errorTolerance   float64
	tau              float64
	Target           float64
	InitialValue     float64
	errLinr          float64
	errDerv          float64
	errIntg          float64
	lastError        float64
	lastMeasurement  float64
	lastOutPut       float64
	bufferCompension float64
	limitMin         float64
	limitMax         float64
	limitMinIntg     float64
	limitMaxIntg     float64
}

type PIDSetting struct {
	Kp             float64
	Ki             float64
	Kd             float64
	Interval       float64
	Tau            float64
	Tolerance      float64
	ErrorTolerance float64
	InitialValue   float64
	LimitMin       float64
	LimitMax       float64
	LimitMinIntg   float64
	LimitMaxIntg   float64
}

func (pidctl *PIDController) initialize() {

	pidctl.Kp = 0
	pidctl.Ki = 0
	pidctl.Kd = 0
	pidctl.Interval = 20000
	pidctl.tau = 0
	pidctl.tolerance = 0
	pidctl.errorTolerance = 0
	pidctl.Target = 0
	pidctl.InitialValue = 0
	pidctl.errLinr = 0
	pidctl.errDerv = 0
	pidctl.errIntg = 0
	pidctl.lastError = 0
	pidctl.lastOutPut = 0
	pidctl.bufferCompension = 0
	pidctl.lastMeasurement = 0
	pidctl.limitMin = 0
	pidctl.limitMax = 0
	pidctl.limitMinIntg = 0
	pidctl.limitMaxIntg = 0

}

func (pidctl *PIDController) Linear(measurement float64) float64 {

	//fmt.Println(pidctl.Target)

	errorNow := pidctl.Target - measurement
	//噪声误差判断是否进入PID控制
	if errorNow < pidctl.errorTolerance && errorNow > -pidctl.errorTolerance {
		return pidctl.lastOutPut
	}

	compension := pidctl.Kp * errorNow

	pidctl.bufferCompension = pidctl.bufferCompension + compension
	output := pidctl.lastOutPut + pidctl.bufferCompension

	if output > pidctl.limitMax {
		output = pidctl.limitMax
	} else if output < pidctl.limitMin {
		output = pidctl.limitMin
	}

	if math.Abs(pidctl.bufferCompension) < pidctl.tolerance {
		return pidctl.lastOutPut
	} else {
		pidctl.lastOutPut = output
		pidctl.bufferCompension = 0
		return output
	}

}

func (pidctl *PIDController) BasicPID(measurement float64) float64 {

	errorNow := pidctl.Target - measurement

	if math.Abs(errorNow) < pidctl.errorTolerance {
		return pidctl.lastOutPut
	}

	pidctl.errLinr = pidctl.Kp * errorNow
	pidctl.errIntg = pidctl.errIntg + pidctl.Ki*errorNow
	pidctl.errDerv = errorNow - pidctl.lastError
	kdTemp := pidctl.Kd

	intgDerv := pidctl.errIntg - pidctl.bufferCompension
	pidctl.bufferCompension = pidctl.errIntg

	if intgDerv > 0 && pidctl.errDerv < 0 || intgDerv < 0 && pidctl.errDerv > 0 {
		if math.Abs(pidctl.errDerv) > 6e5 {
			kdTemp = pidctl.Kd * (math.Abs(pidctl.errDerv)*3.33e-4 - 2e2 + 1)
			if kdTemp > 10 {
				kdTemp = 10
			}
		}
	}

	pidctl.errDerv = kdTemp * pidctl.errDerv

	pidctl.lastError = errorNow

	output := (pidctl.errLinr + pidctl.errDerv + pidctl.errIntg) //

	Vh_deltaT := output - pidctl.lastOutPut

	//pidctl.bufferCompension = pidctl.bufferCompension + compensation

	if output > pidctl.limitMax {
		output = pidctl.limitMax
	} else if output < pidctl.limitMin {
		output = pidctl.limitMin
	}

	if math.Abs(Vh_deltaT) < pidctl.tolerance {
		return pidctl.lastOutPut
	} else {
		pidctl.lastOutPut = output
		return output
	}

}

func (pidctl *PIDController) Reset() {

	pidctl.Target = 0
	pidctl.lastOutPut = pidctl.InitialValue
	pidctl.errLinr = 0
	pidctl.errDerv = 0
	pidctl.errIntg = 0
	pidctl.lastError = 0
	pidctl.lastOutPut = 0
	pidctl.lastMeasurement = 0

}

func (pidctl *PIDController) Setup(setting PIDSetting) {

	pidctl.Kp = setting.Kp
	pidctl.Ki = setting.Ki
	pidctl.Kd = setting.Kd
	pidctl.Interval = setting.Interval
	pidctl.tau = setting.Tau
	pidctl.tolerance = setting.Tolerance
	pidctl.errorTolerance = setting.ErrorTolerance
	pidctl.InitialValue = setting.InitialValue
	pidctl.limitMin = setting.LimitMin
	pidctl.limitMax = setting.LimitMax
	pidctl.limitMinIntg = setting.LimitMinIntg
	pidctl.limitMaxIntg = setting.LimitMaxIntg

}

func NewPIDController() *PIDController {

	var PIDCTL = &PIDController{}
	PIDCTL.initialize()

	return PIDCTL
}
