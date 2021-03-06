### 功能描述

获取变量分组信息

### 请求参数

{{ common_args_desc }}

#### 接口参数

| 字段         |  类型     | 必选   |  描述   |
|--------------|-----------|--------|---------|
| biz_id       |  string   | 是     | 业务ID  |
| var_group_id |  string   | 否     | 可选， 基于变量分组ID查询 |
| name         |  string   | 否     | 可选， 基于变量分组名称查询 |

### 请求参数示例

```json
{
    "bk_app_code": "xxx",
    "bk_app_secret": "xxx",
    "bk_token": "xxx",
    "biz_id": "xxx",
    "var_group_id": "VG-0b67a798-e9c1-11e9-8c23-525400f99278"
}
```

### 返回结果示例

```json
{
    "result": true,
    "code": 0,
    "message": "OK",
    "data": {
        "biz_id": "XXX",
        "var_group_id": "VG-0b67a798-e9c1-11e9-8c23-525400f99278",
        "name": "group-xxxx",
        "creator": "melo",
        "last_modify_by": "melo",
        "memo": "my variable group",
        "state": 0,
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20"
    }
}
```

### 返回结果参数

#### data

| 字段           | 类型      | 描述    |
|----------------|-----------|---------|
| biz_id         |  string   | 业务ID  |
| var_group_id   |  string   | 变量分组ID  |
| name           |  string   | 变量分组名称|
| memo           |  string   | 备注 |
| state          |  integer  | 状态 默认0: 正常 |
| creator        |  string   | 创建者 |
| last_modify_by |  string   | 修改者 |
| created_at     |  string   | 创建时间 |
| updated_at     |  string   | 更新时间 |
