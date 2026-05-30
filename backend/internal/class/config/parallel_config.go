package config

import (
	"fmt"
	"strconv"
	"strings"
)

type ParallelConfig struct {
	Name     string
	MinGrade int   
	MaxGrade int    
}

func ParseParallelsConfig(configStr string) ([]ParallelConfig, error) {
	if strings.TrimSpace(configStr) == "" {
		return []ParallelConfig{}, nil
	}

	var parallels []ParallelConfig
	ranges := strings.Split(configStr, ",")

	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parallel range format: '%s', expected format: '1-4'", r)
		}

		minGrade, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid min grade in range '%s': %w", r, err)
		}

		maxGrade, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid max grade in range '%s': %w", r, err)
		}

		if minGrade > maxGrade {
			return nil, fmt.Errorf("invalid range '%s': min grade (%d) cannot be greater than max grade (%d)", r, minGrade, maxGrade)
		}

		name := fmt.Sprintf("%d-%d параллель", minGrade, maxGrade)

		parallels = append(parallels, ParallelConfig{
			Name:     name,
			MinGrade: minGrade,
			MaxGrade: maxGrade,
		})
	}

	return parallels, nil
}
