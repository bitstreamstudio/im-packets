package packets

import (
	"fmt"
	"io"
)

//PublishPacket is an internal representation of the fields of the
//Publish MQTT packet
type PublishPacket struct {
	FixedHeader
	Payload []byte
}

func (p *PublishPacket) String() string {
	return fmt.Sprintf("%s payload: %s", p.FixedHeader, string(p.Payload))
}

func (p *PublishPacket) Write(w io.Writer) error {
	var err error
	p.FixedHeader.RemainingLength = len(p.Payload)
	packet := p.FixedHeader.pack()
	packet.Write(p.Payload)
	_, err = w.Write(packet.Bytes())

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (p *PublishPacket) Unpack(b io.Reader) error {
	var payloadLength = p.FixedHeader.RemainingLength
	var err error
	if payloadLength < 0 {
		return fmt.Errorf("error unpacking publish, payload length < 0")
	}
	p.Payload = make([]byte, payloadLength)
	_, err = b.Read(p.Payload)

	return err
}

//Copy creates a new PublishPacket with the same topic and payload
//but an empty fixed header, useful for when you want to deliver
//a message with different properties such as Qos but the same
//content
func (p *PublishPacket) Copy() *PublishPacket {
	newP := NewControlPacket(Publish).(*PublishPacket)
	newP.Payload = p.Payload

	return newP
}
