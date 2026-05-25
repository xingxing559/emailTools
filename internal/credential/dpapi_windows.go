//go:build windows

package credential

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

func protect(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	err := windows.CryptProtectData(&in, nil, nil, 0, nil, 0, &out)
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	result := make([]byte, out.Size)
	copy(result, unsafe.Slice(out.Data, out.Size))
	return result, nil
}

func unprotect(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty cipher")
	}
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, 0, &out)
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	result := make([]byte, out.Size)
	copy(result, unsafe.Slice(out.Data, out.Size))
	return result, nil
}
