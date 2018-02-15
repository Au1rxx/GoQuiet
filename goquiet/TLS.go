package goquiet

import (
	"errors"
	//"fmt"
)

type Extension struct {
	typ    []byte
	length int
	data   []byte
}

type ClientHello struct {
	handshake_type          byte
	length                  int
	client_version          []byte
	random                  []byte
	session_id_len          int
	session_id              []byte
	cipher_suites_len       int
	cipher_suites           []byte
	compression_methods_len int
	compression_methods     []byte
	extensions_len          int
	extensions              []Extension
}

func parseExtensions(input []byte, total_len int) (ret []Extension, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Malformed Extensions")
		}
	}()
	pointer := 0
	for pointer < total_len {
		typ := input[pointer : pointer+2]
		pointer += 2
		length := BtoInt(input[pointer : pointer+2])
		pointer += 2
		data := input[pointer : pointer+length]
		pointer += length
		extension := Extension{
			typ,
			length,
			data,
		}
		ret = append(ret, extension)
	}
	return ret, err
}

func ParseClientHello(data []byte) (ret ClientHello, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Malformed ClientHello")
		}
	}()
	pointer := 0
	// Handshake Type
	handshake_type := data[pointer]
	if handshake_type != 0x01 {
		return ClientHello{}, errors.New("Not a ClientHello")
	}
	pointer += 1
	// Length
	length := BtoInt(data[pointer : pointer+3])
	pointer += 3
	if length != len(data[pointer:]) {
		return ClientHello{}, errors.New("Hello length doesn't match")
	}
	// Client Version
	client_version := data[pointer : pointer+2]
	pointer += 2
	// Random
	random := data[pointer : pointer+32]
	pointer += 32
	// Session ID
	session_id_len := int(data[pointer])
	pointer += 1
	session_id := data[pointer : pointer+session_id_len]
	pointer += session_id_len
	// Cipher Suites
	cipher_suites_len := BtoInt(data[pointer : pointer+2])
	pointer += 2
	cipher_suites := data[pointer : pointer+cipher_suites_len]
	pointer += cipher_suites_len
	// Compression Methods
	compression_methods_len := int(data[pointer])
	pointer += 1
	compression_methods := data[pointer : pointer+compression_methods_len]
	pointer += compression_methods_len
	// Extensions
	extensions_len := BtoInt(data[pointer : pointer+2])
	pointer += 2
	extensions, err := parseExtensions(data[pointer:], extensions_len)
	ret = ClientHello{
		handshake_type,
		length,
		client_version,
		random,
		session_id_len,
		session_id,
		cipher_suites_len,
		cipher_suites,
		compression_methods_len,
		compression_methods,
		extensions_len,
		extensions,
	}
	return ret, err
}