package packets

import (
	"encoding/json"
	"fmt"
	"github.com/bitstreamstudio/im-packets/protocol"
	"github.com/golang/protobuf/proto"
	"io"
)

//LoginreqPacket is an internal representation of the fields of the
//Loginreq TCP packet
type LoginreqPacket struct {
	FixedHeader
	protocol.LoginReq
}

func (lr *LoginreqPacket) String() string {
	return fmt.Sprintf("%s %s", lr.FixedHeader.String(), lr.LoginReq.String())
}

func (lr *LoginreqPacket) Write(w io.Writer) error {
	var err error
	var bytes []byte
	if lr.Format == FormatJson {
		bytes, err = json.Marshal(lr)
	} else {
		bytes, err = proto.Marshal(lr)
	}
	if err != nil {
		return err
	}
	lr.RemainingLength = uint32(len(bytes))
	packet := lr.FixedHeader.pack()
	packet.Write(bytes)
	_, err = packet.WriteTo(w)
	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (lr *LoginreqPacket) Unpack(b io.Reader) error {
	var payloadLength = lr.FixedHeader.RemainingLength
	if payloadLength < 0 {
		return fmt.Errorf("error unpacking publish, payload length < 0")
	}
	var err error
	bytes := make([]byte, payloadLength)
	_, err = b.Read(bytes)
	if err != nil {
		return err
	}
	if lr.Format == FormatJson {
		err = json.Unmarshal(bytes, lr)
	} else {
		err = proto.Unmarshal(bytes, lr)
	}
	return err
}
