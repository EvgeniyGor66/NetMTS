package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type SMSData struct {
	Country      string `json:"country"`
	Bandwidth    string `json:"bandwidth"`
	ResponseTime string `json:"response_time"`
	Provider     string `json:"provider"`
}

type MMSData struct {
	Country      string `json:"country"`
	Provider     string `json:"provider"`
	Bandwidth    string `json:"bandwidth"`
	ResponseTime string `json:"response_time"`
}

type VoiceCallData struct {
	Country             string  `json:"country"`
	Bandwidth           string  `json:"bandwidth"`
	ResponseTime        string  `json:"response_time"`
	Provider            string  `json:"provider"`
	ConnectionStability float32 `json:"connection_stability"`
	TTFB                int     `json:"ttfb"`
	VoicePurity         int     `json:"voice_purity"`
	MedianOfCallsTime   int     `json:"median_of_calls_time"`
}

type EmailData struct {
	Country      string `json:"country"`
	Provider     string `json:"provider"`
	DeliveryTime int    `json:"delivery_time"`
}

type BillingData struct {
	CreateCustomer bool `json:"create_customer"`
	Purchase       bool `json:"purchase"`
	Payout         bool `json:"payout"`
	Recurring      bool `json:"recurring"`
	FraudControl   bool `json:"fraud_control"`
	CheckoutPage   bool `json:"checkout_page"`
}

type SupportData struct {
	Topic         string `json:"topic"`
	ActiveTickets int    `json:"active_tickets"`
}

type IncidentData struct {
	Topic  string `json:"topic"`
	Status string `json:"status"` // возможные статусы active и closed
}
type ResultT struct {
	Status bool       `json:"status"` // true, если все этапы сбора данных прошли успешно, false во всех остальных случаях
	Data   ResultSetT `json:"data"`   // заполнен, если все этапы сбора данных прошли успешно, nil во всех остальных случаях
	Error  string     `json:"error"`  // пустая строка если все этапы сбора данных прошли успешно, в случае ошибки заполнено текстом ошибки
}

type ResultSetT struct {
	SMS       [][]SMSData              `json:"sms"`
	MMS       [][]MMSData              `json:"mms"`
	VoiceCall []VoiceCallData          `json:"voice_call"`
	Email     map[string][][]EmailData `json:"email"`
	Billing   BillingData              `json:"billing"`
	Support   []int                    `json:"support"`
	Incidents []IncidentData           `json:"incident"`
}

const (
	prov1 string = "Topolo"
	prov2 string = "Rond"
	prov3 string = "Kildy"
	prov4 string = "TransparentCalls"
	prov5 string = "E-Voice"
	prov6 string = "JustPhone"
)

var emailProv = [...]string{
	"Gmail",
	"Yahoo",
	"Hotmail",
	"MSN",
	"Orange",
	"Comcast",
	"AOL",
	"Live",
	"RediffMail",
	"GMX",
	"Protonmail",
	"Yandex",
	"Mail.ru",
}

func GetDataSms(c SMSData) ([]SMSData, error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var finalSmsData []SMSData
	var stringsData []string

	fContent, err := os.ReadFile("web/sms.data")
	if err != nil {
		errorLog.Println(err)
		return finalSmsData, err
	}

	stringsData = strings.Split(string(fContent), "\n")

	for _, stringData := range stringsData { //разбиваем построчно
		str := strings.Split(stringData, ";")
		if len(str) != 4 {
			continue
		}
		correctData := validation(str)
		c.Country = correctData[0]
		c.Bandwidth = correctData[1]
		c.ResponseTime = correctData[2]
		c.Provider = correctData[3]
		finalSmsData = append(finalSmsData, c)
	}

	return finalSmsData, nil
}

