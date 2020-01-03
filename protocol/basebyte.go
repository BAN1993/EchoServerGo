package protocol

import "encoding/binary"

/*
	1.默认小端
*/

const HeadLen = 4

type Basebyte struct {
	Length uint16
	Xyid   uint16

	data []byte
	index uint16
}

// 将缓存区与外部bytes绑定
// 发送协议时需要调用
func (this *Basebyte) Attach(buf []byte) { this.data = buf }

// 获取当前缓存区长度
func (this *Basebyte) Len() uint16 { return this.Length }

// 获取缓存
func (this *Basebyte) GetBytes() []byte { return this.data }

// 插入包头,这时候len字段还是空的
func (this *Basebyte) PutHead(xyid uint16) {
	this.data = make([]byte, 0)
	this.PutUint16(0) // 一开始是不知道长度的
	this.PutUint16(xyid)
}

// 插入结束,设置len
func (this *Basebyte) PutEnd() {
	this.data[1] = byte(this.Length >> 8)
	this.data[0] = byte(this.Length)
}

func (this *Basebyte) PutUint8(n uint8) {
	b := byte(n)
	this.data = append(this.data, b)
	this.Length += 1
}

func (this *Basebyte) PutInt8(n int8) {
	b := byte(n)
	this.data = append(this.data, b)
	this.Length += 1
}

func (this *Basebyte) PutUint16(n uint16) {
	var b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, n)
	this.data = append(this.data, b...)
	this.Length += 2
}

func (this *Basebyte) PutInt16(n int16) {
	this.PutUint16(uint16(n))
}

func (this *Basebyte) PutUint32(n uint32) {
	var b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	this.data = append(this.data, b...)
	this.Length += 4
}

func (this *Basebyte) PutInt32(n int32) {
	this.PutUint32(uint32(n))
}

func (this *Basebyte) PutUint64(n uint64) {
	var b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	this.data = append(this.data, b...)
	this.Length += 8
}

func (this *Basebyte) PutInt64(n int64) {
	this.PutUint64(uint64(n))
}

func (this *Basebyte) PutString(s string) {
	var slen uint32	= uint32(len(s))
	if slen < 0xff {
		this.PutUint8(uint8(slen))
	} else if slen < 0xfffe {
		this.PutUint8(0xff)
		this.PutUint16(uint16(slen))
	} else {
		this.PutUint8(0xff)
		this.PutUint16(0xffff)
		this.PutUint32(slen)
	}
	this.data = append(this.data, []byte(s)...)
	this.Length += uint16(slen)
}

// 获取包头,并绑定缓冲区
// 缓冲区小于包头或包体大小返回false
// 返回true说明至少能组成一个包,切包头已经解析完成
func (this *Basebyte) GetHeadAndAttach(buf []byte) bool {
	if len(buf) < HeadLen 		{ return false }
	// 先将长度取出,判断整个缓冲区数据是否足够
	plen := binary.LittleEndian.Uint16(buf)
	if plen <= 0 				{ return false }
	if len(buf) < int(plen)		{ return false }

	this.Attach(buf)
	this.Length = this.GetUint16()
	this.Xyid = this.GetUint16()
	return true
}

func (this *Basebyte) GetUint8() uint8 {
	ret := this.data[this.index]
	this.index += 1
	return uint8(ret)
}

func (this *Basebyte) GetInt8() int8 {
	ret := this.data[this.index]
	this.index += 1
	return int8(ret)
}

func (this *Basebyte) GetUint16() uint16 {
	ret := binary.LittleEndian.Uint16(this.data[this.index:this.index+4])
	this.index += 2
	return ret
}

func (this *Basebyte) GetInt16() int16 {
	return int16(this.GetUint16())
}

func (this *Basebyte) GetUint32() uint32 {
	ret := binary.LittleEndian.Uint32(this.data[this.index:this.index+4])
	this.index += 4
	return ret
}

func (this *Basebyte) GetInt32() int32 {
	return int32(this.GetUint32())
}

func (this *Basebyte) GetUint64() uint64 {
	ret := binary.LittleEndian.Uint64(this.data[this.index:this.index+8])
	this.index += 8
	return ret
}

func (this *Basebyte) GetInt64() int64 {
	return int64(this.GetUint64())
}

func (this *Basebyte) GetString() string {
	blen := this.GetUint8()
	if blen < 0xff {
		return this.ReadString(uint32(blen))
	}

	wlen := this.GetUint16()
	if wlen < 0xfffe {
		return this.ReadString(uint32(wlen))
	}

	dwlen := this.GetUint32()
	return this.ReadString(dwlen)
}

func (this *Basebyte) ReadString(len uint32) string {
	ret := string(this.data[this.index:this.index + uint16(len)])
	this.index += uint16(len)
	return ret
}
