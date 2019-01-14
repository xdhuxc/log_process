### 尚未完成的工作



### 部署
1、部署 grafana
```markdown
docker run -d \
      --restart=always \
      --name=grafana 
      -p 3000:3000 grafana/grafana
```

2、部署 influxdb
```markdown
docker run -d \
      --restart=always \
      --name=influxdb \
      -p 8086:8086 \
      -p 8083:8083 \
      -e INFLUXDB_DB=db0 -e INFLUXDB_ADMIN_ENABLED=true \
      -e INFLUXDB_ADMIN_USER=admin -e INFLUXDB_ADMIN_PASSWORD=Admin321 \
      -e INFLUXDB_USER=texadg -e INFLUXDB_USER_PASSWORD=Texadg123 \
      -v /data/influxdb:/var/lib/influxdb \
      influxdb
```


3、运行程序
```markdown
go run src/main.go --path src/sccess.log --influxDBDns=http://52.221.216.74:8086@texadg@Texadg123@xdb@s
```


4、生成测试数据
```markdown
./mock_data
```

5、创建 grafana 图表
```markdown

```

### 编写 prometheus metrics


### 代码地址