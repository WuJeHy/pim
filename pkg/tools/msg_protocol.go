package tools

import "errors"

type ProtocolPackage struct {
	Version  byte
	ModType  int16
	SubType  int16
	UniqueId int64
	Body     []byte
}

func NewProtocolPackage(modType, subType int16, uniqueID int64, body []byte) *ProtocolPackage {
	return &ProtocolPackage{
		Version:  1,
		ModType:  modType,
		SubType:  subType,
		UniqueId: uniqueID,
		Body:     body,
	}
}

func ProcProtocolPackage(data []byte) (p *ProtocolPackage, err error) {

	if len(data) < 13 {
		err = errors.New("msg len error")
		return
	}

	p = new(ProtocolPackage)

	p.Version = data[0]

	p.ModType = int16(data[1])<<8 | int16(data[2])
	p.SubType = int16(data[3])<<8 | int16(data[4])

	p.UniqueId = int64(data[5])<<56 |
		int64(data[6])<<48 |
		int64(data[7])<<40 |
		int64(data[8])<<32 |
		int64(data[9])<<24 |
		int64(data[10])<<16 |
		int64(data[11])<<8 |
		int64(data[12])

	p.Body = data[13:]

	return

}

func (p *ProtocolPackage) Bytes() (data []byte) {

	data = make([]byte, len(p.Body)+13)
	data[0] = p.Version
	data[1] = byte(p.ModType >> 8)
	data[2] = byte(p.ModType)
	data[3] = byte(p.SubType >> 8)
	data[4] = byte(p.SubType)
	data[5] = byte(p.UniqueId >> 56)
	data[6] = byte(p.UniqueId >> 48)
	data[7] = byte(p.UniqueId >> 40)
	data[8] = byte(p.UniqueId >> 32)
	data[9] = byte(p.UniqueId >> 24)
	data[10] = byte(p.UniqueId >> 16)
	data[11] = byte(p.UniqueId >> 8)
	data[12] = byte(p.UniqueId)

	copy(data[13:], p.Body)

	return data

}
