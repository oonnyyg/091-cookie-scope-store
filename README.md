# cookie-scope-store

离线 Cookie 仓库：逐行接收 JSON 指令，按作用域规则为目标 URL 选出应携带的
Cookie，并把持久记录保存到状态目录。仅使用 Go 标准库。

## 运行

```
go run ./cmd/cookiejar --state 状态目录
```

从标准输入逐行读取 JSON 对象，逐行向标准输出写出 JSON 应答。

## 协议

接收 Cookie：

```json
{"op":"receive","url":"https://sub.example.com/a/b","time":100,
 "cookie":{"name":"sid","value":"abc","domain":".example.com",
           "path":"/a","secure":true,"max_age":3600}}
```

查询应携带的 Cookie：

```json
{"op":"query","url":"https://sub.example.com/a/b/c?x=1","time":200}
```

应答：

```json
{"ok":true}                      receive 成功
{"ok":true,"header":"sid=abc"}   query 结果，即 Cookie 请求头
{"ok":false,"error":"..."}       失败；仓库不被修改
```

`time` 为必填的非负整数秒，且不得倒退（重开进程后同样生效，状态目录中记录
了最近时间）。

## 规则摘要

- 仅支持 `http`/`https`、非 IP 的 ASCII 域名；端口不参与匹配。
- Cookie 名称非空，名称与值仅含 ASCII 字母或数字，值可为空。
- `domain` 省略时为主机 Cookie，只匹配来源主机；指定时忽略大小写与一个
  前导点，须等于来源域或为其按点分隔的后缀，之后按同样规则匹配子域。
- `path` 省略时取来源路径最后一个斜线之前的部分，为空则为 `/`；显式路径
  须以 `/` 开头。匹配按 RFC 6265：相等，或为前缀且以 `/` 结尾或紧接 `/`；
  查询参数不参与匹配。
- `secure` 为 true 的 Cookie 只发给 HTTPS。
- `max_age` 为正时从接收时刻计时，到期即失效；非正时删除同名同域同路径
  记录；省略时为会话 Cookie，重开后消失。持久记录写入 `状态目录/state.json`。
- 同 (name, domain, path) 的记录被替换但保留原创建次序；不同路径的同名
  记录可共存。
- 查询结果路径长者在前，同长按创建次序，同名不合并。

## 测试

```
go test ./...
```
