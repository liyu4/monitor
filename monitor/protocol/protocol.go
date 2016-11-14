package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	ConstHeader         = "DMP"
	ConstHeaderLength   = 3
	ConstSaveDataLength = 4
)

func Packet(message []byte) []byte {
	a := append(append([]byte(ConstHeader), IntToBytes(len(message))...), message...)
	return a
}

func Unpack(buffer []byte, readerChannel chan []byte) []byte {
	length := len(buffer)
	var i int
	for i = 0; i < length; i = i + 1 {
		// NO data.
		if length < i+ConstHeaderLength+ConstSaveDataLength {
			break
		}
		if string(buffer[i:i+ConstHeaderLength]) == ConstHeader {
			//Get data length.
			messageLength := BytesToInt(buffer[i+ConstHeaderLength : i+ConstHeaderLength+ConstSaveDataLength])
			if length < i+ConstHeaderLength+ConstSaveDataLength+messageLength {
				break
			}
			data := buffer[i+ConstHeaderLength+ConstSaveDataLength : i+ConstHeaderLength+ConstSaveDataLength+messageLength]
			readerChannel <- data
			i += ConstHeaderLength + ConstSaveDataLength + messageLength - 1
		}
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]
}

func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}
