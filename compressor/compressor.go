package compressor

import (
	"io"
)

type Compressor interface {
	Init() io.Reader
}