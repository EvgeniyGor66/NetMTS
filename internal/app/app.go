package app

import (
	"fmt"
	"log"
	"net/http"
	"netmts/internal/controller"

	"github.com/gorilla/mux"
)

func Run() {
	var smsData controller.SMSData
	var VoiceCallData controller.VoiceCallData
	var EmailData controller.EmailData
	var SupportData controller.SupportData
	var IncidentData controller.IncidentData

	resultSms := controller.GetDataSms(smsData)
	fmt.Println("resultSms", resultSms)

	resultMms := controller.GetMms()
	fmt.Printf("resultMms: %s \n", resultMms)

	resultVoiceCall := controller.GetDataVoiceCall(VoiceCallData)
	fmt.Println("resultVoiceCall", resultVoiceCall)

	resultEmail := controller.GetDataEmail(EmailData)
	fmt.Println("resultEmail", resultEmail)

	resultBilling := controller.GetDataBilling()
	fmt.Println("resultBilling", resultBilling)

	resultSupport := controller.GetSupport(SupportData)
	fmt.Println("resultSupport", resultSupport)

	resultIncident := controller.GetIncident(IncidentData)
	fmt.Println("resultIncident", resultIncident)

	//controller.GetDataBilling()
	resultT, resultSetT := controller.Result()
	fmt.Println("Result:", resultT, "ResultSetT:", resultSetT)

	resultSmsSort := controller.GetResultData(resultSms, resultMms, resultVoiceCall, resultEmail)
	fmt.Println("resultSmsSort", resultSmsSort)

	router := mux.NewRouter()
	router.HandleFunc("/", controller.HandleConnection)
	server := &http.Server{
		Addr:    ":8282",
		Handler: router,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
