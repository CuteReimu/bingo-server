# bingo-server

![](https://img.shields.io/github/languages/top/Touhou-Freshman-Camp/bingo-server "语言")
[![](https://img.shields.io/github/workflow/status/Touhou-Freshman-Camp/bingo-server/Go)](https://github.com/Touhou-Freshman-Camp/bingo-server/actions/workflows/golangci-lint.yml "代码分析")
[![](https://img.shields.io/github/contributors/Touhou-Freshman-Camp/bingo-server)](https://github.com/Touhou-Freshman-Camp/bingo-server/graphs/contributors "贡献者")
[![](https://img.shields.io/github/license/Touhou-Freshman-Camp/bingo-server)](https://github.com/Touhou-Freshman-Camp/bingo-server/blob/master/LICENSE "许可协议")

## 协议

协议全部采用json的格式

| 字段    | 类型  | 备注                                      |
|-------|-----|-----------------------------------------|
| name  | str | 协议名                                     |
| reply | str | 回应的协议名，如果只是推送协议则没有这个字段                  |
| data  | obj | 协议内容，下文一一列举（如果返回协议体为空，则没有这个字段，以便减小协议大小） |

<details><summary>查看示例：</summary>

```json
{
    "name": "error_sc",
    "data": {
      "code": 1,
      "msg": "create room failed"
    }
}
```

</details>

### 下行协议

下行协议一共只有四种：错误信息协议、心跳返回、房间内的全量同步协议、房间外的全量同步协议

**错误信息协议: error_sc**

| 字段   | 类型  | 备注               |
|------|-----|------------------|
| code | int | 错误码，0表示无错误       |
| msg  | str | 错误信息，code为0就无此字段 |

**心跳返回: heart_sc**

| 字段   | 类型  | 备注          |
|------|-----|-------------|
| time | int | 服务端时间戳，单位毫秒 |

**房间外的全量同步协议: global_info_sc**

| 字段  | 类型  | 备注  |
|-----|-----|-----|

**房间内的全量同步协议: room_info_sc**

| 字段    | 类型         | 备注                                    |
|-------|------------|---------------------------------------|
| name  | str        | 自己的用户名                                |
| rid   | str        | 房间号                                   |
| type  | int        | 房间类别，同上                               |
| host  | str        | 主持人的名字                                |
| names | Array[str] | 所有选手的用户名，含自己，数组长度就是选手人数。没进入的位置会留个空字符串 |
| game  | obj        | 根据不同玩法，数据不同，待定                        |

### 登录相关

**登录协议: login_cs**

| 字段    | 类型  | 备注  |
|-------|-----|-----|
| token | str | 识别码 |

**心跳请求协议: heart_cs**

| 字段  | 类型  | 备注  |
|-----|-----|-----|

### 创建/进入房间协议

**创建房间: create_room_cs**

| 字段   | 类型  | 备注                                 |
|------|-----|------------------------------------|
| name | str | 用户名                                |
| rid  | str | 房间号                                |
| type | int | 房间类别，1：bingo标准赛，2：bingo BP赛，3：大富翁。 |

**进入房间：join_room_cs**

| 字段   | 类型  | 备注  |
|------|-----|-----|
| name | str | 用户名 |
| rid  | str | 房间号 |
