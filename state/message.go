package state

type SinkMessage interface {
	String() (string, bool)
}
