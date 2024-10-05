package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
)

var (
	//go:embed templates/package_diff_template.html
	packageDiffTemplate string
	//go:embed templates/index_template.html
	indexTemplate string
)

// ReportBuilder is a struct for building reports based on the differences
// between source and destination directories.
type ReportBuilder struct {
	SrcPath         string             // Source directory path.
	DstPath         string             // Destination directory path.
	OutDir          string             // Output directory path for the reports.
	SrcIsGno        bool               // Indicates if the Src files are gno files.
	packageTemplate *template.Template // Template for generating reports.
	indexTemplate   *template.Template // Template for generating index file of the reports.
}

// PackageDiffTemplateData represents the template data structure for a package's
// differences between source and destination directories.
type PackageDiffTemplateData struct {
	PackageName        string           // Package name.
	SrcFilesCount      int              // Number of files in the source package.
	SrcPackageLocation string           // Location of source files in the source directory.
	DstFileCount       int              // Number of destination files in the package.
	DstPackageLocation string           // Location of destination files in the destination directory.
	FilesDifferences   []FileDifference // Differences in individual files.
}

type IndexTemplate struct {
	Reports []LinkToReport
}

type LinkToReport struct {
	PathToReport string
	PackageName  string
	WasFound     bool
}

// NewReportBuilder creates a new ReportBuilder instance with the specified
// source path, destination path, and output directory. It also initializes
// the packageTemplate using the provided HTML template file.
func NewReportBuilder(srcPath, dstPath, outDir string, srcIsGno bool) (*ReportBuilder, error) {
	packageTemplate, err := template.New("").Parse(packageDiffTemplate)
	if err != nil {
		return nil, err
	}

	indexTemplate, err := template.New("").Parse(indexTemplate)
	if err != nil {
		return nil, err
	}

	return &ReportBuilder{
		SrcPath:         srcPath,
		DstPath:         dstPath,
		OutDir:          outDir,
		SrcIsGno:        srcIsGno,
		packageTemplate: packageTemplate,
		indexTemplate:   indexTemplate,
	}, nil
}

// Build generates reports for differences between packages in the source and
// destination directories. It iterates through each directory, calculates
// differences using PackageDiffChecker, and generates reports using the
// packageTemplate.
func (builder *ReportBuilder) Build() error {
	directories, err := builder.listSrcDirectories()
	if err != nil {
		return err
	}

	indexTemplateData := &IndexTemplate{
		Reports: make([]LinkToReport, 0),
	}

	for _, directory := range directories {
		if directory.FoundInDest {
			srcPackagePath := builder.SrcPath + "/" + directory.Path
			dstPackagePath := builder.DstPath + "/" + directory.Path
			packageChecker, err := NewPackageDiffChecker(srcPackagePath, dstPackagePath, builder.SrcIsGno)
			if err != nil {
				return fmt.Errorf("can't create new PackageDiffChecker: %w", err)
			}

			differences, err := packageChecker.Differences()
			if err != nil {
				return fmt.Errorf("can't compute differences: %w", err)
			}

			data := &PackageDiffTemplateData{
				PackageName:        directory.Path,
				SrcFilesCount:      len(packageChecker.SrcFiles),
				SrcPackageLocation: srcPackagePath,
				DstFileCount:       len(packageChecker.DstFiles),
				DstPackageLocation: dstPackagePath,
				FilesDifferences:   differences.FilesDifferences,
			}

			if err := builder.writePackageTemplate(data, directory.Path); err != nil {
				return err
			}
		}

		indexTemplateData.Reports = append(indexTemplateData.Reports, LinkToReport{
			PathToReport: "./" + directory.Path + "/report.html",
			PackageName:  directory.Path,
			WasFound:     directory.FoundInDest,
		})
	}

	if err := builder.writeIndexTemplate(indexTemplateData); err != nil {
		return err
	}

	return nil
}

type Directory struct {
	Path        string
	FoundInDest bool
}

// listSrcDirectories retrieves a list of directories in the source path.
func (builder *ReportBuilder) listSrcDirectories() ([]Directory, error) {
	dirEntries, err := os.ReadDir(builder.SrcPath)
	if err != nil {
		return nil, err
	}

	destDirectories, err := builder.getSrcDirectories()
	if err != nil {
		return nil, err
	}

	directories := make([]Directory, 0)
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			directories = append(directories, Directory{FoundInDest: destDirectories[dirEntry.Name()], Path: dirEntry.Name()})
		}
	}
	return directories, nil
}

func (builder *ReportBuilder) getSrcDirectories() (map[string]bool, error) {
	dirEntries, err := os.ReadDir(builder.DstPath)
	if err != nil {
		return nil, err
	}

	directories := make(map[string]bool)
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			directories[dirEntry.Name()] = true
		}
	}
	return directories, nil
}

// writeIndexTemplate generates and writes the index template with the given output paths.
func (builder *ReportBuilder) writeIndexTemplate(data *IndexTemplate) error {
	resolvedTemplate := new(bytes.Buffer)
	if err := builder.indexTemplate.Execute(resolvedTemplate, data); err != nil {
		return err
	}

	if err := os.WriteFile(builder.OutDir+"/index.html", resolvedTemplate.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// writePackageTemplate executes the template with the provided data and
// writes the generated report to the output directory.
func (builder *ReportBuilder) writePackageTemplate(templateData any, packageName string) error {
	resolvedTemplate := new(bytes.Buffer)
	if err := builder.packageTemplate.Execute(resolvedTemplate, templateData); err != nil {
		return err
	}

	if err := os.MkdirAll(builder.OutDir+"/"+packageName, 0777); err != nil {
		return err
	}

	if err := os.WriteFile(builder.OutDir+"/"+packageName+"/report.html", resolvedTemplate.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
