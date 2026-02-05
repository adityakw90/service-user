package signal

type ServiceSignal string

const (
	SignalStart   ServiceSignal = "start"
	SignalReject  ServiceSignal = "reject"
	SignalFail    ServiceSignal = "fail"
	SignalSuccess ServiceSignal = "success"
)
