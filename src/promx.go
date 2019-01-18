package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests")

func main() {

	flag.Parse()

	// 初始化一个容器
	diskPercent := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		// 指标名称
		Name: "memory_percent",
		// 帮助信息
		Help: "memory used percent",
	},
		[]string{"percent", "label_key_1", "label_key_2", "abc"})

	prometheus.MustRegister(diskPercent)

	for i := 0; i < 10; i++ {
		logrus.Println("Start to collect memory used percent.")
		v, err := mem.VirtualMemory()
		if err != nil {
			logrus.Println("get memory error: %s", err)
		}
		diskPercent.WithLabelValues("usedMemory", "label_value_1", "label_value_2", "cdef").Set(v.UsedPercent)
		time.Sleep(time.Second)
	}

	http.Handle("/metrics", promhttp.Handler())
	logrus.Infoln(http.ListenAndServe(*addr, nil))
}
