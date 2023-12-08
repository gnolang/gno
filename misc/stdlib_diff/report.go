package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
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
	GnoFileCount       int              // Number of Gno files in the package.
	GnoPackageLocation string           // Location of Gno files in the source directory.
	GoFileCount        int              // Number of Go files in the package.
	GoPackageLocation  string           // Location of Go files in the destination directory.
	FilesDifferences   []FileDifference // Differences in individual files.
}

type IndexTemplate struct {
	Reports []LinkToReport
}

type LinkToReport struct {
	PathToReport string
	PackageName  string
}

// NewReportBuilder creates a new ReportBuilder instance with the specified
// source path, destination path, and output directory. It also initializes
// the packageTemplate using the provided HTML template file.
func NewReportBuilder(srcPath, dstPath, outDir string, srcIsGno bool) (*ReportBuilder, error) {
	packageTemplate, err := template.ParseFiles("templates/package_diff_template.html")
	if err != nil {
		return nil, err
	}

	indexTemplate, err := template.ParseFiles("templates/index_template.html")
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
		srcPackagePath := builder.SrcPath + "/" + directory
		dstPackagePath := builder.DstPath + "/" + directory

		packageChecker, err := NewPackageDiffChecker(srcPackagePath, dstPackagePath, builder.SrcIsGno)
		if err != nil {
			return fmt.Errorf("can't create new PackageDiffChecker: %w", err)
		}

		differences, err := packageChecker.Differences()
		if err != nil {
			return fmt.Errorf("can't compute differences: %w", err)
		}

		data := &PackageDiffTemplateData{
			PackageName:        directory,
			GnoFileCount:       len(packageChecker.SrcFiles),
			GnoPackageLocation: srcPackagePath,
			GoFileCount:        len(packageChecker.DstFiles),
			GoPackageLocation:  dstPackagePath,
			FilesDifferences:   differences.FilesDifferences,
		}

		if err := builder.writePackageTemplate(data, directory); err != nil {
			return err
		}

		indexTemplateData.Reports = append(indexTemplateData.Reports, LinkToReport{
			PathToReport: "./" + directory + "/report.html",
			PackageName:  directory,
		})
	}

	if err := builder.writeIndexTemplate(indexTemplateData); err != nil {
		return err
	}

	return nil
}

// listSrcDirectories retrieves a list of directories in the source path.
func (builder *ReportBuilder) listSrcDirectories() ([]string, error) {
	dirEntries, err := os.ReadDir(builder.SrcPath)
	if err != nil {
		return nil, err
	}

	directories := make([]string, 0)
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			directories = append(directories, dirEntry.Name())
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
