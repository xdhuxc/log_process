package main

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type Reader interface {
	Read(rc chan string)
}

type Writer interface {
	Write(wc chan Message)
}

type ReadFromFile struct {
	path string // 待读取的文件路径
}

type WriteToInfluxDB struct {
	influxDBDns string
}

type LogProcess struct {
	rc     chan string
	wc     chan Message
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

func (r *ReadFromFile) Read(rc chan string) {
	// 读取文件
	f, err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("open file error: %s", err.Error()))
	}

	// 从文件末尾开始逐行读取文件内容
	f.Seek(0, 2)
	readBuffer := bufio.NewReader(f)

	for {
		line, err := readBuffer.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("ReadBytes error: %s", err.Error()))
		}
		rc <- string(line[:len(line)-1])
	}

}

func (w *WriteToInfluxDB) Write(wc chan Message) {
	// 解析模块

	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MyDB,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a point and add to batch
	tags := map[string]string{"cpu": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}

	pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}

	for v := range wc {
		fmt.Println(v)
	}

}

func (lp *LogProcess) Process() {
	// 解析模块

	/**

	127.0.0.1 - - [04/Mar/2018:13:49:52 +0000] http "GET /foo?query=t HTTP/1.0" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854



	正则表达式：
	([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)

	*/
	// expr := '([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)'
	expr := ""

	r := regexp.MustCompile(expr)
	// 获取时区
	location, _ := time.LoadLocation("Asia/Shanghai")

	for v := range lp.rc {
		result := r.FindStringSubmatch(v)
		if len(result) != 14 {
			logrus.Infoln("FindStringSubmatch fail: %s", v)
			continue
		}

		message := &Message{}
		t, err := time.ParseInLocation("/02/Jan/2006:15:04:05 +0000", result[4], location)
		if err != nil {
			logrus.Println("ParseInLocation fail:", err.Error(), result[4])
		}
		message.TimeLocal = t

		message.BytesSend, _ = strconv.Atoi(result[8])
		// GET /foo?query=t HTTP/1.0
		requestLine := strings.Split(result[6], " ")
		if len(requestLine) != 3 {
			logrus.Println("strings.Split fail", result[6])
			continue
		}

		message.Method = requestLine[0]

		u, err := url.Parse(requestLine[1])
		if err != nil {
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

func main() {

	reader := &ReadFromFile{
		path: "src/access.log",
	}

	writer := &WriteToInfluxDB{
		influxDBDns: "",
	}

	lp := &LogProcess{
		rc:     make(chan string),
		wc:     make(chan Message),
		reader: reader,
		writer: writer,
	}

	go lp.reader.Read(lp.rc)
	go lp.Process()
	go lp.writer.Write(lp.wc)

	time.Sleep(30 * time.Second)
}
