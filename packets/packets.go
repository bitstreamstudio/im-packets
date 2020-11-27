package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bitstreamstudio/im-packets/protocol"
	"io"
)

//ControlPacket defines the interface for structs intended to hold
//decoded MQTT packets, either from being read or before being
//written
type ControlPacket interface {
	Write(io.Writer) error
	Unpack(io.Reader) error
	String() string
}

//PacketNames maps the constants for each of the MQTT packet types
//to a string representation of their name.
var PacketNames = map[uint8]string{
	1: "PINGREQ",
	2: "PINGRESP",
	3: "DISCONNECT",
	4: "LOGINREQ",
}

//Below are the constants assigned to each of the MQTT packet types
const (
	Pingreq    = 1
	Pingresp   = 2
	Disconnect = 3
	Loginreq   = 4
)

//Below are the const definitions for error codes returned by
//Connect()
const (
	Accepted                        = 0x00
	ErrRefusedBadProtocolVersion    = 0x01
	ErrRefusedIDRejected            = 0x02
	ErrRefusedServerUnavailable     = 0x03
	ErrRefusedBadUsernameOrPassword = 0x04
	ErrRefusedNotAuthorised         = 0x05
	ErrNetworkError                 = 0xFE
	ErrProtocolViolation            = 0xFF
)

const (
	FormatProto   = 0
	FormatJson    = 1
	FormatDefault = FormatProto
)

//ConnackReturnCodes is a map of the error codes constants for Connect()
//to a string representation of the error
var ConnackReturnCodes = map[uint8]string{
	0:   "Connection Accepted",
	1:   "Connection Refused: Bad Protocol Version",
	2:   "Connection Refused: Client Identifier Rejected",
	3:   "Connection Refused: Server Unavailable",
	4:   "Connection Refused: Username or Password in unknown format",
	5:   "Connection Refused: Not Authorised",
	254: "Connection Error",
	255: "Connection Refused: Protocol Violation",
}

//ConnErrors is a map of the errors codes constants for Connect()
//to a Go error
var ConnErrors = map[byte]error{
	Accepted:                        nil,
	ErrRefusedBadProtocolVersion:    errors.New("unnacceptable protocol version"),
	ErrRefusedIDRejected:            errors.New("identifier rejected"),
	ErrRefusedServerUnavailable:     errors.New("server Unavailable"),
	ErrRefusedBadUsernameOrPassword: errors.New("bad user name or password"),
	ErrRefusedNotAuthorised:         errors.New("not Authorized"),
	ErrNetworkError:                 errors.New("network Error"),
	ErrProtocolViolation:            errors.New("protocol Violation"),
}

//ReadPacket takes an instance of an io.Reader (such as net.Conn) and attempts
//to read an MQTT packet from the stream. It returns a ControlPacket
//representing the decoded MQTT packet and an error. One of these returns will
//always be nil, a nil ControlPacket indicating an error occurred.
func ReadPacket(r io.Reader) (ControlPacket, error) {
	var fh FixedHeader
	b := make([]byte, 1)

	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}

	err = fh.unpack(b[0], r)
	if err != nil {
		return nil, err
	}

	cp, err := NewControlPacketWithHeader(fh)
	if err != nil {
		return nil, err
	}

	packetBytes := make([]byte, fh.RemainingLength)
	n, err := io.ReadFull(r, packetBytes)
	if err != nil {
		return nil, err
	}
	if int64(n) != int64(fh.RemainingLength) {
		return nil, errors.New("failed to read expected data")
	}

	err = cp.Unpack(bytes.NewBuffer(packetBytes))
	return cp, err
}

//NewControlPacket is used to create a new ControlPacket of the type specified
//by packetType, this is usually done by reference to the packet type constants
//defined in packets.go. The newly created ControlPacket is empty and a pointer
//is returned.
func NewControlPacket(packetType byte) ControlPacket {
	switch packetType {
	case Disconnect:
		return &DisconnectPacket{FixedHeader: FixedHeader{MessageType: Disconnect}}
	case Pingreq:
		return &PingreqPacket{FixedHeader: FixedHeader{MessageType: Pingreq}}
	case Pingresp:
		return &PingrespPacket{FixedHeader: FixedHeader{MessageType: Pingresp}}
	case Loginreq:
		return &LoginreqPacket{FixedHeader: FixedHeader{MessageType: Loginreq}, LoginReq: protocol.LoginReq{}}
	}
	return nil
}

