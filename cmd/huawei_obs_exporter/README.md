# 构建
```shell
docker build -f exporter/huawei_obs_exporter/Dockerfile -t lchdzh/huawei-obs-exporter:v0.1 .
```

# 运行
```shell
docker run -d --name huawei-obs-exporter --restart=always \
  --net="host" \
  lchdzh/huawei_obs_exporter:v0.1 \
  --hw-obs-server="IP:PORT" \
  --hw-obs-user="用户名" \
  --hw-obs-pass="密码"
```

## 测试用
```shell
go run exporter/huawei_obs_exporter/main.go --web.listen-address=':18003' --log-level=debug --hw-obs-server="https://172.40.4.17:8088" --hw-obs-user="admin" --hw-obs-pass='Ehualu12#$1'
```
