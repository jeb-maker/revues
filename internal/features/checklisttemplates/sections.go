package checklisttemplates

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/store"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

func emptyEditorSections(pointCount int) []viewtemplates.TemplateEditorSection {
	items := make([]viewtemplates.TemplateEditorRow, pointCount)
	for i := range items {
		items[i].RowIndex = i
	}
	return []viewtemplates.TemplateEditorSection{{
		SectionIndex: 0,
		Items:        items,
	}}
}

func itemsToEditorSections(items []store.TemplateItem) []viewtemplates.TemplateEditorSection {
	if len(items) == 0 {
		return emptyEditorSections(1)
	}

	sections := make([]viewtemplates.TemplateEditorSection, 0)
	var current *viewtemplates.TemplateEditorSection
	rowIdx := 0

	for _, item := range items {
		if current == nil || current.Title != item.Section {
			sections = append(sections, viewtemplates.TemplateEditorSection{
				SectionIndex: len(sections),
				Title:        item.Section,
			})
			current = &sections[len(sections)-1]
		}
		current.Items = append(current.Items, viewtemplates.TemplateEditorRow{
			RowIndex: rowIdx,
			Label:    item.Label,
			HelpText: item.HelpText,
			Required: item.Required,
		})
		rowIdx++
	}
	return sections
}

func groupTemplateItems(items []store.TemplateItem) []viewtemplates.TemplateItemSection {
	if len(items) == 0 {
		return nil
	}

	sections := make([]viewtemplates.TemplateItemSection, 0)
	var current *viewtemplates.TemplateItemSection

	for _, item := range items {
		if current == nil || current.Title != item.Section {
			sections = append(sections, viewtemplates.TemplateItemSection{Title: item.Section})
			current = &sections[len(sections)-1]
		}
		current.Items = append(current.Items, item)
	}
	return sections
}

func sectionsEnabled(sections []viewtemplates.TemplateEditorSection) bool {
	if len(sections) > 1 {
		return true
	}
	for _, sec := range sections {
		if strings.TrimSpace(sec.Title) != "" {
			return true
		}
	}
	return false
}

func parseTemplateItems(r *http.Request) ([]store.TemplateItemInput, string) {
	labels := r.Form["item_label"]
	helps := r.Form["item_help"]
	rowIndices := r.Form["item_row_idx"]
	sectionIdxs := r.Form["item_section_idx"]
	sectionBlockIdxs := r.Form["section_idx"]
	sectionTitles := r.Form["section_title"]
	legacySections := r.Form["item_section"]

	if len(labels) != len(helps) {
		return nil, "Les lignes du modèle sont incohérentes."
	}
	if len(rowIndices) != 0 && len(rowIndices) != len(labels) {
		return nil, "Les lignes du modèle sont incohérentes."
	}

	useSectionIdx := len(sectionIdxs) == len(labels)
	if !useSectionIdx {
		if len(legacySections) != len(labels) {
			return nil, "Les lignes du modèle sont incohérentes."
		}
	}

	titleBySection := sectionTitleMap(sectionBlockIdxs, sectionTitles)

	required := map[int]bool{}
	for _, raw := range r.Form["item_required"] {
		index, err := strconv.Atoi(raw)
		if err != nil || index < 0 {
			return nil, "Point requis invalide."
		}
		required[index] = true
	}

	useRowIdx := len(rowIndices) == len(labels)

	var items []store.TemplateItemInput
	for i := range labels {
		label := strings.TrimSpace(labels[i])
		if label == "" {
			continue
		}

		section := ""
		if useSectionIdx {
			if si, err := strconv.Atoi(sectionIdxs[i]); err == nil {
				section = titleBySection[si]
			}
		} else if i < len(legacySections) {
			section = strings.TrimSpace(legacySections[i])
		}

		isRequired := false
		if useRowIdx {
			if idx, err := strconv.Atoi(rowIndices[i]); err == nil {
				isRequired = required[idx]
			}
		} else {
			isRequired = required[i]
		}

		items = append(items, store.TemplateItemInput{
			Section:  section,
			Label:    label,
			HelpText: strings.TrimSpace(helps[i]),
			Required: isRequired,
		})
	}

	return items, ""
}

