package utils

import (
	"dicom-store-api/models"
	"fmt"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"reflect"
	"regexp"
	"strconv"
)

func ExtractDicomObjectFromDataset(dataset dicom.Dataset, object models.DicomObject) {
	reflection := reflect.TypeOf(object).Elem()

	for i := 0; i < reflection.NumField(); i++ {
		field := reflection.Field(i)
		tagInfo, err := tag.FindByName(field.Tag.Get("dicom"))
		if err != nil {
			continue
		}
		element, _ := dataset.FindElementByTag(tagInfo.Tag)
		if element == nil {
			continue
		}
		if element.Value.ValueType() != 0 {
			panic(fmt.Sprintf("field %s is not a string type", field.Name))
		}

		var stringDatasetValue string
		if tagInfo.VR == "SQ" {
			stringDatasetValue = element.Value.String()
		} else {
			stringDatasetValue = element.Value.GetValue().([]string)[0]
		}
		reflect.ValueOf(object).Elem().FieldByIndex(field.Index).SetString(stringDatasetValue)
	}
}

func GetTagByNameOrCode(tagName string) (tag.Tag, error) {
	isCode, _ := regexp.MatchString("^\\d{8}$", tagName)
	var tagInfo tag.Info
	var err error

	if isCode == true {
		var group int64
		group, err = strconv.ParseInt(tagName[0:4], 16, 0)
		if err != nil {
			return tag.Tag{}, fmt.Errorf("invalid tag name or code")
		}
		var elem int64
		elem, err = strconv.ParseInt(tagName[4:], 16, 0)
		if err != nil {
			return tag.Tag{}, fmt.Errorf("invalid tag name or code")
		}
		tagInfo, err = tag.Find(tag.Tag{Group: uint16(group), Element: uint16(elem)})
	} else {
		tagInfo, err = tag.FindByName(tagName)
	}

	if err != nil {
		return tag.Tag{}, fmt.Errorf("invalid tag name or code")
	}

	return tagInfo.Tag, err
}
