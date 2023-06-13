package web

import (
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {

	r := gin.New()

	apiv1 := r.Group("/api/v1")
	{
		apiv1.GET("/setfpga/:bitfilename", setPLBitFile)
		apiv1.GET("/connect", connectPC)
		apiv1.GET("/connectsocket", connectSocket)
		apiv1.GET("/disconnectsocket", disconnectSocket)
		apiv1.GET("/avgvoltage/:samplenumber", avgVoltage)
		apiv1.GET("/setadcvos/:adcch/:vosnumber", setADCVos)
		apiv1.GET("/setdacvoltage/:dacport/:voltage", setDACVoltage)
		apiv1.GET("/getdacvoltage/:dacport", getDACVoltage)
		apiv1.GET("/setdacoffset/:dacport/:offset", setDACOffset)
		apiv1.GET("/heater/static/start/:temperature/:basevoltage", startHeaterStaticPID)
		apiv1.GET("/heater/static/stop", stopHeaterStaticPID)
		apiv1.GET("/heater/static/temperature/:temperature", setupTemperature)
		apiv1.GET("/heater/pid/common/parameters/:kp/:ki/:kd/:tolerance/:errorTolerance/:initialValue", setupHeaterPIDParameter)
		apiv1.GET("/heater/pid/compensator/parameters/:kp/:ki/:kd/:tolerance/:errorTolerance/:initialValue", setupCompensatorPIDParameter)
		apiv1.GET("/heater/prog/start/:basevoltage/:heatingspeed/:coolspeed/:maxtemperature/:basetemperature", startHeaterProgramPID)
		apiv1.POST("/heater/prog/adv/start/:count/:basevoltage/:basetemperature", startHeaterProgramAdvPID)
		apiv1.GET("/heater/prog/stop", stopHeaterProgramPID)
		apiv1.GET("/heater/compensator/auto/start", startAutoCompensator)
		apiv1.GET("/heater/compensator/auto/stop", stopAutoCompensator)
		apiv1.GET("/heater/compensator/manual/start", startManualCompensator)
		apiv1.GET("/heater/compensator/manual/stop", stopManualCompensator)

	}

	return r

}