func sectionTitleMap(sectionBlockIdxs, sectionTitles []string) map[int]string {
	out := make(map[int]string, len(sectionTitles))
	for i, title := range sectionTitles {
		if i >= len(sectionBlockIdxs) {
			continue
		}
		si, err := strconv.Atoi(sectionBlockIdxs[i])
		if err != nil {
			continue
		}
		out[si] = strings.TrimSpace(title)
	}
	return out
}

func parseTemplateItemsToSections(r *http.Request) []viewtemplates.TemplateEditorSection {
	labels := r.Form["item_label"]
	helps := r.Form["item_help"]
	rowIndices := r.Form["item_row_idx"]
	sectionIdxs := r.Form["item_section_idx"]
	sectionBlockIdxs := r.Form["section_idx"]
	sectionTitles := r.Form["section_title"]
	legacySections := r.Form["item_section"]

	maxLen := len(labels)
	if len(helps) > maxLen {
		maxLen = len(helps)
	}

	required := map[int]bool{}
	for _, raw := range r.Form["item_required"] {
		index, err := strconv.Atoi(raw)
		if err == nil {
			required[index] = true
		}
	}

	useRowIdx := len(rowIndices) == maxLen
	useSectionIdx := len(sectionIdxs) == maxLen

	rows := make([]viewtemplates.TemplateEditorRow, maxLen)
	for i := 0; i < maxLen; i++ {
		if i < len(labels) {
			rows[i].Label = labels[i]
		}
		if i < len(helps) {
			rows[i].HelpText = helps[i]
		}
		if useRowIdx {
			if idx, err := strconv.Atoi(rowIndices[i]); err == nil {
				rows[i].RowIndex = idx
				rows[i].Required = required[idx]
			}
		} else {
			rows[i].RowIndex = i
			rows[i].Required = required[i]
		}
	}

	if useSectionIdx {
		return rowsToEditorSections(rows, sectionBlockIdxs, sectionTitles, sectionIdxs)
	}

	return flatRowsToEditorSections(rows, legacySections)
}

func rowsToEditorSections(rows []viewtemplates.TemplateEditorRow, sectionBlockIdxs, sectionTitles, itemSectionIdxs []string) []viewtemplates.TemplateEditorSection {
	if len(rows) == 0 {
		return emptyEditorSections(1)
	}

	titleBySection := sectionTitleMap(sectionBlockIdxs, sectionTitles)
	sectionsByIdx := map[int]*viewtemplates.TemplateEditorSection{}
	order := make([]int, 0)

	for i, row := range rows {
		si := 0
		if i < len(itemSectionIdxs) {
			if n, err := strconv.Atoi(itemSectionIdxs[i]); err == nil && n >= 0 {
				si = n
			}
		}
		sec, ok := sectionsByIdx[si]
		if !ok {
			sectionsByIdx[si] = &viewtemplates.TemplateEditorSection{
				SectionIndex: si,
				Title:        titleBySection[si],
			}
			order = append(order, si)
			sec = sectionsByIdx[si]
		}
		sec.Items = append(sec.Items, row)
	}

	sections := make([]viewtemplates.TemplateEditorSection, 0, len(order))
	for _, si := range order {
		sections = append(sections, *sectionsByIdx[si])
	}
	if len(sections) == 0 {
		return emptyEditorSections(1)
	}
	return sections
}

func flatRowsToEditorSections(rows []viewtemplates.TemplateEditorRow, sections []string) []viewtemplates.TemplateEditorSection {
	if len(rows) == 0 {
		return emptyEditorSections(1)
	}

	out := make([]viewtemplates.TemplateEditorSection, 0)
	var current *viewtemplates.TemplateEditorSection

	for i, row := range rows {
		title := ""
		if i < len(sections) {
			title = sections[i]
		}
		if strings.TrimSpace(row.Label) == "" && strings.TrimSpace(title) == "" && i > 0 {
			continue
		}
		if current == nil || current.Title != title {
			out = append(out, viewtemplates.TemplateEditorSection{
				SectionIndex: len(out),
				Title:        title,
			})
			current = &out[len(out)-1]
		}
		current.Items = append(current.Items, row)
	}
	if len(out) == 0 {
		return emptyEditorSections(1)
	}
	return out
}