func GetMms() ([]MMSData, error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	correctMms := []MMSData{}
	Mms := []MMSData{}
	r, err := http.Get("http://localhost:8383/mms")
	if err != nil {
		log.Fatal(err)
		return Mms, err
	}

	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		errorLog.Fatal(err)
		return Mms, err
	}
	r.Body.Close()
	err = json.Unmarshal(reqData, &Mms)
	if err != nil {
		errorLog.Fatal(err)
		return Mms, err
	}

	for _, field := range Mms {
		correctData := MMSData{}
		countryCode := field.Country
		provider := field.Provider
		bandwidth, _ := strconv.Atoi(field.Bandwidth)
		time, _ := strconv.Atoi(field.ResponseTime)
		_, ok := checkCountry(countryCode)
		if !ok {
			break
		} else {
			correctData.Country = countryCode
		}
		if provider != prov1 && provider != prov2 && provider != prov3 {
			break
		} else {
			correctData.Provider = provider
		}
		if bandwidth > 100 || bandwidth < 0 {
			break
		} else {
			correctData.Bandwidth = field.Bandwidth
		}
		if time > 0 {
			correctData.ResponseTime = field.ResponseTime
		} else {
			break
		}
		correctMms = append(correctMms, correctData)
	}

	return correctMms, nil
}

func GetDataVoiceCall(c VoiceCallData) ([]VoiceCallData, error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var finalVoiceCallData []VoiceCallData
	var stringsData []string

	fContent, err := os.ReadFile("web/voice.data")
	if err != nil {
		errorLog.Println(err)
		return finalVoiceCallData, err
	}

	stringsData = strings.Split(string(fContent), "\n")

	for _, stringData := range stringsData { //разбиваем построчно
		str := strings.Split(stringData, ";")
		if len(str) != 8 {
			continue
		}
		correctData := validation(str)
		c.Country = correctData[0]
		c.Bandwidth = correctData[1]
		c.ResponseTime = correctData[2]
		c.Provider = correctData[3]
		fTof, _ := strconv.ParseFloat(correctData[4], 32)
		c.ConnectionStability = float32(fTof)
		c.TTFB, _ = strconv.Atoi(correctData[5])
		c.VoicePurity, _ = strconv.Atoi(correctData[6])
		c.MedianOfCallsTime, _ = strconv.Atoi(correctData[7])
		finalVoiceCallData = append(finalVoiceCallData, c)
	}

	return finalVoiceCallData, nil
}

func GetDataEmail(c EmailData) ([]EmailData, error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var finalEmailData []EmailData
	var stringsData []string

	fContent, err := os.ReadFile("web/email.data")
	if err != nil {
		errorLog.Println(err)
		return finalEmailData, err
	}

	stringsData = strings.Split(string(fContent), "\n")

	for _, stringData := range stringsData { //разбиваем построчно
		var correctData = make([]string, 10)
		str := strings.Split(stringData, ";")
		if len(str) != 3 {
			continue
		} else {
			for index, field := range str { //разбиваем строки по полям
				switch index {
				case 0: //проверка первого поля "код страны"
					{
						countryCode := string(field)
						_, ok := checkCountry(countryCode)
						if !ok {
							break
						} else {
							correctData[0] = field
						}
					}
				case 1:
					{
						provider := string(field)
						for _, prov := range emailProv {
							if prov != provider {
								continue
							} else {
								correctData[1] = field
							}
						}
					}
				case 2:
					{
						deliverytime, _ := strconv.Atoi(field)
						if deliverytime > 0 {
							correctData[2] = field
						} else {
							break
						}
					}

				}

			}
			c.Country = correctData[0]
			c.Provider = correctData[1]
			c.DeliveryTime, _ = strconv.Atoi(correctData[2])
			finalEmailData = append(finalEmailData, c)
		}
	}
	return finalEmailData, nil
}

func GetDataBilling() (BillingData, error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	var Billing BillingData
	reversBilling := []byte{0, 0, 0, 0, 0, 0}
	var billingBool []bool

	fContent, err := os.ReadFile("web/billing.data")
	if err != nil {
		errorLog.Println(err)
		return Billing, err
	}

	for i, b := range fContent {
		reversBilling[(len(fContent)-1)-i] = b
	}

	for _, i := range reversBilling {

		if i == '1' {
			billingBool = append(billingBool, false)
		} else {
			billingBool = append(billingBool, true)
		}
	}
	Billing.CreateCustomer = billingBool[0]
	Billing.Purchase = billingBool[1]
	Billing.Payout = billingBool[2]
	Billing.Recurring = billingBool[3]
	Billing.FraudControl = billingBool[4]
	Billing.CheckoutPage = billingBool[5]
	return Billing, nil
}

