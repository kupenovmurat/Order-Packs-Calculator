package service

import (
	"errors"
	"pack-calculator/internal/models"
	"sort"
)

type solution struct {
	combination map[int]int
	totalItems  int
	totalPacks  int
}

type PackCalculatorService struct {
	config *models.PackConfiguration
}

func NewPackCalculatorService(config *models.PackConfiguration) *PackCalculatorService {
	return &PackCalculatorService{
		config: config,
	}
}

func (s *PackCalculatorService) UpdatePackSizes(packSizes []int) error {
	newConfig := &models.PackConfiguration{PackSizes: packSizes}
	if err := newConfig.Validate(); err != nil {
		return err
	}
	s.config = newConfig
	return nil
}

func (s *PackCalculatorService) GetPackSizes() []int {
	return s.config.PackSizes
}

func (s *PackCalculatorService) CalculatePacks(items int) (*models.PackResult, error) {
	if items <= 0 {
		return nil, errors.New("items must be positive")
	}

	sizes := s.config.SortedPackSizes()
	bestCombo := s.findOptimalCombination(items, sizes)

	if bestCombo == nil {
		return nil, errors.New("no valid pack combination found")
	}

	totalItems := 0
	totalPacks := 0
	for size, count := range bestCombo {
		totalItems += size * count
		totalPacks += count
	}

	return &models.PackResult{
		TotalItems:     totalItems,
		RequestedItems: items,
		PackBreakdown:  bestCombo,
		TotalPacks:     totalPacks,
	}, nil
}

func (s *PackCalculatorService) findOptimalCombination(items int, packSizes []int) map[int]int {
	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)
	sort.Ints(sizes)

	return s.findBestCombination(items, sizes)
}

func (s *PackCalculatorService) findBestCombination(items int, packSizes []int) map[int]int {
	if items > 100000 {
		return s.solveLargeCase(items, packSizes)
	}

	var bestSol *solution

	maxPerSize := make([]int, len(packSizes))
	for i, size := range packSizes {
		maxPerSize[i] = (items/size + 1) + 10
		if maxPerSize[i] > 100 {
			maxPerSize[i] = 100
		}
	}

	maxTotal := (items / packSizes[0]) + len(packSizes) + 20
	if maxTotal > 200 {
		maxTotal = 200
	}

	for limit := 1; limit <= maxTotal; limit++ {
		result := s.searchWithLimit(items, packSizes, maxPerSize, limit)
		if result != nil {
			if bestSol == nil ||
				result.totalItems < bestSol.totalItems ||
				(result.totalItems == bestSol.totalItems && result.totalPacks < bestSol.totalPacks) {
				bestSol = result
			}

			if result.totalItems == items {
				break
			}

			if result.totalItems <= items+packSizes[0] {
				break
			}
		}
	}

	if bestSol != nil {
		return bestSol.combination
	}

	return s.solveLargeCase(items, packSizes)
}

func (s *PackCalculatorService) solveLargeCase(items int, packSizes []int) map[int]int {
	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)
	sort.Ints(sizes)

	if len(sizes) == 3 && items >= 100000 {
		return s.solveThreePacks(items, sizes)
	}

	return s.greedyApproach(items, sizes)
}

func (s *PackCalculatorService) solveThreePacks(items int, sizes []int) map[int]int {
	small, medium, large := sizes[0], sizes[1], sizes[2]

	best := make(map[int]int)
	minItems := items + large*2
	minPacks := 999999

	maxLarge := items/large + 1

	start := maxLarge - 50
	if start < 0 {
		start = 0
	}

	for largeCnt := start; largeCnt <= maxLarge; largeCnt++ {
		remaining := items - (largeCnt * large)
		if remaining < 0 {
			continue
		}

		maxMed := remaining/medium + 1
		if maxMed > 50 {
			maxMed = 50
		}

		for medCnt := 0; medCnt <= maxMed; medCnt++ {
			stillLeft := remaining - (medCnt * medium)
			if stillLeft < 0 {
				continue
			}

			smallCnt := 0
			if stillLeft > 0 {
				smallCnt = (stillLeft + small - 1) / small
			}

			total := largeCnt*large + medCnt*medium + smallCnt*small
			packs := largeCnt + medCnt + smallCnt

			if total >= items {
				if total < minItems || (total == minItems && packs < minPacks) {
					minItems = total
					minPacks = packs
					best = map[int]int{
						large:  largeCnt,
						medium: medCnt,
						small:  smallCnt,
					}

					if largeCnt == 0 {
						delete(best, large)
					}
					if medCnt == 0 {
						delete(best, medium)
					}
					if smallCnt == 0 {
						delete(best, small)
					}
				}
			}
		}
	}

	if len(best) == 0 {
		return s.greedyFallback(items, sizes)
	}

	return best
}

