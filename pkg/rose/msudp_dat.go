package rose

import (
	"bytes"
	"fmt"
	"time"

	util "github.com/flowerime/goutil"
)

type MsUDP struct{ Dict }

func NewMsUDP() *MsUDP {
	d := new(MsUDP)
	d.IsPinyin = false
	d.IsBinary = true
	d.Name = "微软用户自定义短语.dat"
	d.Suffix = "dat"
	return d
}

func (d *MsUDP) Parse() {
	r := bytes.NewReader(d.data)

	// 词库偏移量
	r.Seek(0x10, 0)
	offset_start := ReadUint32(r) // 偏移表开始
	entry_start := ReadUint32(r)  // 词条开始
	entry_end := ReadUint32(r)    // 词条结束
	count := ReadUint32(r)        // 词条数
	table := make(Table, 0, count)
	export_stamp := ReadUint32(r) // 导出的时间戳
	export_time := time.Unix(int64(export_stamp), 0)
	fmt.Println(entry_end, export_time)

	for i := _u32; i < count; i++ {
		r.Seek(int64(offset_start+4*i), 0)
		offset := ReadUint32(r)
		r.Seek(int64(entry_start+offset), 0)
		r.Seek(6, 1)
		pos, _ := r.ReadByte() // 顺序
		r.ReadByte()           // 0x06 不明
		r.Seek(4, 1)           // 4 个空字节
		r.Seek(4, 1)           // 时间戳
		// insert_stamp := ReadUint32(r)
		// insert_time := time.Unix(int64(insert_stamp), 0)
		// insert_time = insert_time.Add(946684800 * time.Second)
		// fmt.Println(insert_time)
		code := make([]byte, 0, 1)
		word := make([]byte, 0, 1)
		tmp := make([]byte, 2)
	CODE:
		r.Read(tmp)
		if !bytes.Equal(tmp, []byte{0, 0}) {
			code = append(code, tmp...)
			goto CODE
		}
	WORD:
		r.Read(tmp)
		if !bytes.Equal(tmp, []byte{0, 0}) {
			word = append(word, tmp...)
			goto WORD
		}
		c, _ := Decode(code, "UTF-16LE")
		w, _ := Decode(word, "UTF-16LE")
		// fmt.Println(c, w)
		table = append(table, &TableEntry{w, c, int(pos)})
	}
	d.table = table
}

func (MsUDP) GenFrom(d *Dict) []byte {
	if d.IsPinyin {
		d.PyToTable("")
	}
	table := d.table
	var buf bytes.Buffer
	now := time.Now()
	export_stamp := util.To4Bytes(uint32(now.Unix()))
	insert_stamp := util.To4Bytes(uint32(now.Add(-946684800 * time.Second).Unix()))
	buf.Write([]byte{0x6D, 0x73, 0x63, 0x68, 0x78, 0x75, 0x64, 0x70,
		0x02, 0x00, 0x60, 0x00, 0x01, 0x00, 0x00, 0x00})
	buf.Write(util.To4Bytes(0x40))
	buf.Write(util.To4Bytes(uint32(0x40 + 4*len(table))))
	buf.Write(make([]byte, 4)) // 待定 文件总长
	buf.Write(util.To4Bytes(uint32(len(table))))
	buf.Write(export_stamp)
	buf.Write(make([]byte, 28))
	buf.Write(make([]byte, 4))

	words := make([][]byte, 0, len(table))
	codes := make([][]byte, 0, len(table))
	sum := 0
	for i := range table {
		word, _ := util.Encode(table[i].Word, "UTF-16LE")
		code, _ := util.Encode(table[i].Code, "UTF-16LE")
		words = append(words, word)
		codes = append(codes, code)
		if i != len(table)-1 {
			sum += len(word) + len(code) + 20
			buf.Write(util.To4Bytes(uint32(sum)))
		}
	}
	for i := range table {
		buf.Write([]byte{0x10, 0x00, 0x10, 0x00})
		// fmt.Println(words[i], len(words[i]), codes[i], len(codes[i]))
		buf.Write(util.To2Bytes(uint16(len(codes[i]) + 18)))
		pos := table[i].Pos
		if pos < 1 {
			pos = 1
		}
		buf.WriteByte(byte(pos))
		buf.WriteByte(0x06)
		buf.Write(make([]byte, 4))
		buf.Write(insert_stamp)
		buf.Write(codes[i])
		buf.Write([]byte{0, 0})
		buf.Write(words[i])
		buf.Write([]byte{0, 0})
	}
	b := buf.Bytes()
	copy(b[0x18:0x1c], util.To4Bytes(uint32(len(b))))
	return b
}
