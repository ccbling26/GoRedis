# GoRedis

Golang 的 `netpoller` 基于 IO 多路复用和 `goroutine scheduler` 构建了一个简洁高性能的网络模型，并给开发者提供了 `goroutine-per-connection` 风格的极简接口。具体剖析可以参考 [Go 语言设计与实现-网络轮询器](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-netpoller/)



## Redis 通信协议

Redis 自 2.0 版本起使用了统一的协议 RESP（REdis Serialization Protocol）。

RESP 是一个二进制安全的文本协议，工作在 TCP 协议上，以行为单位读取数据，客户端和服务端发送的命令或数据一律以 `\r\n`（CRLF）作为换行符

> 二进制安全：指允许协议中出现任意字符而不会导致故障
>
> C 语言的字符串以 `\0` 为结尾，不允许字符串中间出现 `\0`，因此 C 语言字符串不是二进制安全的

RESP 定义了 5 种格式

- 简单字符串（Simple String）：服务器用来返回简单的结果，比如 `OK`。非二进制安全，不允许换行
- 错误信息（Error）：服务器用来返回简单的错误信息，比如 `ERR Invalid Syntax`。非二进制安全，不允许换行
- 整数（Integer）：`llen`、`scard` 等命令的返回值，64 位有符号整数
- 字符串（Bulk String）：二进制安全字符串，比如 `get` 等命令的返回值
- 数组（Array，又称 Multi Bulk String）：Bulk String 数组，客户端发送指令以及 `lrange` 等命令响应的格式



RESP 通过第一个字符来表示格式

- 简单字符串：以 `+` 开始， 比如 `+OK\r\n`
- 错误：以 `-` 开始，比如 `-ERR Invalid Synatx\r\n`
- 整数：以 `:` 开始，比如 `:1\r\n`
- 字符串：以 `$` 开始
- 数组：以 `*` 开始



Bulk String 有两行，第一行为 `$` + 正文长度，第二行为实际内容

```bash
$4
a\r\nb

# 将换行符打印出来
$4\r\na\r\nb\r\n
```

> `$-1` 表示 `nil`，比如使用 `get` 命令查询一个不存在的 `key` 时，响应即为 `$-1`



Array 格式第一行为 `*` + 数组长度，其后是相应数量的 Bulk String

```bash
*3
$3
SET
$3
key
$5
value

# 将换行符打印出来
*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
```



## 协议解析器

用于将 Socket 传来的数据还原成 `[][]byte` 格式

来自客户端的请求均为数组格式，它在第一行中标记报文的总行数并使用 `CRLF` 作为分行符



## 参考

[HDT3213-godis](https://github.com/HDT3213/godis)
