# 构建
```shell
docker build -f exporter/console_agent_exporter/Dockerfile -t console-agent-exporter:v0.5.0-proxy .
```

# 运行
## exporter 请求控制台代理
部署在湖外
```shell
docker run -d --name=console-agent-exporter --net=host lchdzh/console-agent-exporter:v0.5.0-proxy --web.listen-address=':9122' --console-agent-server='http://gdas-proxy.sxxa.ehualu.it:38443' --region-id=841
```

部署在湖内
```shell
docker run -d --name=console-agent-exporter --net=host lchdzh/console-agent-exporter:v0.5.0-proxy --web.listen-address=':9122' --console-agent-server='http://localhost:9097'
```

## 测试
```shell
go run exporter/console_agent_exporter/main.go --web.listen-address=':9122' --console-agent-server='http://gdas-proxy.unicom.cqcq.ehualu.it:38443' --log-level=debug 
```

```shell
docker run --rm --name=console-agent-exporter --net=host lchdzh/console-agent-exporter:v0.5.0-proxy --web.listen-address=':9122' --console-agent-server='http://gdas-proxy.unicom.cqcq.ehualu.it:38443' --log-level=debug
```
