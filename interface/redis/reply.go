package redis

// Reply RESP（REdis Serialization Protocol） 信息的接口
type Reply interface {
	ToBytes() []byte
}
