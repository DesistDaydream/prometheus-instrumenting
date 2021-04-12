# 练习
gdas_exporter 与 xsky_exporter 逻辑是一样的。属于高级练习，基于 mysql_exporter 的代码改变而来，其中主要是借鉴馆长的 harbo_exporter。包含非常详尽的注释.

其他的 exporter 就不算是练习了~所以没有注释

# 构建
```
docker build -f simulate_mysql_exporter/xsky_exporter/Dockerfile -t lchdzh/xsky-exporter:v0.2 .
docker build -f simulate_mysql_exporter/gdas_exporter/Dockerfile -t lchdzh/gdas-exporter:v0.2 .
```

# 运行
```
docker run --rm --network=host --name xsky-exporter lchdzh/xsky-exporter:v0.2
```
