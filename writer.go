package loges

type DataStruct struct {
	Status   string `json:"status"`
	DateTime string `json:"date_time"`
	Pc       int64  `json:"pc"`
	Line     int    `json:"line"`
	File     string `json:"file"`
	Func     string `json:"func"`
	Msg      string `json:"msg"`
}
type LogesWriter interface {
	Write(dataStruct *DataStruct) (n int, err error)
}
