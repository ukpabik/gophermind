package model

import "github.com/ukpabik/gophermind/pkg/dataset"

type Model interface {
	Fit(data *dataset.Dataset)
	Predict(data *dataset.Dataset) ([]float64, error)
}
