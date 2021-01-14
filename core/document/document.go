package document

// Document document
type Document struct {
	boost  float64
	fields []Fieldable
}

// setBoost
func (d *Document) setBoost(boost float64) error {
	d.boost = boost
	return nil
}

// getBoost
func (d *Document) getBoost() (float64, error) {
	return d.boost, nil
}

// add field
func (d *Document) add(field Fieldable) error {
	d.fields = append(d.fields, field)
	return nil
}

// get field
func (d *Document) getField(name string) (*Field, error) {
	// fieldable, err := d.getFieldable(name)
	// if err != nil {
	// 	return nil, err
	// }
	return nil, nil
	// field, err := (fieldable).(Field)
	// if err != nil {
	// 	return nil, err
	// }
	// return &field, nil
}

// get fieldable
func (d *Document) getFieldable(name string) (Fieldable, error) {
	return nil, nil
}
