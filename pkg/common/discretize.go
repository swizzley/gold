package common

import (
	"fmt"

	tens "github.com/pbarker/go-rl/pkg/tensor"
	"gorgonia.org/tensor"
)

// EqWidthBinner implements the equal width binning algorithm. This is used to
// discretize continuous spaces.
type EqWidthBinner struct {
	// NumIntervals is the number of bins.
	NumIntervals int

	// High is the upper bound of the continuous space.
	High float32

	// Low is the lower bound of the continuous space.
	Low float32

	width      float32
	boundaries []float32
}

// NewEqWidthBinner returns an EqWidthBin.
func NewEqWidthBinner(numIntervals int, high, low float32) *EqWidthBinner {
	width := (high - low) / float32(numIntervals)
	return &EqWidthBinner{
		NumIntervals: numIntervals,
		High:         high,
		Low:          low,
		width:        width,
	}
}

// Bin the given value.
func (e *EqWidthBinner) Bin(v float32) (int, error) {
	for i := 0; i <= e.NumIntervals; i++ {
		if v < e.Low {
			return 0, fmt.Errorf("v: %f not in range %f to %f", v, e.Low, e.High)
		}
		if v < (e.Low + (float32((i + 1)) * e.width)) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("v: %f not in range %f to %f", v, e.Low, e.High)
}

// DenseEqWidthBinner is an EqWidthBinner applied using tensors.
type DenseEqWidthBinner struct {
	// Intervals to bin.
	Intervals *tensor.Dense

	// Low values are the lower bounds.
	Low *tensor.Dense

	// High values are the upper bounds.
	High *tensor.Dense

	widths *tensor.Dense
	bounds []*tensor.Dense
}

// NewDenseEqWidthBinner is a new dense binner.
func NewDenseEqWidthBinner(intervals, low, high *tensor.Dense) (*DenseEqWidthBinner, error) {
	var err error

	// make types homogenous
	low, err = tens.ToFloat32(low)
	if err != nil {
		return nil, err
	}
	high, err = tens.ToFloat32(high)
	if err != nil {
		return nil, err
	}
	intervals, err = tens.ToFloat32(intervals)
	if err != nil {
		return nil, err
	}

	// width = (max - min)/n
	spread, err := high.Sub(low)
	if err != nil {
		return nil, err
	}
	widths, err := spread.Div(intervals)
	if err != nil {
		return nil, err
	}
	var bounds []*tensor.Dense
	iterator := widths.Iterator()
	for i, err := iterator.Next(); err == nil; i, err = iterator.Next() {
		interval := intervals.GetF32(i)
		l := low.GetF32(i)
		width := widths.GetF32(i)
		backing := []float32{l}
		for j := 0; j <= int(interval); j++ {
			backing = append(backing, backing[j]+width)
		}
		bound := tensor.New(tensor.WithShape(1, len(backing)), tensor.WithBacking(backing))
		bounds = append(bounds, bound)
	}
	return &DenseEqWidthBinner{
		Intervals: intervals,
		Low:       low,
		High:      high,
		widths:    widths,
		bounds:    bounds,
	}, nil
}

// Bin the values.
func (d *DenseEqWidthBinner) Bin(values *tensor.Dense) (*tensor.Dense, error) {
	iterator := values.Iterator()
	backing := []int{}
	for i, err := iterator.Next(); err == nil; i, err = iterator.Next() {
		v := values.GetF32(i)
		bounds := d.bounds[i]
		bIter := bounds.Iterator()
		for j, err := bIter.Next(); err == nil; j, err = bIter.Next() {
			if v < bounds.GetF32(0) {
				return nil, fmt.Errorf("could not bin %v, out of range %v", v, bounds)
			}
			if v < bounds.GetF32(j) {
				backing = append(backing, j)
				break
			}
		}
		if len(backing) <= i {
			return nil, fmt.Errorf("could not bin %v, out of range %v", v, bounds)
		}
	}
	binned := tensor.New(tensor.WithShape(values.Shape()...), tensor.WithBacking(backing))
	return binned, nil
}

// Widths used in binning.
func (d *DenseEqWidthBinner) Widths() *tensor.Dense {
	return d.widths
}

// Bounds used in binning.
func (d *DenseEqWidthBinner) Bounds() []*tensor.Dense {
	return d.bounds
}
