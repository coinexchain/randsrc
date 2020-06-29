package randsrc

import (
	"fmt"
	"encoding/binary"
	"hash"
	"math"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

type RandBytesSrcFromFile struct {
	fname   string
	file    *os.File
	h       hash.Hash
	buf     []byte
	idx     int
}

func NewRandBytesSrcFromFileWithSeed(fname string, seed []byte) RandBytesSrcFromFile {
	rs := RandBytesSrcFromFile{}
	rs.fname = fname
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	//n, err := file.Seek(0, os.SEEK_CUR)
	//if err != nil {
	//	panic(err)
	//}
	rs.file = file
	rs.h, _ = blake2b.New512(seed)
	rs.step()
	return rs
}

func NewRandBytesSrcFromFile(fname string) RandBytesSrcFromFile {
	return NewRandBytesSrcFromFileWithSeed(fname, nil)
}

func (rs *RandBytesSrcFromFile) Close() {
	rs.file.Close()
}

func (rs *RandBytesSrcFromFile) new512bits() []byte {
	var buf [32]byte
	_, err := rs.file.Read(buf[:])
	if err == io.EOF {
		rs.file.Seek(0,0)
		_, err = rs.file.Read(buf[:])
	}
	if err != nil {
		panic(err)
	}
	rs.h.Write(buf[:])
	res := rs.h.Sum(nil)
	if len(res)!=64 {
		panic(fmt.Sprintf("Not 512bits: ", len(res)))
	}
	return res
}

func (rs *RandBytesSrcFromFile) step() {
	var arrA, arrB [16][]byte
	for i := 0; i < 16; i++ {
		arrA[i] = rs.new512bits()
		arrB[i] = rs.new512bits()
	}
	rs.buf = rs.buf[:0]
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			var buf [64]byte
			copy(buf[:], arrA[i])
			for k := 0; k < len(buf); k++ {
				buf[k] ^= arrB[j][k]
			}
			//fmt.Printf("haha %v a%d %v b%d %v\n", buf[:], i, arrA[i], j, arrB[j])
			rs.buf = append(rs.buf, buf[:]...)
		}
	}
	rs.idx = 0
}

func (rs *RandBytesSrcFromFile) GetBytes(n int) []byte {
	res := make([]byte, 0, n)
	for len(res) < n {
		res = append(res, rs.buf[rs.idx])
		rs.idx++
		if rs.idx == len(rs.buf) {
			rs.step()
		}
	}
	return res
}

var chars = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (rs *RandBytesSrcFromFile) GetString(n int) string {
	res := make([]byte, n)
	for i, c := range rs.GetBytes(n) {
		j := int(c)%len(chars)
		res[i] = byte(chars[j])
	}
	return string(res)
}

type RandSrcFromFile struct {
	RandBytesSrcFromFile
}

func NewRandSrcFromFile(fname string) *RandSrcFromFile {
	return NewRandSrcFromFileWithSeed(fname, nil)
}

func NewRandSrcFromFileWithSeed(fname string, seed []byte) *RandSrcFromFile {
	var res RandSrcFromFile
	res.RandBytesSrcFromFile = NewRandBytesSrcFromFileWithSeed(fname, seed)
	return &res
}

func (rs *RandSrcFromFile) GetBool() bool {
	bz := rs.GetBytes(1)
	return bz[0] != 0
}

func (rs *RandSrcFromFile) GetUint8() uint8 {
	bz := rs.GetBytes(1)
	return bz[0]
}

func (rs *RandSrcFromFile) GetUint16() uint16 {
	return binary.LittleEndian.Uint16(rs.GetBytes(2))
}

func (rs *RandSrcFromFile) GetUint32() uint32 {
	return binary.LittleEndian.Uint32(rs.GetBytes(4))
}

func (rs *RandSrcFromFile) GetUint64() uint64 {
	return binary.LittleEndian.Uint64(rs.GetBytes(8))
}

func (rs *RandSrcFromFile) GetInt64() int64 {
	return int64(rs.GetUint64())
}
func (rs *RandSrcFromFile) GetInt32() int32 {
	return int32(rs.GetUint32())
}
func (rs *RandSrcFromFile) GetInt16() int16 {
	return int16(rs.GetUint16())
}
func (rs *RandSrcFromFile) GetInt8() int8 {
	return int8(rs.GetUint8())
}

func (rs *RandSrcFromFile) GetInt() int {
	return int(rs.GetUint64())
}
func (rs *RandSrcFromFile) GetUint() uint {
	return uint(rs.GetUint64())
}

func (rs *RandSrcFromFile) GetFloat64() float64 {
	return math.Float64frombits(rs.GetUint64())
}
func (rs *RandSrcFromFile) GetFloat32() float32 {
	return math.Float32frombits(rs.GetUint32())
}

type RandSrc interface {
	GetBool() bool
	GetInt8() int8
	GetInt16() int16
	GetInt32() int32
	GetInt64() int64
	GetUint8() uint8
	GetUint16() uint16
	GetUint32() uint32
	GetUint64() uint64
	GetFloat32() float32
	GetFloat64() float64
	GetString(n int) string
	GetBytes(n int) []byte
}

var _ RandSrc = &RandSrcFromFile{}

/*
package main

import (
	"fmt"
	"github.com/coinexchain/randsrc"
)

func main() {
	rs := randsrc.NewRandSrcFromFile("a.dat")
	for i := 0; i >=0; i++ {
		if i % 1000 == 0 {
			fmt.Printf("Here %d\n", i)
		}
		rs.GetString(32)
	}
}
*/
