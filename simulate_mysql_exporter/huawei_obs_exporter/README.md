# 构建
```shell
docker build -f huawei_obs_exporter/Dockerfile -t lchdzh/huawei-obs-exporter:v0.1 .
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

测试用
```shell
docker run -d --name huawei-obs-exporter --restart=always \
  --net="host" \
  lchdzh/huawei-obs-exporter:v0.1 \
  --hw-obs-server="IP:PORT" \
  --hw-obs-user="用户名" \
  --hw-obs-pass="密码"
```
