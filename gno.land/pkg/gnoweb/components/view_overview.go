package components

import (
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

const OverviewViewType ViewType = "overview-view"

type ImportLink struct {
	Name string
	Link string
}

type PackageInfo struct {
	Module      string
	GnoVersion  string
	Creator     string
	Height      int
	Draft       bool
	Private     bool
	License     string
	PackageType string
	HasTests    bool
	HasReadme   bool
}

type OverviewData struct {
	PkgPath      string
	Readme       Component
	Functions    []*doc.JSONFunc
	Doc          string
	Files        []string
	FileCounter  int
	Consts       []*doc.JSONValueDecl
	Vars         []*doc.JSONValueDecl
	Types        []*doc.JSONType
	Dirs         []string
	ComponentTOC Component
	Article      ArticleData
	PackageInfo  *PackageInfo
	Imports      []*ImportLink
	GroupedConsts []*GroupedDecl
	GroupedVars   []*GroupedDecl
}

type overviewViewParams struct {
	OverviewData
}

// GroupedDecl represents a grouped declaration
type GroupedDecl struct {
	*doc.JSONValueDecl
	Names string
	ID    string
}

// groupDecls groups declarations by their signature
func groupDecls(decls []*doc.JSONValueDecl, prefix string) []*GroupedDecl {
	var grouped []*GroupedDecl
	
	for _, decl := range decls {
		if len(decl.Values) > 0 {
			names := make([]string, len(decl.Values))
			for i, val := range decl.Values {
				names[i] = val.Name
			}
			grouped = append(grouped, &GroupedDecl{
				JSONValueDecl: decl,
				Names:         strings.Join(names, ", "),
				ID:            prefix + "-" + decl.Values[0].Name,
			})
		}
	}
	
	return grouped
}

func OverviewView(data OverviewData) *View {
	// Group consts and vars (ultra-simple, one pass)
	groupedConsts := groupDecls(data.Consts, "const")
	groupedVars := groupDecls(data.Vars, "var")
	
	// Create TOC data using TocItem structure
	var tocItems []*markdown.TocItem
	
	// Add sections based on available data
	if data.Readme != nil {
		tocItems = append(tocItems, &markdown.TocItem{
			Title: []byte("README"),
			ID:    []byte("readme"),
		})
	}
	
	if len(data.Functions) > 0 {
		section := &markdown.TocItem{
			Title: []byte("Functions"),
			ID:    []byte("functions"),
		}
		for _, fn := range data.Functions {
			section.Items = append(section.Items, &markdown.TocItem{
				Title: []byte(fn.Name),
				ID:    []byte("func-" + fn.Name),
			})
		}
		tocItems = append(tocItems, section)
	}
	
	if len(data.Consts) > 0 {
		tocItems = append(tocItems, &markdown.TocItem{
			Title: []byte("Constants"),
			ID:    []byte("constants"),
		})
	}
	
	if len(data.Vars) > 0 {
		tocItems = append(tocItems, &markdown.TocItem{
			Title: []byte("Variables"),
			ID:    []byte("variables"),
		})
	}
	
	if len(data.Types) > 0 {
		section := &markdown.TocItem{
			Title: []byte("Types"),
			ID:    []byte("types"),
		}
		for _, t := range data.Types {
			section.Items = append(section.Items, &markdown.TocItem{
				Title: []byte(t.Name),
				ID:    []byte("type-" + t.Name),
			})
		}
		tocItems = append(tocItems, section)
	}
	
	if len(data.Files) > 0 {
		tocItems = append(tocItems, &markdown.TocItem{
			Title: []byte("Source Files"),
			ID:    []byte("files"),
		})
	}
	
	if len(data.Dirs) > 0 {
		tocItems = append(tocItems, &markdown.TocItem{
			Title: []byte("Directories"),
			ID:    []byte("directories"),
		})
	}
	
	if len(data.Imports) > 0 {
		tocItems = append(tocItems, &markdown.TocItem{
			Title: []byte("Imports"),
			ID:    []byte("imports"),
		})
	}

	// Create TOC component
	var tocComponent Component
	if len(tocItems) > 0 {
		tocData := &markdown.Toc{Items: tocItems}
		tocComponent = NewTemplateComponent("ui/toc_realm", tocData)
	}

	// Create article content with grouped data
	articleData := OverviewData{
		PkgPath:        data.PkgPath,
		Readme:         data.Readme,
		Functions:      data.Functions,
		Doc:            data.Doc,
		Files:          data.Files,
		FileCounter:    data.FileCounter,
		Consts:         data.Consts,
		Vars:           data.Vars,
		Types:          data.Types,
		Dirs:           data.Dirs,
		Imports:        data.Imports,
		GroupedConsts:  groupedConsts,
		GroupedVars:    groupedVars,
		ComponentTOC:   tocComponent,
		Article: ArticleData{
			ComponentContent: nil, // Will be set below
			Classes:          "overview-view col-span-1 lg:col-span-7 pb-24 text-gray-900",
		},
		PackageInfo: data.PackageInfo,
	}
	
	articleContent := NewTemplateComponent("ui/overview_content", articleData)
	articleData.Article.ComponentContent = articleContent

	viewData := overviewViewParams{
		OverviewData: articleData,
	}

	return NewTemplateView(OverviewViewType, "renderOverview", viewData)
} 