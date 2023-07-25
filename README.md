# bingo-server

![](https://img.shields.io/github/go-mod/go-version/CuteReimu/bingo-server "语言")
[![](https://img.shields.io/github/actions/workflow/status/CuteReimu/bingo-server/golangci-lint.yml?branch=master)](https://github.com/CuteReimu/bingo-server/actions/workflows/golangci-lint.yml "代码分析")
[![](https://img.shields.io/github/contributors/CuteReimu/bingo-server)](https://github.com/CuteReimu/bingo-server/graphs/contributors "贡献者")
[![](https://img.shields.io/github/license/CuteReimu/bingo-server)](https://github.com/CuteReimu/bingo-server/blob/master/LICENSE "许可协议")

## 协议

***因为协议会频繁更改，所以下行协议请暂时不要参照此表格***

协议全部采用json的格式

| 字段      | 类型  | 备注                                      |
|---------|-----|-----------------------------------------|
| name    | str | 协议名                                     |
| reply   | str | 回应的协议名，如果只是推送协议则没有这个字段                  |
| trigger | str | 触发事件的玩家的名字，如果没有则没有这个字段                  |
| data    | obj | 协议内容，下文一一列举（如果返回协议体为空，则没有这个字段，以便减小协议大小） |

示例：

```json
{
  "name": "error_sc",
  "reply": "join_room_cs",
  "trigger": "xxx",
  "data": {
    "code": 1,
    "msg": "create room failed"
  }
}
```

协议与`org.tfcc.bingo.message`下的类(`Dispatcher.kt`、`Handler.kt`、`Message.kt`除外)一一对应。

例如`"name": "error_sc"`对应ErrorSc

## 贡献

持久化存储采用protobuf进行序列化。当你需要修改其结构时，先编辑`data.proto`，然后执行`go generate`即可。