func GetSupport(c SupportData) ([]SupportData, error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	Support := []SupportData{}
	r, err := http.Get("http://localhost:8383/support")
	if err != nil {
		log.Fatal(err)
		return Support, err
	}

	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		errorLog.Fatal(err)
		return Support, err
	}
	r.Body.Close()
	err = json.Unmarshal(reqData, &Support)
	if err != nil {
		errorLog.Fatal(err)
		return Support, err
	}

	return Support, nil
}

func GetIncident(c IncidentData) ([]IncidentData, error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	Incident := []IncidentData{}
	r, err := http.Get("http://localhost:8383/accendent")
	if err != nil {
		log.Fatal(err)
		return Incident, err
	}

	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		errorLog.Fatal(err)
		return Incident, err
	}
	r.Body.Close()
	err = json.Unmarshal(reqData, &Incident)
	if err != nil {
		errorLog.Fatal(err)
		return Incident, err
	}

	return Incident, nil
}

func validation(s []string) []string {
	var correctData = make([]string, 8)
	for index, field := range s { //разбиваем строки по полям
		switch index {
		case 0: //проверка первого поля "код страны"
			{
				countryCode := string(field)
				_, ok := checkCountry(countryCode)
				if !ok {
					break
				} else {
					correctData[0] = field
				}
			}
		case 1:
			{
				bandwidth, _ := strconv.Atoi(field)
				if bandwidth > 100 || bandwidth < 0 {
					break
				} else {
					correctData[1] = field
				}
			}
		case 2:
			{
				time, _ := strconv.Atoi(field)
				if time > 0 {
					correctData[2] = field
				} else {
					break
				}
			}

		case 3:
			{
				provider := string(field)
				if provider != prov1 && provider != prov2 && provider != prov3 && provider != prov4 && provider != prov5 && provider != prov6 {
					break
				} else {
					correctData[3] = field
				}
			}
		case 4:
			{
				ConnectionStability, _ := strconv.ParseFloat(field, 32)
				if ConnectionStability > 0 {
					correctData[4] = field
				} else {
					break
				}
			}
		case 5:
			{
				TTFB, _ := strconv.Atoi(field)
				if TTFB > 0 {
					correctData[5] = field
				} else {
					break
				}
			}
		case 6:
			{
				VoicePurity, _ := strconv.Atoi(field)
				if VoicePurity > 0 {
					correctData[6] = field
				} else {
					break
				}
			}
		case 7:
			{
				MedianOfCallsTime, _ := strconv.Atoi(field)
				if MedianOfCallsTime > 0 {
					correctData[7] = field
				} else {
					break
				}
			}
		}

	}
	return correctData
}

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	var result ResultT
	var smsData SMSData
	var VoiceCallData VoiceCallData
	var EmailData EmailData
	var SupportData SupportData
	var IncidentData IncidentData
	result.Status = true
	resultSms, err := GetDataSms(smsData)
	if err != nil {
		result.Status = false
		result.Error = "Error on collect data"
	}
	fmt.Println("resultSms", resultSms)

	resultMms, err := GetMms()
	if err != nil {
		result.Status = false
		result.Error = "Error on collect data"
	}
	fmt.Printf("resultMms: %s \n", resultMms)

	resultVoiceCall, err := GetDataVoiceCall(VoiceCallData)
	if err != nil {
		result.Status = false
		result.Error = "Error on collect data"
	}
	fmt.Println("resultVoiceCall", resultVoiceCall)

	resultEmail, err := GetDataEmail(EmailData)
	if err != nil {
		result.Status = false
		result.Error = "Error on collect data"
	}
	fmt.Println("resultEmail", resultEmail)

	resultBilling, err := GetDataBilling()
	if err != nil {
		result.Status = false
		result.Error = "Error on collect data"
	}
	fmt.Println("resultBilling", resultBilling)

	resultSupport, err := GetSupport(SupportData)
	if err != nil {
		result.Status = false
		result.Error = "Error on collect data"
	}
	fmt.Println("resultSupport", resultSupport)

	resultIncident, err := GetIncident(IncidentData)
	if err != nil {
		result.Status = false
		result.Error = "Error on collect data"
	}
	fmt.Println("resultIncident", resultIncident)
	resultData := GetResultData(resultSms, resultMms, resultVoiceCall, resultEmail, resultBilling, resultSupport, resultIncident)
	fmt.Println("resultData", resultData)
	result.Data = resultData
	dResultT, err := json.Marshal(result)
	if err != nil {
		errorLog.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(dResultT)
}

