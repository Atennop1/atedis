package server

type RespType byte

const (
	RespTypeString  RespType = '+'
	RespTypeInteger RespType = ':'
	RespTypeBulk    RespType = '$'
	RespTypeArray   RespType = '*'

	RespTypeError RespType = '-'
	RespTypeNull  RespType = '0' // some random character that will represent null
)

type RespValue struct {
	typ RespType

	str   string
	num   int
	bulk  string
	array []RespValue
}
