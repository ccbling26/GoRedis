package protocol

import (
	"GoRedis/interface/redis"
	"bytes"
)

var CRLF = "\r\n" // 在 RESP（REdis Serialization Protocol）中，CRLF 作为分行符号

/* ---- Pong Reply ---- */

type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")

func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

func MakePongReply() *PongReply {
	return &PongReply{}
}

/* ---- OK Reply ---- */

type OKReply struct{}

var okBytes = []byte("+OK\r\n")

func (r *OKReply) ToBytes() []byte {
	return okBytes
}

func MakeOKReply() *OKReply {
	return &OKReply{}
}

/* ---- Null Bulk Reply ---- */

type NullBulkReply struct{}

var nullBulkString = "$-1\r\n"
var nullBulkBytes = []byte(nullBulkString)

func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

/* ---- Empty Bulk Reply ---- */

type EmptyMultiBulkReply struct{}

var emptyMultiBulkBytes = []byte("*0\r\n")

func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

func IsEmptyMultiBulkReply(reply redis.Reply) bool {
	return bytes.Equal(reply.ToBytes(), emptyMultiBulkBytes)
}

/* ---- No Reply ---- */

type NoReply struct{}

var noBytes = []byte("")

func (r *NoReply) ToBytes() []byte {
	return noBytes
}

/* ---- Queued Reply ---- */

type QueuedReply struct{}

var queuedBytes = []byte("+QUEUED\r\n")

func (r *QueuedReply) ToBytes() []byte {
	return queuedBytes
}

func MakeQueuedReply() *QueuedReply {
	return &QueuedReply{}
}
