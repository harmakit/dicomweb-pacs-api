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

func ConvertDicomToImage(dataset dicom.Dataset) (image.Image, error) {
	pixelDataElement, _ := dataset.FindElementByTag(tag.PixelData)
	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)

	if len(pixelDataInfo.Frames) == 0 {
		return nil, fmt.Errorf("no frames found")
	}
	frame := pixelDataInfo.Frames[0]
	bits := 1 << frame.NativeData.BitsPerSample

	windowCenter, _ := dataset.FindElementByTag(tag.WindowCenter)
	var windowCenterValue int
	if _, err := fmt.Sscanf(windowCenter.Value.String(), "[%d.]", &windowCenterValue); err != nil {
		if _, err := fmt.Sscanf(windowCenter.Value.String(), "[%d]", &windowCenterValue); err != nil {
			windowCenterValue = bits/2 - 1
		}
	}

	windowWidth, _ := dataset.FindElementByTag(tag.WindowWidth)
	var windowWidthValue int
	if s, err := strconv.ParseFloat(windowWidth.Value.GetValue().([]string)[0], 64); err == nil {
		fmt.Println(s) // 3.1415927410125732
	}
	if s, err := strconv.ParseInt(windowWidth.Value.GetValue().([]string)[0], 10, 64); err == nil {
		fmt.Println(s) // 3
	}
	if _, err := fmt.Sscanf(windowWidth.Value.String(), "[%d.]", &windowWidthValue); err != nil {
		if _, err := fmt.Sscanf(windowWidth.Value.String(), "[%d]", &windowWidthValue); err != nil {
			windowWidthValue = bits - 1
		}
	}

	nativeData := frame.NativeData.Data
	nativeData = ScalePixelData(nativeData, frame.NativeData.BitsPerSample, Sigmoid, windowCenterValue, windowWidthValue)
	nativeData = NegatePixelData(nativeData, frame.NativeData.BitsPerSample)

	img := image.NewGray(image.Rect(0, 0, frame.NativeData.Cols, frame.NativeData.Rows))
	for j := 0; j < len(nativeData); j++ {
		img.SetGray(j%frame.NativeData.Cols, j/frame.NativeData.Cols, color.Gray{Y: uint8(nativeData[j][0])}) // for now, assume we're not overflowing uint16, assume gray image
	}

	return img, nil
}

func NegatePixelData(data [][]int, bits int) [][]int {
	max := (1 << uint(bits)) - 1
	for i := 0; i < len(data); i++ {
		data[i][0] = max - data[i][0]
	}
	return data
}

func ScalePixelData(data [][]int, bits int, scalingFunction ScalingFunction, windowWidth, windowCenter int) [][]int {
	switch scalingFunction {
	case Linear:
		return scaleLinear(data, bits, windowWidth, windowCenter)
	case LinearExact:
		return scaleLinearExact(data, bits, windowWidth, windowCenter)
	case Sigmoid:
		return scaleSigmoid(data, bits, windowWidth, windowCenter)
	default:
		return data
	}
}

func scaleLinear(data [][]int, bits int, width int, center int) [][]int {
	return data
}

func scaleLinearExact(data [][]int, bits int, width int, center int) [][]int {
	return data
}

func scaleSigmoid(data [][]int, bits int, width int, center int) [][]int {
	yMax := (1 << uint(bits)) - 1
	yMin := 0

	//for i := 0; i < len(data); i++ {
	//	for j := 0; j < len(data[i]); j++ {
	//		if data[i][j] > yMax {
	//			yMax = data[i][j]
	//		}
	//		if data[i][j] < yMin {
	//			yMin = data[i][j]
	//		}
	//	}
	//}

	origFreqMap := make(map[int]int)
	freqMap := make(map[int]int)
	for i := 0; i < len(data); i++ {
		origFreqMap[data[i][0]]++
		value := int(float64(yMax-yMin)/(1+math.Exp(float64(-4*(data[i][0]-center))/float64(width))) + float64(yMin))
		freqMap[value]++
		//fmt.Println(fmt.Sprintf("%d -> %d", data[i][j], value))
		data[i][0] = value
	}

	fmt.Println(freqMap)

	return data
}
