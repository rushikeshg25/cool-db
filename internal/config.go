package internal

import (
	"encoding/binary"
	"os"
	"sync"

	"github.com/rushikeshg25/coolDb/internal/errors"
)

const (
	DefaultPageSize           = 4096
	DefaultReservedSpace      = 0
	DefaultMaxEmbeddedPayload = 32
	DefaultMinEmbeddedPayload = 32
)

type CoolConfig struct {
	DbFile string
	Header DatabaseHeader
	F      *os.File
	Wg     *sync.WaitGroup
}

type DatabaseHeader struct {
	HeaderString        [16]byte
	PageSize            uint16
	ReservedSpace       uint8
	MaxEmbeddedPayload  uint8
	MinEmbeddedPayload  uint8
	FileChangeCounter   uint32
	DatabaseSize        uint32
	SchemaChangeCounter uint32
	SchemaFormatVersion uint8
	LargestRootBtree    uint32
	TextEncoding        uint8
	CoolDbVersion       uint32
}

func InitFileConfig() *DatabaseHeader {
	header := &DatabaseHeader{
		HeaderString:        [16]byte{},
		PageSize:            DefaultPageSize,
		ReservedSpace:       DefaultReservedSpace,
		MaxEmbeddedPayload:  DefaultMaxEmbeddedPayload,
		MinEmbeddedPayload:  DefaultMinEmbeddedPayload,
		FileChangeCounter:   0,
		DatabaseSize:        0,
		SchemaChangeCounter: 0,
		SchemaFormatVersion: 0,
		LargestRootBtree:    0,
		TextEncoding:        0,
		CoolDbVersion:       0,
	}
	copy(header.HeaderString[:], "cooldb format 1\000")
	return header
}

func ParseFileConfig(file string, filePath string, f *os.File) (*CoolConfig, error) {
	var wg sync.WaitGroup
	f.Seek(0, 0)
	header := &DatabaseHeader{}
	err := binary.Read(f, binary.LittleEndian, header)
	if err != nil {
		return nil, err
	}
	if string(header.HeaderString[:]) != "cooldb format 1\000" {
		return nil, errors.InvalidFileErr(filePath)
	}
	return &CoolConfig{
		DbFile: filePath,
		Header: *header,
		F:      f,
		Wg:     &wg,
	}, nil
}
