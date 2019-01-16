### 尚未完成的工作



### 部署
1、部署 grafana
```markdown
docker run -d \
      --restart=always \
      --name=grafana \
      -p 3000:3000 \
      grafana/grafana
```

2、部署 influxdb
```markdown
docker run -d \
      --restart=always \
      --name=influxdb \
      -p 8086:8086 \
      -p 8083:8083 \
      -e INFLUXDB_DB=xdb -e INFLUXDB_ADMIN_ENABLED=true \
      -e INFLUXDB_ADMIN_USER=admin -e INFLUXDB_ADMIN_PASSWORD=Admin321 \
      -e INFLUXDB_USER=xdhuxc -e INFLUXDB_USER_PASSWORD=Xdhuxc123 \
      influxdb
```

3、运行程序
```markdown
go run src/main.go --path src/sccess.log --influxDBDns=http://52.221.216.74:8086@xdhuxc@Xdhuxc123@xdb@s
```


4、生成测试数据
```markdown
go run src/mock_data.go
```

5、创建 grafana 图表
```markdown

```

6、查看系统监控数据
```markdown
curl http://127.0.0.1:9193/monitor
{
        "handleLine": 162,
        "tps": 1,
        "readChanLen": 0,
        "writeChanLen": 0,
        "runTime": "3m35.831542614s",
        "errNum": 0
}
```


QPS
延迟
出口流量
接口QPS
接口延迟
接口出口流量

系统监控



### 编写 prometheus metrics


### 代码地址