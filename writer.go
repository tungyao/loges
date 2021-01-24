package loges

type LogesWriter interface {
	Write(p []interface{}) (n int, err error)
}
