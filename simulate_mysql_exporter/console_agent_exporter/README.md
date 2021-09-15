# 构建
```shell
docker build -f simulate_mysql_exporter/xsky_exporter/Dockerfile -t lchdzh/xsky-exporter:v0.6 .
docker build -f simulate_mysql_exporter/gdas_exporter/Dockerfile -t lchdzh/gdas-exporter:v0.6 .
docker build -f simulate_mysql_exporter/consoler_exporter/Dockerfile -t lchdzh/consoler-exporter:v0.3-proxy .
```

# 运行
## exporter 请求控制台代理
```shell
docker run -d --name=consoler-exporter-heb --net=host lchdzh/consoler-exporter:test-proxy --web.listen-address=':9122' --consoler-server='http://192.168.11.2:9097' --region-id=971
```

测试用
```shell
go run simulate_mysql_exporter/consoler_exporter/main.go --web.listen-address=':8080' --consoler-server='http://gdas-proxy.heb.ehualu.it:38443' --region-id=971 --log-level=debug
```
