package parser

import (
	"GoRedis/interface/redis"
	"GoRedis/lib/logger"
	"GoRedis/redis/protocol"
	"bufio"
	"bytes"
	"errors"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

type PayLoad struct {
	Data redis.Reply
	Err  error
}

// ParseStream 通过 io.Reader 读取数据并将结果通过 channel 将结果返回给调用者
func ParseStream(reader io.Reader) <-chan *PayLoad {
	ch := make(chan *PayLoad)
	go parse(reader, ch) // parse0 will close the channel
	return ch
}

// ParseOne 解析 []byte 并返回 redis.Reply
func ParseOne(data []byte) (redis.Reply, error) {
	reader := bytes.NewReader(data)
	ch := ParseStream(reader)
	payload := <-ch
	if payload == nil {
		return nil, errors.New("no reply")
	}
	return payload.Data, payload.Err
}

func parse(rawReader io.Reader, ch chan *PayLoad) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err, string(debug.Stack()))
		}
	}()

	reader := bufio.NewReader(rawReader)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			ch <- &PayLoad{Err: err}
			close(ch)
			return
		}
		if len(line) < 2 || line[len(line)-2] != '\r' {
			continue
		}
		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		switch line[0] {
		case '+': // 简单字符串
			content := string(line[1:])
			ch <- &PayLoad{Data: protocol.MakeSimpleStringReplu(content)}
			if strings.HasPrefix(content, "FULLRESYNC") {
				err = parseRDBBulkString(reader, ch)
				if err != nil {
					ch <- &PayLoad{Err: err}
					close(ch)
					return
				}
			}
		case '-': // 错误
			ch <- &PayLoad{
				Data: protocol.MakeErrorReply(string(line[1:])),
			}
		case ':': // 整数
			value, err := strconv.ParseInt(string(line[1:]), 10, 64)
			if err != nil {
				protocolError(ch, "illegal number "+string(line[1:]))
				continue
			}
			ch <- &PayLoad{
				Data: protocol.MakeIntReply(value),
			}
		case '$': // 字符串
			err = parseBulkString(line, reader, ch)
			if err != nil {
				ch <- &PayLoad{Err: err}
				close(ch)
				return
			}
		case '*': // 数组
			err = parseArray(line, reader, ch)
			if err != nil {
				ch <- &PayLoad{Err: err}
				close(ch)
				return
			}
		default:
			args := bytes.Split(line, []byte{' '})
			ch <- &PayLoad{Data: protocol.MakeMultiBulkReply(args)}
		}
	}
}

// parseRDBBulkString RDB 和 AOF 中没有 CRLF，需要额外处理
func parseRDBBulkString(reader *bufio.Reader, ch chan<- *PayLoad) error {
	header, err := reader.ReadBytes('\n')
	if err != nil {
		return err
	}
	header = bytes.TrimSuffix(header, []byte{'\r', '\n'})
	if len(header) == 0 {
		return errors.New("empty header")
	}
	strLen, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || strLen <= 0 {
		return errors.New("illegal bulk header: " + string(header))
	}
	body := make([]byte, strLen)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	ch <- &PayLoad{Data: protocol.MakeBulkReply(body)}
	return nil
}

func parseBulkString(header []byte, reader *bufio.Reader, ch chan<- *PayLoad) error {
	strLen, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || strLen < -1 {
		protocolError(ch, "illeagal bulk string header: "+string(header))
		return nil
	} else if strLen == -1 {
		ch <- &PayLoad{Data: protocol.MakeNullBulkReply()}
		return nil
	}
	body := make([]byte, strLen+2)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	ch <- &PayLoad{Data: protocol.MakeBulkReply(body[:len(body)-2])}
	return nil
}

func parseArray(header []byte, reader *bufio.Reader, ch chan<- *PayLoad) error {
	numOfStr, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil {
		protocolError(ch, "illegal array header: "+string(header[1:]))
		return nil
	} else if numOfStr == 0 {
		ch <- &PayLoad{Data: protocol.MakeEmptyMultiBulkReply()}
		return nil
	}
	lines := make([][]byte, 0, numOfStr)
	for i := int64(0); i < numOfStr; i++ {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		length := len(line)
		if length < 4 || line[length-2] != '\r' || line[0] != '$' {
			protocolError(ch, "illeagal bulk string header: "+string(line))
			break
		}
		strLen, err := strconv.ParseInt(string(line[1:length-2]), 10, 64)
		if err != nil || strLen < -1 {
			protocolError(ch, "illegal bulk string length "+string(line))
			break
		} else if strLen == -1 {
			lines = append(lines, []byte{})
		} else {
			body := make([]byte, strLen+2)
			_, err := io.ReadFull(reader, body)
			if err != nil {
				return err
			}
			lines = append(lines, body[:len(body)-2])
		}
	}
	ch <- &PayLoad{Data: protocol.MakeMultiBulkReply(lines)}
	return nil
}

func protocolError(ch chan<- *PayLoad, msg string) {
	err := errors.New("protocol error: " + msg)
	ch <- &PayLoad{
		Err: err,
	}
}
