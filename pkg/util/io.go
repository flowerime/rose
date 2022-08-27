package util

import (
	"encoding/binary"
	"io"
)

// 读 2 个字节，按小端序转为 int
func ReadUint16(r io.Reader) int {
	tmp := make([]byte, 2)
	r.Read(tmp)
	return int(binary.LittleEndian.Uint16(tmp))
}

// 读 4 个字节，按小端序转为 int
func ReadUint32(r io.Reader) int {
	tmp := make([]byte, 4)
	r.Read(tmp)
	return int(binary.LittleEndian.Uint32(tmp))
}

// 字节（小端）转为整数
func BytesToInt(b []byte) int {
	var ret int
	for i, v := range b {
		ret |= int(v) << (i << 3)
	}
	return ret
}
