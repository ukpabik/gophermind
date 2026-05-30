package dataset

type FeatureType int

const (
	BinaryType FeatureType = iota
	CategoricalType
	Float64Type
)

type Feature interface {
	GetName() string
	GetType() FeatureType
}

type BinaryFeature struct {
	Name   string
	Type   FeatureType
	Values []uint8
}

func (bf *BinaryFeature) GetName() string      { return bf.Name }
func (bf *BinaryFeature) GetType() FeatureType { return BinaryType }

type Float64Feature struct {
	Name   string
	Type   FeatureType
	Values []float64
}

func (ff *Float64Feature) GetName() string      { return ff.Name }
func (ff *Float64Feature) GetType() FeatureType { return Float64Type }
