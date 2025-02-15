package sshserver

const (
	StatusContinue = "continue"
	StatusStop     = "stop"
)

const (
	StatusReady int32 = iota
	StatusWaitStart
	StatusRunning
	StatusWaitStop
	StatusStopping
	StatusFinished
)
