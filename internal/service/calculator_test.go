package service

import (
	"pack-calculator/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackCalculatorService_CalculatePacks(t *testing.T) {
	config := models.NewPackConfiguration()
	service := NewPackCalculatorService(config)

	tests := []struct {
		name          string
		items         int
		expectedPacks map[int]int
		expectedItems int
		expectedTotal int
		shouldError   bool
	}{
		{
			name:          "1 item",
			items:         1,
			expectedPacks: map[int]int{250: 1},
			expectedItems: 250,
			expectedTotal: 1,
		},
		{
			name:          "250 items",
			items:         250,
			expectedPacks: map[int]int{250: 1},
			expectedItems: 250,
			expectedTotal: 1,
		},
		{
			name:          "251 items",
			items:         251,
			expectedPacks: map[int]int{500: 1},
			expectedItems: 500,
			expectedTotal: 1,
		},
		{
			name:          "501 items",
			items:         501,
			expectedPacks: map[int]int{500: 1, 250: 1},
			expectedItems: 750,
			expectedTotal: 2,
		},
		{
			name:          "12001 items",
			items:         12001,
			expectedPacks: map[int]int{5000: 2, 2000: 1, 250: 1},
			expectedItems: 12250,
			expectedTotal: 4,
		},
		{
			name:        "0 items should error",
			items:       0,
			shouldError: true,
		},
		{
			name:        "negative items should error",
			items:       -1,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculatePacks(tt.items)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.items, result.RequestedItems)
			assert.Equal(t, tt.expectedItems, result.TotalItems)
			assert.Equal(t, tt.expectedTotal, result.TotalPacks)
			assert.Equal(t, tt.expectedPacks, result.PackBreakdown)

			assert.GreaterOrEqual(t, result.TotalItems, result.RequestedItems)

			totalFromBreakdown := 0
			packsFromBreakdown := 0
			for size, count := range result.PackBreakdown {
				totalFromBreakdown += size * count
				packsFromBreakdown += count
			}
			assert.Equal(t, result.TotalItems, totalFromBreakdown)
			assert.Equal(t, result.TotalPacks, packsFromBreakdown)
		})
	}
}

func TestPackCalculatorService_EdgeCase(t *testing.T) {
	packSizes := []int{23, 31, 53}
	items := 500000

	config := &models.PackConfiguration{PackSizes: packSizes}
	service := NewPackCalculatorService(config)

	result, err := service.CalculatePacks(items)
	require.NoError(t, err)
	require.NotNil(t, result)

	expectedPacks := map[int]int{23: 2, 31: 7, 53: 9429}
	assert.Equal(t, expectedPacks, result.PackBreakdown)

	expectedTotal := 23*2 + 31*7 + 53*9429
	assert.Equal(t, expectedTotal, result.TotalItems)
	assert.Equal(t, 2+7+9429, result.TotalPacks)
	assert.Equal(t, items, result.RequestedItems)

	assert.GreaterOrEqual(t, result.TotalItems, items)

	t.Logf("Edge case result: %+v", result)
	t.Logf("Total items: %d, Total packs: %d", result.TotalItems, result.TotalPacks)
}

func TestPackCalculatorService_CalculatePacksWithCustomSizes(t *testing.T) {
	config := models.NewPackConfiguration()
	service := NewPackCalculatorService(config)

	tests := []struct {
		name          string
		items         int
		customSizes   []int
		shouldError   bool
		expectedPacks map[int]int
	}{
		{
			name:          "custom sizes - simple case",
			items:         100,
			customSizes:   []int{10, 25, 50},
			expectedPacks: map[int]int{50: 2},
		},
		{
			name:          "custom sizes - complex case",
			items:         37,
			customSizes:   []int{10, 25},
			expectedPacks: map[int]int{25: 1, 10: 2},
		},
		{
			name:        "empty custom sizes should error",
			items:       100,
			customSizes: []int{},
			shouldError: true,
		},
		{
			name:        "invalid custom sizes should error",
			items:       100,
			customSizes: []int{10, 0, 25},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculatePacksWithCustomSizes(tt.items, tt.customSizes)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.items, result.RequestedItems)
			assert.Equal(t, tt.expectedPacks, result.PackBreakdown)
			assert.GreaterOrEqual(t, result.TotalItems, result.RequestedItems)
		})
	}
}

func TestPackCalculatorService_UpdatePackSizes(t *testing.T) {
	config := models.NewPackConfiguration()
	service := NewPackCalculatorService(config)

	newSizes := []int{100, 200, 300}
	err := service.UpdatePackSizes(newSizes)
	assert.NoError(t, err)
	assert.Equal(t, newSizes, service.GetPackSizes())

	invalidSizes := []int{100, 0, 300}
	err = service.UpdatePackSizes(invalidSizes)
	assert.Error(t, err)
	assert.Equal(t, newSizes, service.GetPackSizes())

	err = service.UpdatePackSizes([]int{})
	assert.Error(t, err)
	assert.Equal(t, newSizes, service.GetPackSizes())
}

func TestPackCalculatorService_GetPackSizes(t *testing.T) {
	config := models.NewPackConfiguration()
	service := NewPackCalculatorService(config)

	sizes := service.GetPackSizes()
	expected := []int{250, 500, 1000, 2000, 5000}
	assert.Equal(t, expected, sizes)
}

func BenchmarkCalculatePacks_SmallOrder(b *testing.B) {
	config := models.NewPackConfiguration()
	service := NewPackCalculatorService(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CalculatePacks(1000)
	}
}

func BenchmarkCalculatePacks_LargeOrder(b *testing.B) {
	config := models.NewPackConfiguration()
	service := NewPackCalculatorService(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CalculatePacks(100000)
	}
}

func BenchmarkCalculatePacks_EdgeCase(b *testing.B) {
	packSizes := []int{23, 31, 53}
	config := &models.PackConfiguration{PackSizes: packSizes}
	service := NewPackCalculatorService(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CalculatePacks(500000)
	}
}
