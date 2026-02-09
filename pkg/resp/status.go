package resp

type ResultCode interface {
	error
	Code() int64
}
