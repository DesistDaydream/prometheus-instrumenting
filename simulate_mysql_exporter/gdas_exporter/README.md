# 概述
从 ceph-mgr 的 8443 端口获取 对象存储 的指标数据

# 构建
```shell
docker build -f gdas_exporter/Dockerfile -t lchdzh/gdas-exporter:v0.2.0 .
```

# 运行
## exporter 请求控制台代理
部署在湖外
```shell
docker run -d --name=gdas-exporter --net=host \
lchdzh/gdas-exporter:v0.2.0 \
--web.listen-address=':18003' \
--gdas-server='http://gdas-proxy.unicom.cqcq.ehualu.it:38443'
```


## 测试
```shell
go run simulate_mysql_exporter/gdas_exporter/main.go --web.listen-address=':18003' --log-level=debug --gdas-server='http://gdas-proxy.unicom.cqcq.ehualu.it:38443'
```

```shell
docker run --rm --name=gdas-exporter --net=host lchdzh/gdas-exporter:v0.2.0 --web.listen-address=':18003' --log-level=debug --gdas-server='http://gdas-proxy.unicom.cqcq.ehualu.it:38443'
```