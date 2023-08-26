package pinyin

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/nopdan/rose/encoder"
	"github.com/nopdan/rose/util"
)

type Custom struct {
	Template
	Sep   byte // 分隔符
	PySep byte // 拼音分隔符

	// w 词，c 无前缀拼音，p 有前缀拼音，f 词频
	Rule []string
}

func NewCustom(rule string) *Custom {
	f := new(Custom)
	s := strings.Split(rule, "|")
	matchSep := func(s string) byte {
		switch s {
		case "t":
			return '\t'
		case "s":
			return ' '
		}
		return s[0]
	}
	f.Sep = matchSep(s[0])
	f.PySep = matchSep(s[1])
	if f.PySep == 0 || f.Sep == 0 {
		return nil
	}
	f.Rule = s[2:]
	return f
}

func NewSogou() *Custom {
	f := NewCustom("s|'|p|w")
	f.Name = "搜狗拼音.txt"
	f.ID = "sogou"
	return f
}

func NewQQ() *Custom {
	f := NewCustom("s|'|c|w|f")
	f.Name = "QQ 拼音.txt"
	f.ID = "qq"
	return f
}

func NewBaidu() *Custom {
	f := NewCustom("t|'|w|c|f")
	f.Name = "百度拼音.txt"
	f.ID = "baidu"
	return f
}

func NewGoogle() *Custom {
	f := NewCustom("t|s|w|f|c")
	f.Name = "谷歌拼音.txt"
	f.ID = "google"
	return f
}

func NewRime() *Custom {
	f := NewCustom("t|s|w|c|f")
	f.Name = "rime 拼音.txt"
	f.ID = "rime"
	return f
}

func (f *Custom) Unmarshal(r *bytes.Reader) []*Entry {
	di := make([]*Entry, 0, r.Size()>>8)

	enc := encoder.New("pinyin")
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		e := strings.Split(scan.Text(), string(f.Sep))
		if len(e) < len(f.Rule) {
			continue
		}
		var word string
		pinyin := make([]string, 0, 1)
		freq := 1
		for i, v := range f.Rule {
			switch v {
			case "w":
				word = e[i]
			case "f":
				freq, _ = strconv.Atoi(e[i])
			case "c", "p":
				tmp := strings.TrimLeft(e[i], string(f.PySep))
				pinyin = strings.Split(tmp, string(f.PySep))
			}
		}
		// 纯词生成拼音
		if len(pinyin) == 0 {
			pinyin = enc.Encode(word)
		}
		di = append(di, &Entry{word, pinyin, freq})
	}
	return di
}

// 拼音通用格式生成
func (f *Custom) Marshal(di []*Entry) []byte {
	var buf bytes.Buffer
	for _, v := range di {
		for i, val := range f.Rule {
			if i != 0 {
				buf.WriteByte(f.Sep)
			}
			switch val {
			case "w":
				buf.WriteString(v.Word)
			case "f":
				buf.WriteString(strconv.Itoa(v.Freq))
			case "c", "p":
				if f.Rule[i] == "p" {
					buf.WriteByte(f.PySep)
				}
				pinyin := strings.Join(v.Pinyin, string(f.PySep))
				buf.WriteString(pinyin)
			}
		}
		buf.WriteString(util.LineBreak)
	}
	return buf.Bytes()
}
