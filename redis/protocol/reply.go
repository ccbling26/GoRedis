package protocol

import (
	"GoRedis/interface/redis"
	"bytes"
	"fmt"
)

/* ---- Simple String Reply ---- */

type SimpleStringReply struct {
	Data string
}

func MakeSimpleStringReplu(data string) *SimpleStringReply {
	return &SimpleStringReply{
		Data: data,
	}
}

func (r *SimpleStringReply) ToBytes() []byte {
	return []byte(fmt.Sprintf("+%s%s", r.Data, CRLF))
}

/* ---- Error Reply ---- */

// ErrorReply 实现了 redis.Reply 接口
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// StandardErrorReply 实现了 ErrorReply 接口
type StandardErrorReply struct {
	Msg string
}

func MakeErrorReply(msg string) *StandardErrorReply {
	return &StandardErrorReply{
		Msg: msg,
	}
}

func (r *StandardErrorReply) Error() string {
	return r.Msg
}

func (r *StandardErrorReply) ToBytes() []byte {
	return []byte(fmt.Sprintf("-%s%s", r.Msg, CRLF))
}

func IsErrorReply(reply redis.Reply) bool {
	return reply.ToBytes()[0] == '-'
}

/* ---- Int Reply ---- */

type IntReply struct {
	Data int64
}

func MakeIntReply(data int64) *IntReply {
	return &IntReply{
		Data: data,
	}
}

func (r *IntReply) ToBytes() []byte {
	return []byte(fmt.Sprintf(":%d%s", r.Data, CRLF))
}

/* ---- Bulk Reply ---- */

type BulkReply struct {
	Arg []byte
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

func (r *BulkReply) ToBytes() []byte {
	if r.Arg == nil {
		return nullBulkBytes
	}
	return []byte(fmt.Sprintf("$%d%s%s%s", len(r.Arg), CRLF, string(r.Arg), CRLF))
}

/* ---- Multi Bulk Reply ---- */

type MultiBulkReply struct {
	Args [][]byte
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

func (r *MultiBulkReply) ToBytes() []byte {
	argsLength := len(r.Args)
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("*%d%s", argsLength, CRLF))
	for _, arg := range r.Args {
		if arg == nil {
			buffer.WriteString(nullBulkString)
		} else {
			buffer.WriteString(fmt.Sprintf("$%d%s%s%s", len(arg), CRLF, string(arg), CRLF))
		}
	}
	return buffer.Bytes()
}
