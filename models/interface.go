package models

import "github.com/suyashkumar/dicom/pkg/tag"

type DicomObject interface {
	GetObjectIdFieldTag() tag.Tag
}
