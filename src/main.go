package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	TypeHandleLine = 0
	TypeErrNum     = 1
)

var TypeMonitorChan = make(chan int, 200)

func (m *Monitor) start(lp *LogProcess) {

	go func() {
		for n := range TypeMonitorChan {
			switch n {
			case TypeErrNum:
				m.data.ErrNum += 1
			case TypeHandleLine:
				m.data.HandleLine += 1
			}
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			<-ticker.C
			m.tpsSlice = append(m.tpsSlice, m.data.HandleLine)
			if len(m.tpsSlice) > 2 {
				m.tpsSlice = m.tpsSlice[1:]
			}
		}
	}()

	http.HandleFunc("/monitor", func(writer http.ResponseWriter, request *http.Request) {
		m.data.RunTime = time.Now().Sub(m.startTime).String()
		m.data.ReadChanLen = len(lp.rc)
		m.data.WriteChanLen = len(lp.wc)

		if len(m.tpsSlice) >= 2 {
			m.data.TPS = float64(m.tpsSlice[1]-m.tpsSlice[0]) / 5
		}

		// 将结构体转换为字节数组
		result, _ := json.MarshalIndent(m.data, "", "\t")
		// 输出 JSON 格式
		io.WriteString(writer, string(result))
	})
	// 此方法是阻塞的
	http.ListenAndServe(":9193", nil)
}

func (r *ReadFromFile) Read(rc chan string) {
	// 读取文件
	f, err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("open file error: %s", err.Error()))
	}

	// 从文件末尾开始逐行读取文件内容
	f.Seek(0, 2)
	readBuffer := bufio.NewReader(f)

	/**
	实际生产环境中，文件可能按日或者周进行切割，需要处理这种情况
	*/

	for {
		line, err := readBuffer.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("ReadBytes error: %s", err.Error()))
		}
		TypeMonitorChan <- TypeHandleLine
		rc <- string(line[:len(line)-1])
	}

}

func (w *WtireToPrometheus) Write(wc chan *Message) {

}

func (w *WriteToInfluxDB) Write(wc chan *Message) {
	// 解析模块

	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     w.InfluxDB.Address,
		Username: w.InfluxDB.Username,
		Password: w.InfluxDB.Password,
	})
	if err != nil {
		logrus.Fatal(err)
	}
	defer c.Close()

	for v := range wc {
		fmt.Println(v)
		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  w.InfluxDB.Database,
			Precision: w.InfluxDB.Precision,
		})
		if err != nil {
			logrus.Fatal(err)
		}

		// Create a point and add to batch
		/**
		tags: Path, Method, Scheme, Status
		*/
		tags := map[string]string{
			"Path":   v.Path,
			"Method": v.Method,
			"Scheme": v.Scheme,
			"Status": v.Status,
		}
		/**
		fields: UpstreamTime, RequestTime, BytesSend
		*/
		fields := map[string]interface{}{
			"UpstreamTime": v.UpstreamTime,
			"RequestTime":  v.RequestTime,
			"BytesSend":    v.BytesSend,
		}

		pt, err := client.NewPoint("nginx_log", tags, fields, v.TimeLocal)
		if err != nil {
			logrus.Fatal(err)
		}
		bp.AddPoint(pt)

		// Write the batch
		if err := c.Write(bp); err != nil {
			logrus.Fatal(err)
		}
	}

}

func (lp *LogProcess) Process() {
	// 解析模块

	/**

	127.0.0.1 - - [04/Mar/2018:13:49:52 +0000] http "GET /foo?query=t HTTP/1.0" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854



	正则表达式：

	([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)
	*/
	expr := `([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`
	r := regexp.MustCompile(expr)
	// 获取时区
	location, _ := time.LoadLocation("Asia/Shanghai")

	for v := range lp.rc {
		result := r.FindStringSubmatch(v)
		if len(result) != 14 {
			TypeMonitorChan <- TypeErrNum
			logrus.Infoln("FindStringSubmatch fail: %s", v)
			continue
		}

		message := &Message{}
		t, err := time.ParseInLocation("02/Jan/2006:15:04:05 +0000", result[4], location)
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			logrus.Println("ParseInLocation fail:", err.Error(), result[4])
			continue
		}
		message.TimeLocal = t

		message.BytesSend, _ = strconv.Atoi(result[8])
		// GET /foo?query=t HTTP/1.0
		requestLine := strings.Split(result[6], " ")
		if len(requestLine) != 3 {
			TypeMonitorChan <- TypeErrNum
			logrus.Println("strings.Split fail", result[6])
			continue
		}

		message.Method = requestLine[0]

		u, err := url.Parse(requestLine[1])
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			logrus.Println("URL parse fail:", err)
			continue
		}

		message.Path = u.Path

		message.Scheme = result[5]

		message.Status = result[7]

		message.UpstreamTime, _ = strconv.ParseFloat(result[12], 64)
		message.RequestTime, _ = strconv.ParseFloat(result[13], 64)

		lp.wc <- message
	}

}

var conf Configuration

func init() {
	bytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logrus.Fatalf("Read file error: %s", err.Error())
	}
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		logrus.Fatalf("yaml Unmarshal error: %s", err.Error())
	}
}

func main() {

	var path string

	flag.StringVar(&path, "path", "src/access.log", "read file path")

	// 解析参数
	flag.Parse()

	reader := &ReadFromFile{
		path: path,
	}

	writer := &WriteToInfluxDB{
		InfluxDB: conf.InfluxDB,
	}

	lp := &LogProcess{
		rc:     make(chan string, 200),
		wc:     make(chan *Message, 200),
		reader: reader,
		writer: writer,
	}

	go lp.reader.Read(lp.rc)
	/**
	根据处理速度调整 go routine 的数量，可以使用参数设置
	*/
	for i := 0; i < 2; i++ {
		go lp.Process()
	}

	for i := 0; i < 4; i++ {
		go lp.writer.Write(lp.wc)
	}

	m := &Monitor{
		startTime: time.Now(),
		data:      SystemInfo{},
	}

	m.start(lp)
}
