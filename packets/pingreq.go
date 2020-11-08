package packets

import (
	"io"
)

//PingreqPacket is an internal representation of the fields of the
//Pingreq MQTT packet
type PingreqPacket struct {
	FixedHeader
}

func (pr *PingreqPacket) String() string {
	return pr.FixedHeader.String()
}

func (pr *PingreqPacket) Write(w io.Writer) error {
	packet := pr.FixedHeader.pack()
	_, err := packet.WriteTo(w)

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (pr *PingreqPacket) Unpack(b io.Reader) error {
	return nil
}