package resp

type StatusCode interface {
	error
	Code() int64
	Message() string
	IsAffectStability() bool
	Extra() map[string]string
}
