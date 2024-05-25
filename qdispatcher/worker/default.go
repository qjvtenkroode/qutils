package worker

type Message struct {
	/*
	   Message is used to pass a payload down to a particular worker.
	   The topic is used to register workers on and route the message.
	*/
	Topic   string
	Payload string
}

type Worker interface {
	AddMessage(Message)
	Start()
	Stop()
}


