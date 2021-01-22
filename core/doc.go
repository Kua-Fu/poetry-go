package core

// Document document
type Document struct {
	Boost  float64
	Fields []Field
}

// SetBoost setBoost
func (d *Document) SetBoost(boost float64) error {
	d.Boost = boost
	return nil
}

// Add add field
func (d *Document) Add(field Field) error {
	d.Fields = append(d.Fields, field)
	return nil
}
