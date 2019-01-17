package main

import "time"

type Reader interface {
	Read(rc chan string)
}

type Writer interface {
	Write(wc chan *Message)
}

type ReadFromFile struct {
	path string // 待读取的文件路径
}

type WriteToInfluxDB struct {
	InfluxDB InfluxDB
}

type WtireToPrometheus struct {
}

type LogProcess struct {
	rc     chan string
	wc     chan *Message
	reader Reader
	writer Writer
}

type Message struct {
	TimeLocal    time.Time
	BytesSend    int
	Path         string
	Method       string
	Scheme       string
	Status       string
	UpstreamTime float64
	RequestTime  float64
}

// 系统状态监控
type SystemInfo struct {
	HandleLine   int     `json:"handleLine"`   // 总处理日志行数
	TPS          float64 `json:"tps"`          // 系统吞吐量
	ReadChanLen  int     `json:"readChanLen"`  // read channel 长度
	WriteChanLen int     `json:"writeChanLen"` // Write channel 长度
	RunTime      string  `json:"runTime"`      // 运行总时间
	ErrNum       int     `json:"errNum"`       // 错误数
}

type Monitor struct {
	startTime time.Time
	data      SystemInfo
	tpsSlice  []int
}

type Configuration struct {
	InfluxDB InfluxDB `yaml: "influxdb"`
}

type InfluxDB struct {
	Address   string `yaml:"address"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Database  string `yaml:"database"`
	Precision string `yaml:"precision"`
}
