package fs

import (
	"bytes"
	"crypto/sha1"
	"dicom-store-api/models"
	"encoding/hex"
	"github.com/suyashkumar/dicom/pkg/tag"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const ROOT = "./"
const UPLOADS_DIR = "uploads"
const DICOM_PREFIX = "dicom"
const DICOM_EXT = ".dcm"

func Save(filepath string, data []byte) error {

	dirpath := filepath[:len(filepath)-len(filepath[strings.LastIndex(filepath, string(os.PathSeparator))+1:])]
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		err := os.MkdirAll(dirpath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	r := bytes.NewReader(data)
	_, err = io.Copy(out, r)
	return err
}

func GetDicomPath(study *models.Study, series *models.Series, instance *models.Instance) string {
	studyId := getDicomObjectPathString(study)
	seriesId := getDicomObjectPathString(series)
	instanceId := getDicomObjectPathString(instance)

	return ROOT + filepath.Join(UPLOADS_DIR, DICOM_PREFIX, studyId, seriesId, instanceId+DICOM_EXT)
}

func getDicomObjectPathString(object models.DicomObject) string {
	tagInfo, _ := tag.Find(object.GetObjectIdFieldTag())
	id := reflect.ValueOf(object).Elem().FieldByName(tagInfo.Name).String()

	hash := sha1.New()
	hash.Write([]byte(id))

	return hex.EncodeToString(hash.Sum(nil))
}