func Result() ([]byte, []byte) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	var ResultT ResultT
	var ResultSetT ResultSetT
	dResultT, err := json.Marshal(ResultT)
	if err != nil {
		errorLog.Fatal(err)
	}
	dResultSetT, err := json.Marshal(ResultSetT)
	if err != nil {
		errorLog.Fatal(err)
	}

	return dResultT, dResultSetT
}

func GetResultData(sms []SMSData, mms []MMSData, VoiceCall []VoiceCallData, resultEmail []EmailData, resultBilling BillingData, resultSupport []SupportData, Incidents []IncidentData) ResultSetT {
	var (
		resultSetT   ResultSetT
		finalSmsData []SMSData
		smsSort      [][]SMSData
		mmsSort      [][]MMSData
		finalMmsData []MMSData
	)
	for _, smsD := range sms {
		country, _ := checkCountry(smsD.Country)
		smsD.Country = country
		finalSmsData = append(finalSmsData, smsD)
	}
	sort.Slice(finalSmsData, func(i, j int) bool {
		return finalSmsData[i].Country < finalSmsData[j].Country
	})
	var finalSmsDataByCountry []SMSData
	finalSmsDataByCountry = append(finalSmsDataByCountry, finalSmsData...)
	sort.Slice(finalSmsData, func(i, j int) bool {
		return finalSmsData[i].Provider < finalSmsData[j].Provider
	})
	finalSmsDataByProvider := finalSmsData
	smsSort = [][]SMSData{finalSmsDataByProvider, finalSmsDataByCountry}
	resultSetT.SMS = smsSort

	for _, mmsD := range mms {
		country, _ := checkCountry(mmsD.Country)
		mmsD.Country = country
		finalMmsData = append(finalMmsData, mmsD)
	}
	sort.Slice(finalMmsData, func(i, j int) bool {
		return finalMmsData[i].Country < finalMmsData[j].Country
	})
	var finalMmsDataByCountry []MMSData
	finalMmsDataByCountry = append(finalMmsDataByCountry, finalMmsData...)
	sort.Slice(finalMmsData, func(i, j int) bool {
		return finalMmsData[i].Provider < finalMmsData[j].Provider
	})
	finalMmsDataByProvider := finalMmsData
	mmsSort = [][]MMSData{finalMmsDataByProvider, finalMmsDataByCountry}
	resultSetT.MMS = mmsSort

	resultSetT.VoiceCall = VoiceCall

	email := make(map[string][][]EmailData)
	emailSortByCountry := make(map[string][]EmailData)
	emailDataSortDt := resultEmail
	sort.SliceStable(emailDataSortDt, func(i, j int) bool { return emailDataSortDt[i].DeliveryTime < emailDataSortDt[j].DeliveryTime })

	for _, val := range emailDataSortDt {
		emailSortByCountry[val.Country] = append(emailSortByCountry[val.Country], val)
	}

	for i, val := range emailSortByCountry {
		var emailDtFast []EmailData
		var emailDtSlow []EmailData
		for i, x := range val {
			if i < 3 {
				emailDtFast = append(emailDtFast, x)
			}
			if i > len(val)-4 {
				emailDtSlow = append(emailDtSlow, x)
			}
		}
		email[i] = append(email[i], emailDtFast, emailDtSlow)
	}
	resultSetT.Email = email
	resultSetT.Billing = resultBilling

	//
	supportTickets := 0
	for _, supD := range resultSupport {
		supportTickets += supD.ActiveTickets
	}
	supportLoad := 1
	switch supportTickets {
	case 0, 1, 2, 3, 4, 5, 6, 7, 8:
		supportLoad = 1
	case 9, 10, 11, 12, 13, 14, 15, 16:
		supportLoad = 2
	default:
		supportLoad = 3
	}

	supportTime := supportTickets * 200 //время ожидания ответа в секундах
	resultSetT.Support = append(resultSetT.Support, supportLoad, supportTime/60)

	sort.SliceStable(Incidents, func(i, j int) bool { return Incidents[i].Status < Incidents[j].Status })

	resultSetT.Incidents = Incidents

	return resultSetT
}

