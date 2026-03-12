package server

import (
	"fmt"
	"sync"
)

type CommandHandler func([]RespValue) RespValue

var Handlers = map[string]CommandHandler{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

func ping(args []RespValue) RespValue {
	if len(args) == 0 {
		return RespValue{typ: RespTypeString, str: "PONG"}
	}

	return RespValue{typ: RespTypeString, str: args[0].bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(args []RespValue) RespValue {
	if len(args) != 2 {
		return RespValue{typ: RespTypeError, str: fmt.Sprintf("wrong number of arguments for 'set' command: %d", len(args))}
	}

	SETsMu.Lock()
	SETs[args[0].bulk] = args[1].bulk
	SETsMu.Unlock()

	return RespValue{typ: RespTypeString, str: "OK"}
}

func get(args []RespValue) RespValue {
	if len(args) != 1 {
		return RespValue{typ: RespTypeError, str: fmt.Sprintf("wrong number of arguments for 'get' command: %d", len(args))}
	}

	SETsMu.RLock()
	value, ok := SETs[args[0].bulk]
	SETsMu.RUnlock()

	if !ok {
		return RespValue{typ: RespTypeNull}
	}

	return RespValue{typ: RespTypeBulk, bulk: value}
}

type hash map[string]string

var HSETs = map[string]hash{}
var HSETsMu = sync.RWMutex{}

func hset(args []RespValue) RespValue {
	if len(args) != 3 {
		return RespValue{typ: RespTypeError, str: fmt.Sprintf("wrong number of arguments for 'hset' command: %d", len(args))}
	}

	HSETsMu.Lock()
	if _, ok := HSETs[args[0].bulk]; !ok {
		HSETs[args[0].bulk] = hash{}
	}
	HSETs[args[0].bulk][args[1].bulk] = args[2].bulk
	HSETsMu.Unlock()

	return RespValue{typ: RespTypeString, str: "OK"}
}

func hget(args []RespValue) RespValue {
	if len(args) != 2 {
		return RespValue{typ: RespTypeError, str: fmt.Sprintf("wrong number of arguments for 'hget' command: %d", len(args))}
	}

	SETsMu.RLock()
	hash, ok := HSETs[args[0].bulk]
	if !ok {
		SETsMu.RUnlock()
		return RespValue{typ: RespTypeNull}
	}

	value, ok := hash[args[1].bulk]
	SETsMu.RUnlock()

	if !ok {
		return RespValue{typ: RespTypeNull}
	}

	return RespValue{typ: RespTypeBulk, bulk: value}
}

func hgetall(args []RespValue) RespValue {
	if len(args) != 1 {
		return RespValue{typ: RespTypeError, str: fmt.Sprintf("wrong number of arguments for 'hgetall' command: %d", len(args))}
	}

	SETsMu.RLock()
	hash, ok := HSETs[args[0].bulk]
	SETsMu.RUnlock()

	if !ok {
		return RespValue{typ: RespTypeNull}
	}

	result := make([]RespValue, 0, len(hash)*2)
	for k, v := range hash {
		result = append(result, RespValue{typ: RespTypeBulk, bulk: k})
		result = append(result, RespValue{typ: RespTypeBulk, bulk: v})
	}

	return RespValue{typ: RespTypeArray, array: result}
}
