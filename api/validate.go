package api

import (
	"cbs/models"

	"github.com/mitchellh/mapstructure"

	"gopkg.in/go-playground/validator.v9"
)

// UniverseFieldStructLevelValidation validates a universe field
// while paying special attention to its set "meta" type
func UniverseFieldStructLevelValidation(sl validator.StructLevel) {
	var err error
	field := sl.Current().Interface().(models.UniverseGuideField)
	v := sl.Validator()
	switch field.Type {
	case models.GuideFieldText:
		m := models.UniverseGuideMetaText{}
		err = mapstructure.Decode(field.Meta, &m)
		field.Meta = m
	case models.GuideFieldDescription:
		m := models.UniverseGuideMetaDescription{}
		err = mapstructure.Decode(field.Meta, &m)
		field.Meta = m
	case models.GuideFieldNumber:
		m := models.UniverseGuideMetaNumber{}
		err = mapstructure.Decode(field.Meta, &m)
		field.Meta = m
	case models.GuideFieldProgress:
		m := models.UniverseGuideMetaProgress{}
		err = mapstructure.Decode(field.Meta, &m)
		field.Meta = m
	case models.GuideFieldOptions:
		m := models.UniverseGuideMetaOptions{}
		err = mapstructure.Decode(field.Meta, &m)
		field.Meta = m
	case models.GuideFieldList:
		m := models.UniverseGuideMetaList{}
		err = mapstructure.Decode(field.Meta, &m)
		field.Meta = m
	}
	if err != nil {
		sl.ReportError(field, "meta", "Meta", "", "")
		return
	}
	if err := v.Struct(field.Meta); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			sl.ReportError(field, "field", "Field", "", "")
			return
		}
		sl.ReportValidationErrors("field", "field", ve)
	}
}
