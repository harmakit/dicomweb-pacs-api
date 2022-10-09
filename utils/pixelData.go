package utils

import (
	"fmt"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"image"
	"image/color"
	"math"
	"strconv"
)

type ScalingFunction int

const (
	Linear ScalingFunction = iota
	LinearExact
	Sigmoid
)

type RenderImageWindowParameters struct {
	WindowCenter int
	WindowWidth  int
	Function     ScalingFunction
}

type RenderImageViewportParameters struct {
	ViewportWidth  int
	ViewportHeight int
	SliceX         int
	SliceY         int
	SliceWidth     int
	SliceHeight    int
}

func RenderImage(
	dataset dicom.Dataset,
	viewportParameters RenderImageViewportParameters,
	windowParameters RenderImageWindowParameters,
) (image.Image, error) {
	pixelDataElement, _ := dataset.FindElementByTag(tag.PixelData)
	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)

	if len(pixelDataInfo.Frames) == 0 {
		return nil, fmt.Errorf("no frames found")
	}
	frame := pixelDataInfo.Frames[0]

	bitsAllocated, _ := dataset.FindElementByTag(tag.BitsAllocated)
	bits := bitsAllocated.Value.GetValue().([]int)[0]

	bitsStored, _ := dataset.FindElementByTag(tag.BitsStored)
	bitsStoredValue := bitsStored.Value.GetValue().([]int)[0]

	highBit, _ := dataset.FindElementByTag(tag.HighBit)
	pixelRepresentation, _ := dataset.FindElementByTag(tag.PixelRepresentation)
	smallestPixelValue, _ := dataset.FindElementByTag(tag.SmallestImagePixelValue)
	largestPixelValue, _ := dataset.FindElementByTag(tag.LargestImagePixelValue)

	rowsElement, _ := dataset.FindElementByTag(tag.Rows)
	rows := rowsElement.Value.GetValue().([]int)[0]
	columnsElement, _ := dataset.FindElementByTag(tag.Columns)
	columns := columnsElement.Value.GetValue().([]int)[0]

	_ = highBit
	_ = pixelRepresentation
	_ = smallestPixelValue
	_ = largestPixelValue

	var pixelData []int

	if frame.Encapsulated {
		//bitsStoredValue -= 1
		bits = bitsStoredValue

		reader := ""
		for i := 0; i < len(frame.EncapsulatedData.Data); i++ {
			reader += fmt.Sprintf("%08b", frame.EncapsulatedData.Data[i])
			if len(reader) > bitsStoredValue {
				part := reader[:bitsStoredValue]
				reader = reader[bitsStoredValue:]
				value, _ := strconv.ParseInt(part, 2, 16)
				pixelData = append(pixelData, int(value))
			}
			//if len(pixelData) != 0 && len(pixelData)%2 == 0 {
			//	// swap last two elemets
			//	pixelData[len(pixelData)-1], pixelData[len(pixelData)-2] = pixelData[len(pixelData)-2], pixelData[len(pixelData)-1]
			//}
		}
	} else {
		rows = frame.NativeData.Rows
		columns = frame.NativeData.Cols

		for i := 0; i < len(frame.NativeData.Data); i++ {
			for j := 0; j < len(frame.NativeData.Data[i]); j++ {
				pixelData = append(pixelData, frame.NativeData.Data[i][j])
			}
		}
	}

	//canApplyWindowing := windowParameters.WindowCenter != 0 && windowParameters.WindowWidth != 0
	//if canApplyWindowing {
	//	pixelData = applyWindowing(pixelData, bits, windowParameters)
	//}

	pixelData = negatePixelData(pixelData, bits)

	if bits > 8 { // convert to uint8 if necessary
		for i := 0; i < len(pixelData); i++ {
			pixelData[i] = pixelData[i] >> (bits - 8)
		}
	}

	rows /= 2
	columns /= 2

	img := image.NewGray(image.Rect(0, 0, rows, columns))
	for j := 0; j < len(pixelData); j++ {
		img.SetGray(j%rows, j/columns, color.Gray{Y: uint8(pixelData[j])})
	}

	return img, nil
}

func negatePixelData(data []int, bits int) []int {
	max := (1 << uint(bits)) - 1
	for i := 0; i < len(data); i++ {
		data[i] = max - data[i]
	}
	return data
}

func applyWindowing(data []int, bits int, windowParameters RenderImageWindowParameters) []int {
	switch windowParameters.Function {
	case Linear:
		return applyLinearWindowing(data, bits, windowParameters.WindowWidth, windowParameters.WindowCenter)
	case LinearExact:
		return applyLinearExactWindowing(data, bits, windowParameters.WindowWidth, windowParameters.WindowCenter)
	case Sigmoid:
		return applySigmoidWindowing(data, bits, windowParameters.WindowWidth, windowParameters.WindowCenter)
	default:
		return data
	}
}

func applyLinearWindowing(data []int, bits int, width int, center int) []int {
	return data
}

func applyLinearExactWindowing(data []int, bits int, width int, center int) []int {
	return data
}

func applySigmoidWindowing(data []int, bits int, width int, center int) []int {
	yMax := (1 << uint(bits)) - 1
	yMin := 0

	for i := 0; i < len(data); i++ {
		value := int(float64(yMax-yMin)/(1+math.Exp(float64(-4*(data[i]-center))/float64(width))) + float64(yMin))
		data[i] = value
	}

	return data
}
