package notion

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jeb-maker/revues/internal/store"
)

type ColumnMapping struct {
	Label, Section, HelpText, Required string
}

type ImportPreview struct {
	TemplateName string
	Items        []store.TemplateItemInput
}

func DefaultMapping(db DatabaseInfo) ColumnMapping {
	var mapping ColumnMapping
	for _, prop := range db.Properties {
		switch prop.Type {
		case "title":
			if mapping.Label == "" {
				mapping.Label = prop.Name
			}
		case "select", "status":
			if mapping.Section == "" {
				mapping.Section = prop.Name
			}
		case "rich_text":
			if mapping.HelpText == "" {
				mapping.HelpText = prop.Name
			}
		case "checkbox":
			if mapping.Required == "" {
				mapping.Required = prop.Name
			}
		}
	}
	return mapping
}

func ValidateMapping(db DatabaseInfo, mapping ColumnMapping) error {
	label := strings.TrimSpace(mapping.Label)
	if label == "" {
		return errors.New("sélectionnez la colonne libellé")
	}
	for _, prop := range db.Properties {
		if prop.Name == label {
			if prop.Type != "title" {
				return fmt.Errorf("la colonne %q doit être de type titre Notion", label)
			}
			goto ok
		}
	}
	return fmt.Errorf("colonne %q introuvable dans la base Notion", label)
ok:
	for _, name := range []string{mapping.Section, mapping.HelpText, mapping.Required} {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if !propertyExists(db, name) {
			return fmt.Errorf("colonne %q introuvable dans la base Notion", name)
		}
	}
	return nil
}

func BuildImportPreview(db DatabaseInfo, pages []DatabasePage, mapping ColumnMapping, templateName string) (ImportPreview, error) {
	if err := ValidateMapping(db, mapping); err != nil {
		return ImportPreview{}, err
	}
	name := strings.TrimSpace(templateName)
	if name == "" {
		name = db.Title
	}
	if name == "" {
		name = "Modèle Notion"
	}
	var items []store.TemplateItemInput
	for _, page := range pages {
		label := strings.TrimSpace(extractPropertyText(page.Properties[mapping.Label], "title"))
		if label == "" {
			continue
		}
		item := store.TemplateItemInput{Label: label}
		if sec := strings.TrimSpace(mapping.Section); sec != "" {
			item.Section = extractPropertyText(page.Properties[sec], propertyType(db, sec))
		}
		if help := strings.TrimSpace(mapping.HelpText); help != "" {
			item.HelpText = extractPropertyText(page.Properties[help], propertyType(db, help))
		}
		if req := strings.TrimSpace(mapping.Required); req != "" {
			item.Required = extractPropertyBool(page.Properties[req])
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return ImportPreview{}, errors.New("aucun point importable trouvé dans la base Notion")
	}
	return ImportPreview{TemplateName: name, Items: items}, nil
}

func propertyExists(db DatabaseInfo, name string) bool {
	for _, prop := range db.Properties {
		if prop.Name == name {
			return true
		}
	}
	return false
}

func propertyType(db DatabaseInfo, name string) string {
	for _, prop := range db.Properties {
		if prop.Name == name {
			return prop.Type
		}
	}
	return ""
}

func extractPropertyText(raw json.RawMessage, propType string) string {
	if len(raw) == 0 {
		return ""
	}
	switch propType {
	case "title":
		var v struct {
			Title []plainTextFragment `json:"title"`
		}
		if json.Unmarshal(raw, &v) == nil {
			return joinPlainText(v.Title)
		}
	case "rich_text":
		var v struct {
			RichText []plainTextFragment `json:"rich_text"`
		}
		if json.Unmarshal(raw, &v) == nil {
			return joinPlainText(v.RichText)
		}
	case "select":
		var v struct {
			Select *struct {
				Name string `json:"name"`
			} `json:"select"`
		}
		if json.Unmarshal(raw, &v) == nil && v.Select != nil {
			return v.Select.Name
		}
	case "status":
		var v struct {
			Status *struct {
				Name string `json:"name"`
			} `json:"status"`
		}
		if json.Unmarshal(raw, &v) == nil && v.Status != nil {
			return v.Status.Name
		}
	}
	return ""
}

func extractPropertyBool(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var v struct {
		Checkbox bool `json:"checkbox"`
	}
	if json.Unmarshal(raw, &v) == nil {
		return v.Checkbox
	}
	return false
}
