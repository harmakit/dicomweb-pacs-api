package models

import (
	"github.com/suyashkumar/dicom/pkg/tag"
	"time"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/go-pg/pg/orm"
)

type Study struct {
	TableName struct{} `sql:"study"`

	ID        int       `json:"-" sql:",pk"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	StudyDate                     string `json:"study_date" dicom:"StudyDate"`
	StudyTime                     string `json:"study_time" dicom:"StudyTime"`
	AccessionNumber               string `json:"accession_number" dicom:"AccessionNumber"`
	InstanceAvailability          string `json:"instance_availability" dicom:"InstanceAvailability"`
	ModalitiesInStudy             string `json:"modalities_in_study" dicom:"ModalitiesInStudy"`
	ReferringPhysicianName        string `json:"referring_physician_name" dicom:"ReferringPhysicianName"`
	TimeZoneOffsetFromUTC         string `json:"time_zone_offset_from_utc" dicom:"TimeZoneOffsetFromUTC"`
	RetrieveURL                   string `json:"retrieve_url" dicom:"RetrieveURL"`
	PatientName                   string `json:"patient_name" dicom:"PatientName"`
	PatientID                     string `json:"patient_id" dicom:"PatientID"`
	PatientBirthDate              string `json:"patient_birth_date" dicom:"PatientBirthDate"`
	PatientSex                    string `json:"patient_sex" dicom:"PatientSex"`
	StudyInstanceUID              string `json:"study_instance_uid" sql:",unique" dicom:"StudyInstanceUID"`
	StudyID                       string `json:"study_id" dicom:"StudyID"`
	NumberOfStudyRelatedSeries    string `json:"number_of_study_related_series" dicom:"NumberOfStudyRelatedSeries"`
	NumberOfStudyRelatedInstances string `json:"number_of_study_related_instances" dicom:"NumberOfStudyRelatedInstances"`
}

func (s *Study) GetObjectIdFieldTag() tag.Tag {
	return tag.StudyInstanceUID
}

// BeforeInsert hook executed before database insert operation.
func (s *Study) BeforeInsert(db orm.DB) error {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	return nil
}

// BeforeUpdate hook executed before database update operation.
func (s *Study) BeforeUpdate(db orm.DB) error {
	s.UpdatedAt = time.Now()
	return s.Validate()
}

// Validate validates Study struct and returns validation errors.
func (s *Study) Validate() error {
	return validation.ValidateStruct(s) //validation.Field(&p.Patient, validation.Required, validation.In("patient1", "patient2")),
}
