package models

import (
	"github.com/suyashkumar/dicom/pkg/tag"
	"time"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/go-pg/pg/orm"
)

type Series struct {
	TableName struct{} `sql:"series"`

	ID        int       `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Study     *Study    `json:"study"`

	Modality                        string `json:"modality"`
	SeriesInstanceUID               string `json:"series_instance_uid"`
	SeriesNumber                    string `json:"series_number"`
	PerformedProcedureStepStartDate string `json:"performed_procedure_step_start_date"`
	PerformedProcedureStepStartTime string `json:"performed_procedure_step_start_time"`
	RequestAttributesSequence       string `json:"request_attributes_sequence"`
	ScheduledProcedureStepID        string `json:"scheduled_procedure_step_id"`
	RequestedProcedureID            string `json:"requested_procedure_id"`
}

func (s Series) GetObjectIdFieldTag() tag.Tag {
	return tag.SeriesInstanceUID
}

// BeforeInsert hook executed before database insert operation.
func (s *Series) BeforeInsert(db orm.DB) error {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	return nil
}

// BeforeUpdate hook executed before database update operation.
func (s *Series) BeforeUpdate(db orm.DB) error {
	s.UpdatedAt = time.Now()
	return s.Validate()
}

// Validate validates Series struct and returns validation errors.
func (s *Series) Validate() error {
	return validation.ValidateStruct(s)
}
