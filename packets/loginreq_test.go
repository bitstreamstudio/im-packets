package packets

import (
	"bytes"
	"fmt"
	"testing"
)

func TestLoginreqPacket(t *testing.T) {
	b := new(bytes.Buffer)
	packet := NewControlPacket(Loginreq).(*LoginreqPacket)
	packet.Format = 1
	packet.UserId = "test"
	packet.Token = "token"
	fmt.Println(b.String())
	err2 := packet.Write(b)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	readPacket, err := ReadPacket(b)
	if err != nil {
		fmt.Println(err)
		return
	}
	loginreqPacket := readPacket.(*LoginreqPacket)
	fmt.Println(loginreqPacket)
}
