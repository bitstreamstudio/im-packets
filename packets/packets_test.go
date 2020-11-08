package packets

import (
	"bytes"
	"testing"
)

func TestPacketNames(t *testing.T) {
	if PacketNames[1] != "PUBLISH" {
		t.Errorf("PacketNames[3] is %s, should be %s", PacketNames[1], "PUBLISH")
	}
	if PacketNames[2] != "PUBACK" {
		t.Errorf("PacketNames[4] is %s, should be %s", PacketNames[2], "PUBACK")
	}
	if PacketNames[3] != "PINGREQ" {
		t.Errorf("PacketNames[12] is %s, should be %s", PacketNames[3], "PINGREQ")
	}
	if PacketNames[4] != "PINGRESP" {
		t.Errorf("PacketNames[13] is %s, should be %s", PacketNames[4], "PINGRESP")
	}
	if PacketNames[5] != "DISCONNECT" {
		t.Errorf("PacketNames[14] is %s, should be %s", PacketNames[5], "DISCONNECT")
	}
}

func TestPacketConsts(t *testing.T) {
	if Publish != 1 {
		t.Errorf("Const for Publish is %d, should be %d", Publish, 1)
	}
	if Puback != 2 {
		t.Errorf("Const for Puback is %d, should be %d", Puback, 2)
	}
	if Pingreq != 3 {
		t.Errorf("Const for Pingreq is %d, should be %d", Pingreq, 3)
	}
	if Pingresp != 4 {
		t.Errorf("Const for Pingresp is %d, should be %d", Pingresp, 4)
	}
	if Disconnect != 5 {
		t.Errorf("Const for Disconnect is %d, should be %d", Disconnect, 5)
	}
}

func TestPackUnpackControlPackets(t *testing.T) {
	packets := []ControlPacket{
		NewControlPacket(Publish).(*PublishPacket),
		NewControlPacket(Puback).(*PubackPacket),
		NewControlPacket(Pingreq).(*PingreqPacket),
		NewControlPacket(Pingresp).(*PingrespPacket),
		NewControlPacket(Disconnect).(*DisconnectPacket),
	}
	buf := new(bytes.Buffer)
	for _, packet := range packets {
		buf.Reset()
		if err := packet.Write(buf); err != nil {
			t.Errorf("Write of %T returned error: %s", packet, err)
		}
		read, err := ReadPacket(buf)
		if err != nil {
			t.Errorf("Read of packed %T returned error: %s", packet, err)
		}
		if read.String() != packet.String() {
			t.Errorf("Read of packed %T did not equal original.\nExpected: %v\n     Got: %v", packet, packet, read)
		}
	}
}

func TestEncoding(t *testing.T) {
	if res, err := decodeByte(bytes.NewBuffer([]byte{0x56})); res != 0x56 || err != nil {
		t.Errorf("decodeByte([0x56]) did not return (0x56, nil) but (0x%X, %v)", res, err)
	}
	if res, err := decodeUint16(bytes.NewBuffer([]byte{0x56, 0x78})); res != 22136 || err != nil {
		t.Errorf("decodeUint16([0x5678]) did not return (22136, nil) but (%d, %v)", res, err)
	}
	if res := encodeUint16(22136); !bytes.Equal(res, []byte{0x56, 0x78}) {
		t.Errorf("encodeUint16(22136) did not return [0x5678] but [0x%X]", res)
	}

	strings := map[string][]byte{
		"foo":         {0x00, 0x03, 'f', 'o', 'o'},
		"\U0000FEFF":  {0x00, 0x03, 0xEF, 0xBB, 0xBF},
		"A\U0002A6D4": {0x00, 0x05, 'A', 0xF0, 0xAA, 0x9B, 0x94},
	}
	for str, encoded := range strings {
		if res, err := decodeString(bytes.NewBuffer(encoded)); res != str || err != nil {
			t.Errorf("decodeString(%v) did not return (%q, nil), but (%q, %v)", encoded, str, res, err)
		}
		if res := encodeString(str); !bytes.Equal(res, encoded) {
			t.Errorf("encodeString(%q) did not return [0x%X], but [0x%X]", str, encoded, res)
		}
	}

	lengths := map[int][]byte{
		0:         {0x00},
		127:       {0x7F},
		128:       {0x80, 0x01},
		16383:     {0xFF, 0x7F},
		16384:     {0x80, 0x80, 0x01},
		2097151:   {0xFF, 0xFF, 0x7F},
		2097152:   {0x80, 0x80, 0x80, 0x01},
		268435455: {0xFF, 0xFF, 0xFF, 0x7F},
	}
	for length, encoded := range lengths {
		if res, err := decodeLength(bytes.NewBuffer(encoded)); res != length || err != nil {
			t.Errorf("decodeLength([0x%X]) did not return (%d, nil) but (%d, %v)", encoded, length, res, err)
		}
		if res := encodeLength(length); !bytes.Equal(res, encoded) {
			t.Errorf("encodeLength(%d) did not return [0x%X], but [0x%X]", length, encoded, res)
		}
	}
}
