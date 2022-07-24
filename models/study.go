package models

import (
	"time"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/go-pg/pg/orm"
)

type Study struct {
	TableName struct{} `sql:"study"`

	ID        int       `json:"-"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	Patient string `json:"patient,omitempty"`
}

// BeforeInsert hook executed before database insert operation.
func (p *Study) BeforeInsert(db orm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook executed before database update operation.
func (p *Study) BeforeUpdate(db orm.DB) error {
	p.UpdatedAt = time.Now()
	return p.Validate()
}

// Validate validates Study struct and returns validation errors.
func (p *Study) Validate() error {

	return validation.ValidateStruct(p,
		validation.Field(&p.Patient, validation.Required, validation.In("patient1", "patient2")),
	)
}
