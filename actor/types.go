package actor

type DeadLetter struct {
	Target  *PID
	Message any
}

type Termination struct {
	PID *PID
}

type InternalError struct {
	From string
	Err  error
}

type Initialized struct{}

type Started struct{}

type Stopped struct{}

type WithSender struct {
	Message any
	Sender  *PID
}
