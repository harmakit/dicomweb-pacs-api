package fs

import (
	"bytes"
	"io"
	"os"
)

func Save(filepath string, data []byte) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	r := bytes.NewReader(data)
	_, err = io.Copy(out, r)
	return err
}
