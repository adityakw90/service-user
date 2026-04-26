package signal

type SignalType string

const (
	SignalStart   SignalType = "start"
	SignalReject  SignalType = "reject"
	SignalFail    SignalType = "fail"
	SignalSuccess SignalType = "success"
	SignalError   SignalType = "error"
)