func checkCountry(c string) (string, bool) {
	Country := map[string]string{
		"AU": "Австралия",
		"AT": "Австрия",
		"AZ": "Азербайджан",
		"AX": "Аландские острова",
		"AL": "Албания",
		"DZ": "Алжир",
		"VI": "Виргинские Острова (США)",
		"AS": "Американское Самоа",
		"AI": "Ангилья",
		"AO": "Ангола",
		"AD": "Андорра",
		"AQ": "Антарктика",
		"AG": "Антигуа и Барбуда",
		"AR": "Аргентина",
		"AM": "Армения",
		"AW": "Аруба",
		"AF": "Афганистан",
		"BS": "Багамские Острова",
		"BD": "Бангладеш",
		"BB": "Барбадос",
		"BH": "Бахрейн",
		"BZ": "Белиз",
		"BY": "Белоруссия",
		"BE": "Бельгия",
		"BJ": "Бенин",
		"BM": "Бермуды",
		"BG": "Болгария",
		"BO": "Боливия",
		"BQ": "Бонайре, Синт-Эстатиус и Саба",
		"BA": "Босния и Герцеговина",
		"BW": "Ботсвана",
		"BR": "Бразилия",
		"IO": "Британская территория в Индийском океане",
		"VG": "Виргинские Острова (Великобритания)",
		"BN": "Бруней",
		"BF": "Буркина-Фасо",
		"BI": "Бурунди",
		"BT": "Бутан",
		"VU": "Вануату",
		"VA": "Ватикан",
		"GB": "Великобритания",
		"HU": "Венгрия",
		"VE": "Венесуэла",
		"UM": "Внешние малые острова США",
		"TL": "Восточный Тимор",
		"VN": "Вьетнам",
		"GA": "Габон",
		"HT": "Гаити",
		"GY": "Гайана",
		"GM": "Гамбия",
		"GH": "Гана",
		"GP": "Гваделупа",
		"GT": "Гватемала",
		"GF": "Гвиана",
		"GN": "Гвинея",
		"GW": "Гвинея-Бисау",
		"DE": "Германия",
		"GG": "Гернси",
		"GI": "Гибралтар",
		"HN": "Гондурас",
		"HK": "Гонконг",
		"GD": "Гренада",
		"GL": "Гренландия",
		"GR": "Греция",
		"GE": "Грузия",
		"GU": "Гуам",
		"DK": "Дания",
		"JE": "Джерси",
		"DJ": "Джибути",
		"DM": "Доминика",
		"DO": "Доминиканская Республика",
		"CD": "ДР Конго",
		"EU": "Европейский союз",
		"EG": "Египет",
		"ZM": "Замбия",
		"EH": "САДР",
		"ZW": "Зимбабве",
		"IL": "Израиль",
		"IN": "Индия",
		"ID": "Индонезия",
		"JO": "Иордания",
		"IQ": "Ирак",
		"IR": "Иран",
		"IE": "Ирландия",
		"IS": "Исландия",
		"ES": "Испания",
		"IT": "Италия",
		"YE": "Йемен",
		"CV": "Кабо-Верде",
		"KZ": "Казахстан",
		"KY": "Острова Кайман",
		"KH": "Камбоджа",
		"CM": "Камерун",
		"CA": "Канада",
		"QA": "Катар",
		"KE": "Кения",
		"CY": "Кипр",
		"KG": "Киргизия",
		"KI": "Кирибати",
		"TW": "Китайская Республика",
		"KP": "КНДР (Корейская Народно-Демократическая Республика)",
		"CN": "Китай (Китайская Народная Республика)",
		"CC": "Кокосовые острова",
		"CO": "Колумбия",
		"KM": "Коморы",
		"CR": "Коста-Рика",
		"CI": "Кот-д’Ивуар",
		"CU": "Куба",
		"KW": "Кувейт",
		"CW": "Кюрасао",
		"LA": "Лаос",
		"LV": "Латвия",
		"LS": "Лесото",
		"LR": "Либерия",
		"LB": "Ливан",
		"LY": "Ливия",
		"LT": "Литва",
		"LI": "Лихтенштейн",
		"LU": "Люксембург",
		"MU": "Маврикий",
		"MR": "Мавритания",
		"MG": "Мадагаскар",
		"YT": "Майотта",
		"MO": "Макао",
		"MK": "Северная Македония",
		"MW": "Малави",
		"MY": "Малайзия",
		"ML": "Мали",
		"MV": "Мальдивы",
		"MT": "Мальта",
		"MA": "Марокко",
		"MQ": "Мартиника",
		"MH": "Маршалловы Острова",
		"MX": "Мексика",
		"FM": "Микронезия",
		"MZ": "Мозамбик",
		"MD": "Молдавия",
		"MC": "Монако",
		"MN": "Монголия",
		"MS": "Монтсеррат",
		"MM": "Мьянма",
		"NA": "Намибия",
		"NR": "Науру",
		"NP": "Непал",
		"NE": "Нигер",
		"NG": "Нигерия",
		"NL": "Нидерланды",
		"NI": "Никарагуа",
		"NU": "Ниуэ",
		"NZ": "Новая Зеландия",
		"NC": "Новая Каледония",
		"NO": "Норвегия",
		"AE": "ОАЭ",
		"OM": "Оман",
		"BV": "Остров Буве",
		"IM": "Остров Мэн",
		"CK": "Острова Кука",
		"NF": "Остров Норфолк",
		"CX": "Остров Рождества",
		"PN": "Острова Питкэрн",
		"SH": "Остров Святой Елены",
		"PK": "Пакистан",
		"PW": "Палау",
		"PS": "Государство Палестина",
		"PA": "Панама",
		"PG": "Папуа — Новая Гвинея",
		"PY": "Парагвай",
		"PE": "Перу",
		"PL": "Польша",
		"PT": "Португалия",
		"PR": "Пуэрто-Рико",
		"CG": "Республика Конго",
		"KR": "Республика Корея",
		"RE": "Реюньон",
		"RU": "Россия",
		"RW": "Руанда",
		"RO": "Румыния",
		"SV": "Сальвадор",
		"WS": "Самоа",
		"SM": "Сан-Марино",
		"ST": "Сан-Томе и Принсипи",
		"SA": "Саудовская Аравия",
		"SZ": "Эсватини",
		"MP": "Северные Марианские Острова",
		"SC": "Сейшельские Острова",
		"BL": "Сен-Бартелеми",
		"MF": "Сен-Мартен",
		"PM": "Сен-Пьер и Микелон",
		"SN": "Сенегал",
		"VC": "Сент-Винсент и Гренадины",
		"KN": "Сент-Китс и Невис",
		"LC": "Сент-Люсия",
		"RS": "Сербия",
		"SG": "Сингапур",
		"SX": "Синт-Мартен",
		"SY": "Сирия",
		"SK": "Словакия",
		"SI": "Словения",
		"SB": "Соломоновы Острова",
		"SO": "Сомали",
		"SD": "Судан",
		"SR": "Суринам",
		"US": "США",
		"SL": "Сьерра-Леоне",
		"TJ": "Таджикистан",
		"TH": "Таиланд",
		"TZ": "Танзания",
		"TC": "Теркс и Кайкос",
		"TG": "Того",
		"TK": "Токелау",
		"TO": "Тонга",
		"TT": "Тринидад и Тобаго",
		"TV": "Тувалу",
		"TN": "Тунис",
		"TM": "Туркменистан",
		"TR": "Турция",
		"UG": "Уганда",
		"UZ": "Узбекистан",
		"UA": "Украина",
		"WF": "Уоллис и Футуна",
		"UY": "Уругвай",
		"FO": "Фарерские острова",
		"FJ": "Фиджи",
		"PH": "Филиппины",
		"FI": "Финляндия",
		"FK": "Фолклендские острова",
		"FR": "Франция",
		"PF": "Французская Полинезия",
		"TF": "Французские Южные и Антарктические территории",
		"HM": "Херд и Макдональд",
		"HR": "Хорватия",
		"CF": "ЦАР",
		"TD": "Чад",
		"ME": "Черногория",
		"CZ": "Чехия",
		"CL": "Чили",
		"CH": "Швейцария",
		"SE": "Швеция",
		"SJ": "Шпицберген и Ян-Майен",
		"LK": "Шри-Ланка",
		"EC": "Эквадор",
		"GQ": "Экваториальная Гвинея",
		"ER": "Эритрея",
		"EE": "Эстония",
		"ET": "Эфиопия",
		"ZA": "ЮАР",
		"GS": "Южная Георгия и Южные Сандвичевы Острова",
		"SS": "Южный Судан",
		"JM": "Ямайка",
		"JP": "Япония",
	}
	country, ok := Country[c]
	return country, ok
}
