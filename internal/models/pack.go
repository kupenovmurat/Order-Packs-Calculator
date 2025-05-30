package models

import (
	"errors"
	"sort"
)

type PackSize struct {
	Size int `json:"size" binding:"required,min=1"`
}

type PackConfiguration struct {
	PackSizes []int `json:"pack_sizes" binding:"required,min=1"`
}

type PackResult struct {
	TotalItems     int         `json:"total_items"`
	RequestedItems int         `json:"requested_items"`
	PackBreakdown  map[int]int `json:"pack_breakdown"`
	TotalPacks     int         `json:"total_packs"`
}

type CalculateRequest struct {
	Items     int   `json:"items" binding:"required,min=1"`
	PackSizes []int `json:"pack_sizes,omitempty"`
}

func (pc *PackConfiguration) Validate() error {
	if len(pc.PackSizes) == 0 {
		return errors.New("at least one pack size must be provided")
	}

	for _, size := range pc.PackSizes {
		if size <= 0 {
			return errors.New("all pack sizes must be positive")
		}
	}

	return nil
}

func (pc *PackConfiguration) SortedPackSizes() []int {
	sizes := make([]int, len(pc.PackSizes))
	copy(sizes, pc.PackSizes)
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	return sizes
}

func NewPackConfiguration() *PackConfiguration {
	return &PackConfiguration{
		PackSizes: []int{250, 500, 1000, 2000, 5000},
	}
}