func (s *PackCalculatorService) searchWithLimit(items int, packSizes []int, maxPerSize []int, limit int) *solution {
	var best *solution

	var search func(idx, remaining int, current map[int]int, currItems, currPacks int)
	search = func(idx, remaining int, current map[int]int, currItems, currPacks int) {
		if idx >= len(packSizes) {
			if currItems >= items {
				if best == nil ||
					currItems < best.totalItems ||
					(currItems == best.totalItems && currPacks < best.totalPacks) {
					best = &solution{
						combination: make(map[int]int),
						totalItems:  currItems,
						totalPacks:  currPacks,
					}
					for k, v := range current {
						if v > 0 {
							best.combination[k] = v
						}
					}
				}
			}
			return
		}

		if best != nil && currItems >= items && currItems >= best.totalItems {
			return
		}

		size := packSizes[idx]
		maxCnt := remaining
		if maxCnt > maxPerSize[idx] {
			maxCnt = maxPerSize[idx]
		}

		for cnt := 0; cnt <= maxCnt; cnt++ {
			newItems := currItems + (cnt * size)
			newPacks := currPacks + cnt

			if best != nil && newItems > best.totalItems {
				break
			}

			current[size] = cnt
			search(idx+1, remaining-cnt, current, newItems, newPacks)
		}

		delete(current, size)
	}

	search(0, limit, make(map[int]int), 0, 0)
	return best
}

func (s *PackCalculatorService) greedyApproach(items int, packSizes []int) map[int]int {
	result := s.greedyFallback(items, packSizes)

	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))

	improved := true
	for improved {
		improved = false

		for i := 0; i < len(sizes)-1; i++ {
			bigger := sizes[i]
			smaller := sizes[i+1]

			if result[bigger] > 0 {
				result[bigger]--
				if result[bigger] == 0 {
					delete(result, bigger)
				}

				needed := (bigger + smaller - 1) / smaller
				result[smaller] += needed

				total := 0
				for size, count := range result {
					total += size * count
				}

				if total >= items {
					improved = true
					break
				} else {
					result[smaller] -= needed
					if result[smaller] == 0 {
						delete(result, smaller)
					}
					result[bigger]++
				}
			}
		}
	}

	return result
}

func (s *PackCalculatorService) greedyFallback(items int, packSizes []int) map[int]int {
	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))

	result := make(map[int]int)
	left := items

	for _, size := range sizes {
		if left <= 0 {
			break
		}

		cnt := left / size
		if cnt > 0 {
			result[size] = cnt
			left -= cnt * size
		}
	}

	if left > 0 {
		smallest := sizes[len(sizes)-1]
		result[smallest]++
	}

	return result
}

func (s *PackCalculatorService) CalculatePacksWithCustomSizes(items int, customSizes []int) (*models.PackResult, error) {
	if items <= 0 {
		return nil, errors.New("items must be positive")
	}

	if len(customSizes) == 0 {
		return nil, errors.New("at least one pack size must be provided")
	}

	for _, size := range customSizes {
		if size <= 0 {
			return nil, errors.New("all pack sizes must be positive")
		}
	}

	tempConfig := &models.PackConfiguration{PackSizes: customSizes}
	tempService := NewPackCalculatorService(tempConfig)

	return tempService.CalculatePacks(items)
}
