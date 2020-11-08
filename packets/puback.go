package packets

import (
	"fmt"
	"io"
)

//PubackPacket is an internal representation of the fields of the
//Puback MQTT packet
type PubackPacket struct {
	FixedHeader
	MessageID  uint16
	ReturnCode byte
}

func (pa *PubackPacket) String() string {
	return fmt.Sprintf("%s MessageID: %d ReturnCode: %d", pa.FixedHeader, pa.MessageID, pa.ReturnCode)
}

func (pa *PubackPacket) Write(w io.Writer) error {
	var err error
	pa.FixedHeader.RemainingLength = 3
	packet := pa.FixedHeader.pack()
	packet.Write(encodeUint16(pa.MessageID))
	packet.WriteByte(pa.ReturnCode)
	_, err = packet.WriteTo(w)

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (pa *PubackPacket) Unpack(b io.Reader) error {
	var err error
	pa.MessageID, err = decodeUint16(b)
	pa.ReturnCode, err = decodeByte(b)
	return err
}
