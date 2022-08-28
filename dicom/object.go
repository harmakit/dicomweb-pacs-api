package dicom

import "github.com/suyashkumar/dicom/pkg/tag"

type Object struct {
	objectIdFieldTag tag.Tag
}

func (object Object) GetObjectIdFieldTag() tag.Tag {
	return object.objectIdFieldTag
}
