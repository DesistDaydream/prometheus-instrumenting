# 业务 Exporter

## 获取 TOKEN
```
curl -XPOST'http://10.20.5.98:8056/api/v1/auth/tokens:login'   -H 'Content-Type: application/json'   --data-binary '{"auth":{"name":"admin","password":"admin"}}'
```
响应信息
```
{
  "token": {
    "create": "2020-12-30T10:30:51.675098636+08:00",
    "expires": "2020-12-30T11:30:51.675086075+08:00",
    "roles": [
      "admin"
    ],
    "user": {
      "create": "2020-05-28T03:40:52.284746Z",
      "email": "a@a.com",
      "enabled": true,
      "external_id": "",
      "id": 3,
      "identity_platform": null,
      "name": "admin",
      "password_last_update": "2020-05-28T03:40:52.284749Z",
      "roles": null
    },
    "uuid": "de196a59254343d297d133cf24954e98",
    "valid": true
  }
}

```

## 获取使用量
```
curl 'http://10.20.5.98:8056/api/v1/cluster' -H 'Cookie:XMS_AUTH_TOKEN=de196a59254343d297d133cf24954e98'
```

响应信息
```
{
  "cluster": {
    "access_token": null,
    "access_url": "",
    "create": "2020-05-28T03:33:26.914084Z",
    "disk_lighting_mode": "automatic",
    "down_out_interval": 10800,
    "elasticsearch_enabled": true,
    "fs_id": "cbf8ce5d-45b8-43af-850c-f1b2f2944007",
    "id": 1,
    "maintained": false,
    "name": "EHL",
    "os_gateway_oplog_switch": false,
    "samples": [
      {
        "actual_kbyte": 472029616128,
        "create": "2020-12-30T10:33:12+08:00",
        "data_kbyte": 95215078037,
        "degraded_percent": 0,
        "error_kbyte": 0,
        "healthy_percent": 1,
        "os_down_bandwidth_kbyte": 0,
        "os_down_iops": 29,
        "os_merge_speed": 0,
        "os_up_bandwidth_kbyte": 13155,
        "os_up_iops": 1,
        "read_bandwidth_kbyte": 539741,
        "read_iops": 583956,
        "read_latency_us": 128,
        "recovery_bandwidth_kbyte": 0,
        "recovery_iops": 0,
        "recovery_percent": 0,
        "total_kbyte": 410013757440,
        "unavailable_percent": 0,
        "used_kbyte": 87030048128,
        "write_bandwidth_kbyte": 14632,
        "write_iops": 4698,
        "write_latency_us": 1586
      }
    ],
    "snmp_enabled": false,
    "stats_reserved_days": 90,
    "status": "",
    "update": "2020-05-28T03:33:26.914087Z",
    "version": "SDS_4.2.009.5"
  }
}
```
