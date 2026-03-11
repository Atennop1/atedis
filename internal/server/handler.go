package server

type CommandHandler func([]RespValue) RespValue

var Handlers = map[string]CommandHandler{
	"PING": ping,
}

func ping(args []RespValue) RespValue {
	if len(args) == 0 {
		return RespValue{typ: RespTypeString, str: "PONG"}
	}

	return RespValue{typ: RespTypeString, str: args[0].bulk}
}
