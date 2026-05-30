package dataset_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ukpabik/gophermind/pkg/dataset"
)

func TestParseFromURL(t *testing.T) {
	fileURL := "https://download.mlcc.google.com/mledu-datasets/chicago_taxi_train.csv"

	// Columns: 'TRIP_MILES', 'TRIP_SECONDS', 'FARE', 'COMPANY', 'PAYMENT_TYPE', 'TIP_RATE'
	data, err := dataset.ParseCSVFromURL(fileURL)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	assert.True(t, len(data.Features) > 0, "Dataset is empty when it should be instantiated with values")
	checkedType := false

	for _, feature := range data.Features {
		switch feature.GetName() {
		case "TRIP_MILES":
			checkedType = true
			floatFeat, ok := feature.(*dataset.Float64Feature)
			if assert.True(t, ok, "Feature is not the correct type") {
				assert.True(t, len(floatFeat.Values) > 0, "Feature has no values")
			}
		}
	}

	assert.True(t, checkedType, "TRIP_MILES is not a column in this dataset")
}
