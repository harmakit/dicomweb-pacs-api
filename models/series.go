package models

import (
	"github.com/suyashkumar/dicom/pkg/tag"
	"time"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/go-pg/pg/orm"
)

type Series struct {
	TableName struct{} `sql:"series"`

	ID        int       `json:"-" sql:",pk"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	StudyId   int
	Study     *Study `json:"study"`

	Modality                        string `json:"modality" dicom:"Modality"`
	SeriesInstanceUID               string `json:"series_instance_uid" dicom:"SeriesInstanceUID"`
	SeriesNumber                    string `json:"series_number" dicom:"SeriesNumber"`
	PerformedProcedureStepStartDate string `json:"performed_procedure_step_start_date" dicom:"PerformedProcedureStepStartDate"`
	PerformedProcedureStepStartTime string `json:"performed_procedure_step_start_time" dicom:"PerformedProcedureStepStartTime"`
	RequestAttributesSequence       string `json:"request_attributes_sequence" dicom:"RequestAttributesSequence"`
	ScheduledProcedureStepID        string `json:"scheduled_procedure_step_id" dicom:"ScheduledProcedureStepID"`
	RequestedProcedureID            string `json:"requested_procedure_id" dicom:"RequestedProcedureID"`
}

func (s *Series) GetObjectIdFieldTag() tag.Tag {
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
