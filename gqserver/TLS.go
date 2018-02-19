package gqserver

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
)

// Contains every field in a ClientHello message
type ClientHello struct {
	handshakeType         byte
	length                int
	clientVersion         []byte
	random                []byte
	sessionIdLen          int
	sessionId             []byte
	cipherSuitesLen       int
	cipherSuites          []byte
	compressionMethodsLen int
	compressionMethods    []byte
	extensionsLen         int
	extensions            map[[2]byte][]byte
}

func parseExtensions(input []byte) (ret map[[2]byte][]byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			err = errors.New("Malformed Extensions")
		}
	}()
	pointer := 0
	total_len := len(input)
	ret = make(map[[2]byte][]byte)
	for pointer < total_len {
		var typ [2]byte
		copy(typ[:], input[pointer:pointer+2])
		pointer += 2
		length := BtoInt(input[pointer : pointer+2])
		pointer += 2
		data := input[pointer : pointer+length]
		pointer += length
		ret[typ] = data
	}
	return ret, err
}

func peelRecordLayer(data []byte) (ret []byte, err error) {
	ret = data[5:]
	return
}

// ParseClientHello parses everything on top of the TLS layer
// (including the record layer) into ClientHello type
func ParseClientHello(data []byte) (ret *ClientHello, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Malformed ClientHello")
		}
	}()
	data, err = peelRecordLayer(data)
	pointer := 0
	// Handshake Type
	handshakeType := data[pointer]
	if handshakeType != 0x01 {
		return ret, errors.New("Not a ClientHello")
	}
	pointer += 1
	// Length
	length := BtoInt(data[pointer : pointer+3])
	pointer += 3
	if length != len(data[pointer:]) {
		return ret, errors.New("Hello length doesn't match")
	}
	// Client Version
	clientVersion := data[pointer : pointer+2]
	pointer += 2
	// Random
	random := data[pointer : pointer+32]
	pointer += 32
	// Session ID
	sessionIdLen := int(data[pointer])
	pointer += 1
	sessionId := data[pointer : pointer+sessionIdLen]
	pointer += sessionIdLen
	// Cipher Suites
	cipherSuitesLen := BtoInt(data[pointer : pointer+2])
	pointer += 2
	cipherSuites := data[pointer : pointer+cipherSuitesLen]
	pointer += cipherSuitesLen
	// Compression Methods
	compressionMethodsLen := int(data[pointer])
	pointer += 1
	compressionMethods := data[pointer : pointer+compressionMethodsLen]
	pointer += compressionMethodsLen
	// Extensions
	extensionsLen := BtoInt(data[pointer : pointer+2])
	pointer += 2
	extensions, err := parseExtensions(data[pointer:])
	ret = &ClientHello{
		handshakeType,
		length,
		clientVersion,
		random,
		sessionIdLen,
		sessionId,
		cipherSuitesLen,
		cipherSuites,
		compressionMethodsLen,
		compressionMethods,
		extensionsLen,
		extensions,
	}
	return
}

func addRecordLayer(input []byte, typ []byte) []byte {
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(input)))
	ret := append(typ, []byte{0x03, 0x03}...)
	ret = append(ret, length...)
	ret = append(ret, input...)
	return ret
}

func composeServerHello(client_hello *ClientHello) []byte {
	var serverHello [10][]byte
	serverHello[0] = []byte{0x02}             // handshake type
	serverHello[1] = []byte{0x00, 0x00, 0x4d} // length 77
	serverHello[2] = []byte{0x03, 0x03}       // server version
	random := make([]byte, 32)
	binary.BigEndian.PutUint32(random, rand.Uint32())
	serverHello[3] = random                               // random
	serverHello[4] = []byte{0x20}                         // session id length 32
	serverHello[5] = client_hello.sessionId               // session id
	serverHello[6] = []byte{0xc0, 0x30}                   // cipher suite TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	serverHello[7] = []byte{0x00}                         // compression method null
	serverHello[8] = []byte{0x00, 0x05}                   // extensions length 5
	serverHello[9] = []byte{0xff, 0x01, 0x00, 0x01, 0x00} // extensions renegotiation_info
	ret := []byte{}
	for i := 0; i < 10; i++ {
		ret = append(ret, serverHello[i]...)
	}
	return ret
}

func ComposeReply(client_hello *ClientHello) []byte {
	sh_bytes := addRecordLayer(composeServerHello(client_hello), []byte{0x16})
	ccs_bytes := addRecordLayer([]byte{0x01}, []byte{0x14})
	finished := make([]byte, 64)
	r := rand.Uint64()
	binary.BigEndian.PutUint64(finished, r)
	finished = finished[0:40]
	f_bytes := addRecordLayer(finished, []byte{0x16})
	ret := append(sh_bytes, ccs_bytes...)
	ret = append(ret, f_bytes...)
	return ret
}