//NewControlPacketWithHeader is used to create a new ControlPacket of the type
//specified within the FixedHeader that is passed to the function.
//The newly created ControlPacket is empty and a pointer is returned.
func NewControlPacketWithHeader(fh FixedHeader) (ControlPacket, error) {
	switch fh.MessageType {
	case Disconnect:
		return &DisconnectPacket{FixedHeader: fh}, nil
	case Pingreq:
		return &PingreqPacket{FixedHeader: fh}, nil
	case Pingresp:
		return &PingrespPacket{FixedHeader: fh}, nil
	case Loginreq:
		return &LoginreqPacket{FixedHeader: fh, LoginReq: protocol.LoginReq{}}, nil
	}

	return nil, fmt.Errorf("unsupported packet type 0x%x", fh.MessageType)
}

//FixedHeader is a struct to hold the decoded information from
//the fixed header of an MQTT ControlPacket
type FixedHeader struct {
	MessageType     byte   `json:"-"`
	MsqSeq          uint32 `json:"-"`
	Version         byte   `json:"-"`
	Format          byte   `json:"-"`
	Flag            byte   `json:"-"`
	RemainingLength uint32 `json:"-"`
}

func (fh FixedHeader) String() string {
	return fmt.Sprintf("%s: msgSeq:%d version:%d format:%d flag:%d rLength:%d", PacketNames[fh.MessageType], fh.MsqSeq, fh.Version, fh.Format, fh.Flag, fh.RemainingLength)
}

func boolToByte(b bool) byte {
	switch b {
	case true:
		return 1
	default:
		return 0
	}
}

func (fh *FixedHeader) pack() bytes.Buffer {
	var header bytes.Buffer
	header.WriteByte(fh.MessageType)
	header.Write(encodeUint32(fh.MsqSeq))
	header.WriteByte(fh.Version)
	header.WriteByte(fh.Format)
	header.WriteByte(fh.Flag)
	header.Write(encodeUint32(fh.RemainingLength))
	return header
}

func (fh *FixedHeader) unpack(typeAndFlags byte, r io.Reader) error {
	fh.MessageType = typeAndFlags
	var err error
	fh.MsqSeq, err = decodeUint32(r)
	if err != nil {
		return err
	}
	fh.Version, err = decodeByte(r)
	if err != nil {
		return err
	}
	fh.Format, err = decodeByte(r)
	if err != nil {
		return err
	}
	fh.Flag, err = decodeByte(r)
	if err != nil {
		return err
	}
	fh.RemainingLength, err = decodeUint32(r)
	return err
}

func decodeByte(b io.Reader) (byte, error) {
	num := make([]byte, 1)
	_, err := b.Read(num)
	if err != nil {
		return 0, err
	}

	return num[0], nil
}

func decodeUint16(b io.Reader) (uint16, error) {
	num := make([]byte, 2)
	_, err := b.Read(num)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(num), nil
}

func decodeUint32(b io.Reader) (uint32, error) {
	num := make([]byte, 4)
	_, err := b.Read(num)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(num), nil
}

func encodeUint16(num uint16) []byte {
	bytesResult := make([]byte, 2)
	binary.BigEndian.PutUint16(bytesResult, num)
	return bytesResult
}

func encodeUint32(num uint32) []byte {
	bytesResult := make([]byte, 4)
	binary.BigEndian.PutUint32(bytesResult, num)
	return bytesResult
}

func encodeString(field string) []byte {
	return encodeBytes([]byte(field))
}

func decodeString(b io.Reader) (string, error) {
	buf, err := decodeBytes(b)
	return string(buf), err
}

func decodeBytes(b io.Reader) ([]byte, error) {
	fieldLength, err := decodeUint16(b)
	if err != nil {
		return nil, err
	}

	field := make([]byte, fieldLength)
	_, err = b.Read(field)
	if err != nil {
		return nil, err
	}

	return field, nil
}

func encodeBytes(field []byte) []byte {
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, field...)
}
