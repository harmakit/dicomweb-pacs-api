package migrate

import (
	"fmt"

	"github.com/go-pg/migrations"
)

const studyTable = `
CREATE TABLE study (
id serial NOT NULL,
created_at timestamp with time zone NOT NULL DEFAULT current_timestamp,
updated_at timestamp with time zone NOT NULL DEFAULT current_timestamp,

study_date varchar(8),
study_time varchar(16),
accession_number varchar(16),
instance_availability varchar(16),
modalities_in_study text,
referring_physician_name varchar(255),
time_zone_offset_from_utc varchar(16),
retrieve_url varchar(2047),
patient_name varchar(255),
patient_id varchar(64),
patient_birth_date varchar(8),
patient_sex varchar(16),
study_instance_uid varchar(64) NOT NULL UNIQUE,
study_id varchar(64),
number_of_study_related_series varchar(12),
number_of_study_related_instances varchar(12),
    
PRIMARY KEY (id)
)`

const seriesTable = `
CREATE TABLE series (
id serial NOT NULL,
created_at timestamp with time zone NOT NULL DEFAULT current_timestamp,
updated_at timestamp with time zone NOT NULL DEFAULT current_timestamp,
study_id int NOT NULL REFERENCES study (id) ON DELETE CASCADE,

modality varchar(16),
series_instance_uid varchar(64) NOT NULL UNIQUE,
series_number varchar(15),
performed_procedure_step_start_date varchar(8),
performed_procedure_step_start_time varchar(16),
request_attributes_sequence text,
scheduled_procedure_step_id varchar(16),
requested_procedure_id varchar(16),
    
PRIMARY KEY (id)
)`

const instanceTable = `
CREATE TABLE instance (
id serial NOT NULL,
created_at timestamp with time zone NOT NULL DEFAULT current_timestamp,
updated_at timestamp with time zone NOT NULL DEFAULT current_timestamp,
series_id int NOT NULL REFERENCES series (id) ON DELETE CASCADE,

sop_class_uid varchar(64),
sop_instance_uid varchar(64) NOT NULL UNIQUE,
instance_number varchar(15),
    
PRIMARY KEY (id)
)`

func init() {
	up := []string{
		studyTable,
		seriesTable,
		instanceTable,
	}

	down := []string{
		`DROP TABLE study`,
		`DROP TABLE series`,
		`DROP TABLE instance`,
	}

	migrations.Register(func(db migrations.DB) error {
		fmt.Println("create tables")
		for _, q := range up {
			_, err := db.Exec(q)
			if err != nil {
				return err
			}
		}
		return nil
	}, func(db migrations.DB) error {
		fmt.Println("drop tables")
		for _, q := range down {
			_, err := db.Exec(q)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
