package models

import (
	"github.com/suyashkumar/dicom/pkg/tag"
	"time"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/go-pg/pg/orm"
)

type Instance struct {
	TableName struct{} `sql:"instance"`

	ID        int       `json:"-" sql:",pk"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SeriesId  int
	Series    *Series `json:"series"`

	SOPClassUID    string `json:"sop_class_uid" dicom:"SOPClassUID"`
	SOPInstanceUID string `json:"sop_instance_uid" dicom:"SOPInstanceUID"`
	InstanceNumber string `json:"instance_number" dicom:"InstanceNumber"`
}

func (i *Instance) GetObjectIdFieldTag() tag.Tag {
	return tag.SOPInstanceUID
}

// BeforeInsert hook executed before database insert operation.
func (i *Instance) BeforeInsert(db orm.DB) error {
	now := time.Now()
	i.CreatedAt = now
	i.UpdatedAt = now
	return nil
}

// BeforeUpdate hook executed before database update operation.
func (i *Instance) BeforeUpdate(db orm.DB) error {
	i.UpdatedAt = time.Now()
	return i.Validate()
}

// Validate validates Instance struct and returns validation errors.
func (i *Instance) Validate() error {
	return validation.ValidateStruct(i)
}
