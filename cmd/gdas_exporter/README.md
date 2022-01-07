# 概述


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
--gdas-server='https://172.38.30.192:8003'
--gdas-pass=''
```


## 测试
```shell
go run exporter/gdas_exporter/main.go --web.listen-address=':18003' --log-level=debug --gdas-server='https://172.38.30.192:8003' --gdas-pass='Euleros!@#123'
```

```shell
docker run --rm --name=gdas-exporter --net=host lchdzh/gdas-exporter:v0.2.0 --web.listen-address=':18003' --log-level=debug --gdas-server='https://172.38.30.192:8003'
```
