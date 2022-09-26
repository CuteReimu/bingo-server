# bingo-server

![](https://img.shields.io/github/languages/top/Touhou-Freshman-Camp/bingo-server "语言")
[![](https://img.shields.io/github/workflow/status/Touhou-Freshman-Camp/bingo-server/Go)](https://github.com/Touhou-Freshman-Camp/bingo-server/actions/workflows/golangci-lint.yml "代码分析")
[![](https://img.shields.io/github/contributors/Touhou-Freshman-Camp/bingo-server)](https://github.com/Touhou-Freshman-Camp/bingo-server/graphs/contributors "贡献者")
[![](https://img.shields.io/github/license/Touhou-Freshman-Camp/bingo-server)](https://github.com/Touhou-Freshman-Camp/bingo-server/blob/master/LICENSE "许可协议")

## 协议

协议全部采用json的格式

| 字段    | 类型  | 备注                     |
|-------|-----|------------------------|
| name  | str | 协议名                    |
| reply | str | 回应的协议名，如果只是推送协议则没有这个字段 |
| data  | obj | 协议内容，下文一一列举            |

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

### 基础协议

**错误信息协议: error_sc**

| 字段   | 类型  | 备注         |
|------|-----|------------|
| code | int | 错误码，0表示无错误 |
| msg  | str | 错误信息       |

**心跳: heart_cs**

| 字段  | 类型  | 备注  |
|-----|-----|-----|

**心跳返回: heart_sc**

| 字段   | 类型  | 备注          |
|------|-----|-------------|
| time | int | 服务端时间戳，单位毫秒 |

### 创建/进入房间协议

**创建房间: create_room_cs**
**创建房间返回: create_room_sc**
**进入房间：join_room_cs**

| 字段    | 类型  | 备注                                |
|-------|-----|-----------------------------------|
| token | str | 识别码                               |
| name  | str | 用户名                               |
| rid   | str | 房间号                               |
| type  | int | 房间类别，1：bingo标准赛，2：bingo BP赛，3：大富翁。（进入房间不需要填这个字段） |

**进入房间返回: join_room_sc**

| 字段    | 类型         | 备注                                    |
|-------|------------|---------------------------------------|
| name  | str        | 用户名                                   |
| rid   | str        | 房间号                                   |
| type  | int        | 房间类别，同上                               |
| names | Array[str] | 所有选手的用户名，含自己，数组长度就是选手人数。没进入的位置会留个空字符串 |

**通知其他人进入房间（转播和选手通用）: join_room_ntf_sc**

| 字段    | 类型  | 备注                                |
|-------|-----|-----------------------------------|
| name  | str | 加入的人的用户名                          |

**通知其他人离开房间（转播和选手通用）: leave_room_ntf_sc**

| 字段    | 类型  | 备注              |
|-------|-----|-----------------|
| index | str | 加入的人的index，从0开始 |
