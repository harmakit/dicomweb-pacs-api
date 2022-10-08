package utils

import (
	"math"
)

type ScalingFunction int

const (
	Linear ScalingFunction = iota
	LinearExact
	Sigmoid
)

func NegatePixelData(data [][]int, bits int) [][]int {
	max := (1 << uint(bits)) - 1
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			data[i][j] = max - data[i][j]
		}
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
	yMax := 0
	yMin := (1 << uint(bits)) - 1

	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			if data[i][j] > yMax {
				yMax = data[i][j]
			}
			if data[i][j] < yMin {
				yMin = data[i][j]
			}
		}
	}

	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			value := int(float64(yMax-yMin)/(1+math.Exp(float64(-4*(data[i][j]-center))/float64(width))) + float64(yMin))
			//fmt.Println(fmt.Sprintf("%d -> %d", data[i][j], value))
			data[i][j] = value
		}
	}

	return data
}
