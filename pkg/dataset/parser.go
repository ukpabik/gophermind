package dataset

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ColumnMapping struct {
	FeatureRef Feature
	Category   string
}

func ParseCSVFromURL(fileURL string) (*Dataset, error) {
	if !strings.HasSuffix(fileURL, ".csv") {
		return nil, errors.New("file is not a remote .csv file")
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "csv_data_*.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %v", err)
	}

	_, _ = tmpFile.Seek(0, 0)
	reader := csv.NewReader(tmpFile)

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %v", err)
	}

	numCols := len(headers)
	isFloatCol := make([]bool, numCols)
	uniqueValsCol := make([]map[string]bool, numCols)
	for i := range numCols {
		isFloatCol[i] = true
		uniqueValsCol[i] = make(map[string]bool)
	}

	totalRows := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error streaming type inference: %v", err)
		}
		totalRows++

		for colIdx, val := range record {
			val = strings.TrimSpace(val)
			if val == "" {
				continue
			}
			uniqueValsCol[colIdx][val] = true

			if isFloatCol[colIdx] {
				if _, err := strconv.ParseFloat(val, 64); err != nil {
					isFloatCol[colIdx] = false
				}
			}
		}
	}

	if totalRows == 0 {
		return nil, errors.New("csv contains no data rows")
	}

	features := make([]Feature, 0)
	csvToFeaturesMap := make([][]ColumnMapping, numCols)

	for i := range numCols {
		numUnique := len(uniqueValsCol[i])

		if isFloatCol[i] && numUnique > 0 {
			f := &Float64Feature{
				Name:   headers[i],
				Type:   Float64Type,
				Values: make([]float64, totalRows),
			}
			features = append(features, f)
			csvToFeaturesMap[i] = []ColumnMapping{{FeatureRef: f}}

		} else if numUnique == 2 {
			b := &BinaryFeature{
				Name:   headers[i],
				Type:   BinaryType,
				Values: make([]uint8, totalRows),
			}
			features = append(features, b)
			csvToFeaturesMap[i] = []ColumnMapping{{FeatureRef: b}}

		} else {
			csvToFeaturesMap[i] = make([]ColumnMapping, 0, numUnique)

			for val := range uniqueValsCol[i] {
				cleanVal := strings.ReplaceAll(val, " ", "_")
				flattenedName := fmt.Sprintf("%s_IS_%s", headers[i], cleanVal)

				bSub := &BinaryFeature{
					Name:   flattenedName,
					Type:   BinaryType,
					Values: make([]uint8, totalRows),
				}
				features = append(features, bSub)
				csvToFeaturesMap[i] = append(csvToFeaturesMap[i], ColumnMapping{FeatureRef: bSub, Category: val})
			}
		}
	}

	_, _ = tmpFile.Seek(0, 0)
	reader = csv.NewReader(tmpFile)
	_, _ = reader.Read()

	currentRow := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error streaming data collection: %v", err)
		}

		for colIdx, val := range record {
			val = strings.TrimSpace(val)
			mappings := csvToFeaturesMap[colIdx]

			for _, mapping := range mappings {
				switch feat := mapping.FeatureRef.(type) {

				case *Float64Feature:
					if val != "" {
						fVal, _ := strconv.ParseFloat(val, 64)
						feat.Values[currentRow] = fVal
					}

				case *BinaryFeature:
					if mapping.Category != "" {
						if val == mapping.Category {
							feat.Values[currentRow] = 1
						} else {
							feat.Values[currentRow] = 0
						}
					} else {
						var targetKey string
						for k := range uniqueValsCol[colIdx] {
							targetKey = k
							break
						}
						if val == targetKey {
							feat.Values[currentRow] = 1
						} else {
							feat.Values[currentRow] = 0
						}
					}
				}
			}
		}
		currentRow++
	}

	return &Dataset{Features: features}, nil
}
