package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"time"
)

func main() {
	// 生成日志数据

	/**
	nginx 日志格式为：
		$remote_addr
		$http_x_forwarded_for
	    $remote_user
		[$time_local]
		$scheme "$request"
		$status $body_bytes_sent
		"$http_referer"
	    "$http_user_agent"
	    "$gzip_ratio"
	    $upstream_response_time
	    $request_time
	示例：
	172.0.0.12	-	-	[22/Dec/2017:03:31:35 +0000]	https	"GET /status.html HTTP/1.0"	200	3	"-"	"KeepAliveClient"	"-"	-	0.000
	*/

	file, err := os.OpenFile("src/access.log", os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Open file error: %s", err.Error()))
	}
	defer file.Close()

	// 构造日志数据
	for {

		for i := 1; i < 4; i++ {
			now := time.Now()
			rand.Seed(now.UnixNano())
			paths := []string{"/foo", "/bar", "/baz", "/qux", "/foo", "/bar", "/bar", "/bar"}
			// [0, n)
			path := paths[rand.Intn(len(paths))]
			// [0.0,1.0]
			requestTime := rand.Float64()
			if path == "/foo" {
				requestTime = requestTime + 1.4
			}
			scheme := "http"
			if now.UnixNano()/1000%2 == 1 {
				scheme = "https"
			}
			dateTime := now.Format("02/Jan/2006:15:04:05")

			code := 200

			if now.Unix()%10 == 1 {
				code = 500
			}

			bytesSend := rand.Intn(1000) + 500
			if path == "/foo" {
				bytesSend = bytesSend + 1000
			}

			line := fmt.Sprintf("127.0.0.11 - - [%s +0000] %s \"GET %s HTTP/1.0\" %d %d \"-\" \"KeepAliveClient\" \"-\" - %.3f\n", dateTime, scheme, path, code, bytesSend, requestTime)
			_, err := file.Write([]byte(line))
			if err != nil {
				logrus.Infoln("WriteToFile error: ", err)
			}
			fmt.Println("Write to access.log successfully!")
			time.Sleep(time.Second)
		}

	}
	time.Sleep(time.Millisecond * 2000)
}
