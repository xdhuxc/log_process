package main

import (
	"strings"
	"fmt"
	"time"
	"os"
	"bufio"
	"io"
)

type Reader interface {
	Read(rc chan string)
}

type Writer interface {
	Write(wc chan string)
}

type LogProcess struct {
	// 从读取模块向解析模块传输数据
	rc chan string
	// 从解析模块向写入模块传输数据
	wc chan string
	reader Reader
	writer Writer

}

type ReadFromFile struct {
	// 日志文件路径
	path string
}

type WriteToInfluxDB struct {
	// influxDB 数据源信息
	influxDBDsn string
}

/**
	写入模块，负责将数据写入到 InfluxDB 中
 */
func (w *WriteToInfluxDB) Write(wc chan string) {

	for v := range wc {
		fmt.Println(v)
	}
}


/**
	读取模块，负责从文件中读取日志记录
 */
func (r *ReadFromFile) Read(rc chan string) {
	/**
		此处使用指针的原因：
		1、如果结构体很大的话，使用指针则不用复制，在性能上会有很大的优势。
		2、使用指针可以直接修改结构体参数的值，而不是修改某一个结构体实例的参数值
   */
   // 打开文件
   f, err := os.Open(r.path)
   if err != nil {
   		panic(fmt.Sprintf("open file error: %s", err.Error()))
   }

   // 从文件末尾开始逐行读取文件内容
   f.Seek(0, 2)
   readBuffer := bufio.NewReader(f)

   for {
	   // 读取文件直到遇到 \n
	   line, err := readBuffer.ReadBytes('\n')
	   if err == io.EOF {//读取文件到结尾时处理
	   		time.Sleep(500 * time.Millisecond)
	   		continue
	   } else if err != nil {
	   		panic(fmt.Sprintf("ReadBytes error: %s", err.Error()))
	   }

	   rc <- string(line[:len(line)-1])
   }



}

/**
	解析模块，负责解析读到的数据
 */
func (lp *LogProcess) Process() {
	// 循环读取 rc 中的内容
	for v := range lp.rc {
		lp.wc <- strings.ToUpper(v)
	}
}


func main() {
	r := &ReadFromFile{
		path: "./access.log",
	}

	w := &WriteToInfluxDB{
		influxDBDsn: "",
	}


	// 初始化 LogProcess 实例
	lp := &LogProcess{
		rc: make(chan string),
		wc: make(chan string),
		reader: r,
		writer: w,
	}

	go lp.reader.Read(lp.rc)
	go lp.Process()
	go lp.writer.Write(lp.wc)

	time.Sleep(30)
}
