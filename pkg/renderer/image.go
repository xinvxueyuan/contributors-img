package renderer

import "contrib.rocks/pkg/model"

type Image interface {
	model.FileHandle
	Bytes() []byte
}
