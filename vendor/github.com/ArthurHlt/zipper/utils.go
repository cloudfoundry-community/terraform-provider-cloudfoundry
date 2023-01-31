package zipper

import (
	"bytes"
	"crypto/sha1"
	"debug/elf"
	"debug/macho"
	"encoding/binary"
	"encoding/hex"
	"io"
	"path/filepath"
	"strings"
)

const (
	chunkForSha1 = 5 * 1024
)

// check if path is a web url
func IsWebURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}



// Create sha1 from a reader by loading in maximum 5kb
func GetSha1FromReader(reader io.Reader) (string, error) {
	buf, err := Chunk(reader, chunkForSha1)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write(buf)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func Chunk(reader io.Reader, chunkSize int64) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := io.CopyN(buf, reader, chunkSize)
	if err != nil && err != io.EOF {
		return []byte{}, err
	}
	// we don't want to retrieve everything
	// so we close if it closeable
	if closer, ok := reader.(io.Closer); ok {
		err := closer.Close()
		if err != nil {
			return []byte{}, err
		}
	}
	return buf.Bytes(), nil
}

// check if file has one if extensions given
func HasExtFile(path string, extensions ...string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false
	}
	for _, extension := range extensions {
		if extension == ext {
			return true
		}
	}
	return false
}

// check if file is a zip file by extension
func IsZipFileExt(path string) bool {
	return HasExtFile(path, ZIP_FILE_EXT...)
}

// check if file is a zip file
func IsZipFile(reader io.Reader) bool {
	buf, err := Chunk(reader, 4)
	if err != nil {
		panic(err)
	}
	if len(buf) < 4 {
		return false
	}
	return buf[0] == 0x50 && buf[1] == 0x4b && (buf[2] == 0x03 || buf[2] == 0x05 || buf[2] == 0x07) && (buf[3] == 0x04 || buf[3] == 0x06 || buf[3] == 0x08)
}

// check if file is an executable file
func IsExecutable(reader io.Reader) bool {
	buf, err := Chunk(reader, 4)
	if err != nil {
		panic(err)
	}
	if len(buf) < 4 {
		return false
	}
	le := binary.LittleEndian.Uint32(buf)
	be := binary.BigEndian.Uint32(buf)

	return string(buf) == elf.ELFMAG || // elf
		string(buf[:2]) == "MZ" || // .exe windows
		string(buf[:2]) == "#!" || // shebang
		macho.Magic32 == le || macho.Magic32 == be || macho.Magic64 == le || macho.Magic64 == be || macho.MagicFat == le || macho.MagicFat == be // mach-o
}

// check if file is a tar file
func IsTarFile(path string) bool {
	return HasExtFile(path, TAR_FILE_EXT...)
}

// check if file is a tgz file
func IsTarGzFile(path string) bool {
	isTgz := HasExtFile(path, TARGZ_FILE_EXT...)
	if isTgz {
		return true
	}
	isGz := HasExtFile(path, GZIP_FILE_EXT...)
	if !isGz {
		return false
	}
	return IsTarFile(filepath.Ext(strings.TrimSuffix(path, filepath.Ext(path))))
}
