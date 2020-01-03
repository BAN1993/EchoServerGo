package protocol

const (
	XYID_HEARTBEAT uint16 = 1
)

type HeartBeat struct {
	Basebyte
	Timestamp uint32
}
func (this *HeartBeat) Encode() []byte {
	this.PutHead(XYID_HEARTBEAT)
	this.PutUint32(this.Timestamp)
	this.PutEnd()
	return this.GetBytes()
}
func (this *HeartBeat) Decode() {
	this.Timestamp = this.GetUint32()
}
