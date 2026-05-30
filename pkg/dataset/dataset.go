package dataset

type Dataset struct {
	Features []Feature
}

func NewDataset(features []Feature) *Dataset {
	return &Dataset{
		Features: features,
	}
}
