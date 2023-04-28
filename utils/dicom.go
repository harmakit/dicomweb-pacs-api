package utils

import (
	"dicom-store-api/models"
	"encoding/json"
	"fmt"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"reflect"
	"regexp"
	"strconv"
)

func ExtractDicomObjectFromDataset(dataset dicom.Dataset, object models.DicomObject) error {
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

		stringValue, err := GetStringValueFromElement(element)
		if err != nil {
			continue
		}
		reflect.ValueOf(object).Elem().FieldByIndex(field.Index).SetString(stringValue)
	}

	return nil
}

func GetStringValueFromElement(element *dicom.Element) (string, error) {
	value, err := getValueFromElement(element)
	if err != nil {
		return "", err
	}
	rt := reflect.TypeOf(value)
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		byteValue, _ := json.Marshal(value)
		return string(byteValue), nil
	}
	return fmt.Sprintf("%v", value), nil
}

type PatientNameValueStruct struct {
	Alphabetic  string `json:",omitempty"`
	Ideographic string `json:",omitempty"`
	Phonetic    string `json:",omitempty"`
}

func FormatStringValueForResponse(tagInfo tag.Info, value string) (any, error) {
	isSingleValue := tagInfo.VM == "1" && tagInfo.VR != "SQ"

	if tagInfo.VR == "PN" {
		return []any{&PatientNameValueStruct{
			Alphabetic: value,
		}}, nil
	}

	if isSingleValue {
		return []string{value}, nil
	}
	if value == "[]" || value == "" {
		return [0]string{}, nil
	}

	var s []any
	err := json.Unmarshal([]byte(value), &s)
	return s, err
}

func getValueFromElement(element *dicom.Element) (any, error) {
	tagInfo, err := tag.Find(element.Tag)
	if err != nil {
		return nil, err
	}

	isSingleValue := tagInfo.VM == "1"
	switch element.ValueRepresentation {
	case tag.VRStringList:
		if isSingleValue {
			return element.Value.GetValue().([]string)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRBytes:
		if isSingleValue {
			return element.Value.GetValue().([]byte)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRString:
		return element.Value.GetValue().([]string)[0], nil
	case tag.VRUInt16List:
		if isSingleValue {
			return element.Value.GetValue().([]int)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRUInt32List:
		if isSingleValue {
			return element.Value.GetValue().([]int)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRInt16List:
		if isSingleValue {
			return element.Value.GetValue().([]int16)[0], nil
		} else {
			return element.Value.GetValue().([]int16), nil
		}
	case tag.VRInt32List:
		if isSingleValue {
			return element.Value.GetValue().([]int32)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRFloat32List:
		if isSingleValue {
			return element.Value.GetValue().([]float32)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRFloat64List:
		if isSingleValue {
			return element.Value.GetValue().([]float64)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRSequence:
		var container [][]any
		for _, item := range element.Value.GetValue().([]*dicom.SequenceItemValue) {
			var itemContainer []any
			for _, itemElement := range item.GetValue().([]*dicom.Element) {
				itemElementValue, err := getValueFromElement(itemElement)
				if err != nil {
					continue
				}
				itemContainer = append(itemContainer, itemElementValue)
			}
			container = append(container, itemContainer)
		}
		if isSingleValue {
			if len(container) != 1 {
				return nil, fmt.Errorf("invalid sequence value")
			}
			return container[0], nil
		} else {
			return container, nil
		}
	case tag.VRItem:
		if isSingleValue {
			return element.Value.GetValue().([]*dicom.Element)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRTagList:
		if isSingleValue {
			return element.Value.GetValue().([]tag.Tag)[0], nil
		} else {
			return element.Value.GetValue(), nil
		}
	case tag.VRDate:
		return element.Value.GetValue().([]string)[0], nil
	case tag.VRPixelData:
		return element.Value.GetValue(), nil
	}

	return nil, fmt.Errorf("unknown value representation %s", element.ValueRepresentation)
}

func GetTagByNameOrCode(tagName string) (tag.Tag, error) {
	isCode, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, tagName)
	isCode = isCode && len(tagName) == 8
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
