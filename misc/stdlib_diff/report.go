package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	PathToReport   string
	PackageName    string
	MissingGo      bool
	MissingGno     bool
	Subdirectories []LinkToReport
}

// NewReportBuilder creates a new ReportBuilder instance with the specified
// source path, destination path, and output directory. It also initializes
// the packageTemplate using the provided HTML template file.
func NewReportBuilder(srcPath, dstPath, outDir string) (*ReportBuilder, error) {
	packageTemplate, err := template.New("").Parse(packageDiffTemplate)
	if err != nil {
		return nil, err
	}

	indexTemplate, err := template.New("").Parse(indexTemplate)
	if err != nil {
		return nil, err
	}

	// filepath.EvalSymlinks will return the original path if there are no simlinks associated to the given path
	realSrcPath, err := filepath.EvalSymlinks(srcPath)
	if err != nil {
		return nil, err
	}

	realDstPath, err := filepath.EvalSymlinks(dstPath)
	if err != nil {
		return nil, err
	}

	realOutPath, err := filepath.EvalSymlinks(outDir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		// Create output if not exist
		err = os.MkdirAll(outDir, 0o777)
		if err != nil {
			return nil, err
		}
		realOutPath = outDir
	}
	return &ReportBuilder{
		// Trim suffix / in order to standardize paths accept path with or without `/`
		SrcPath:         strings.TrimSuffix(realSrcPath, `/`),
		DstPath:         strings.TrimSuffix(realDstPath, `/`),
		OutDir:          strings.TrimSuffix(realOutPath, `/`),
		packageTemplate: packageTemplate,
		indexTemplate:   indexTemplate,
	}, nil
}

// Build generates reports for differences between packages in the source and
// destination directories. It iterates through each directory, calculates
// differences using PackageDiffChecker, and generates reports using the
// packageTemplate.
func (builder *ReportBuilder) Build() error {
	directories, err := builder.listDirectories()
	if err != nil {
		return err
	}

	indexTemplateData := &IndexTemplate{
		Reports: make([]LinkToReport, 0),
	}

	for _, directory := range directories {
		if err := builder.ExecuteDiffTemplate(directory); err != nil {
			return err
		}
		report := LinkToReport{
			PathToReport:   "./" + directory.Path + "/report.html",
			PackageName:    directory.Path,
			MissingGno:     !directory.FoundInDest,
			MissingGo:      !directory.FoundInSrc,
			Subdirectories: make([]LinkToReport, 0),
		}
		for _, subDirectory := range directory.Children {
			if err := builder.ExecuteDiffTemplate(subDirectory); err != nil {
				return err
			}
			report.Subdirectories = append(report.Subdirectories, LinkToReport{
				PathToReport: "./" + subDirectory.Path + "/report.html",
				PackageName:  subDirectory.Path,
				MissingGno:   !subDirectory.FoundInDest,
				MissingGo:    !subDirectory.FoundInSrc,
			})

		}
		indexTemplateData.Reports = append(indexTemplateData.Reports, report)

	}

	if err := builder.writeIndexTemplate(indexTemplateData); err != nil {
		return err
	}

	return nil
}

func (builder *ReportBuilder) ExecuteDiffTemplate(directory *Directory) error {
	if !directory.FoundInDest {
		return nil
	}

	srcPackagePath := builder.SrcPath + "/" + directory.Path
	dstPackagePath := builder.DstPath + "/" + directory.Path
	packageChecker, err := NewPackageDiffChecker(srcPackagePath, dstPackagePath)
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

	return builder.writePackageTemplate(data, directory.Path)
}

type Directory struct {
	Path        string
	FoundInDest bool
	FoundInSrc  bool
	Children    []*Directory
}

// listDirectories retrieves a list of directories in the source path.
func (builder *ReportBuilder) listDirectories() ([]*Directory, error) {
	allSubdirectories, srcDirectories, destDirectories, err := builder.findDirectories()
	if err != nil {
		return nil, err
	}

	notfound := []string{}
	directories := make(map[string]*Directory)
	res := make([]*Directory, 0)

	for _, folderName := range allSubdirectories {
		if slices.ContainsFunc(notfound, func(s string) bool {
			return strings.HasPrefix(folderName, s)
		}) {
			// this directory is not found in either source or destination skipping subsdirectories
			continue
		}

		newDir := &Directory{
			Path:        folderName,
			FoundInDest: destDirectories[folderName],
			FoundInSrc:  srcDirectories[folderName],
			Children:    make([]*Directory, 0),
		}

		if isRootFolder(folderName) {
			directories[folderName] = newDir
			res = append(res, newDir)
		} else {
			directory := directories[getRootFolder(folderName)]
			directory.Children = append(directory.Children, newDir)
			directories[getRootFolder(folderName)] = directory
		}

		if !newDir.FoundInDest && !newDir.FoundInSrc {
			notfound = append(notfound, folderName)
		}
	}

	return res, err
}

func isRootFolder(path string) bool {
	return !strings.Contains(path, "/")
}

func getRootFolder(path string) string {
	return strings.Split(path, "/")[0]
}

func (builder *ReportBuilder) getAllSubdirectories(rootPath string) ([]string, error) {
	directories := make([]string, 0)
	err := filepath.WalkDir(rootPath, func(path string, dirEntry fs.DirEntry, err error) error {
		if path == rootPath {
			return nil
		}

		if dirEntry.IsDir() {
			folderName := strings.TrimPrefix(path, rootPath+"/")
			directories = append(directories, folderName)
		}
		return nil
	})
	return directories, err
}

// writeIndexTemplate generates and writes the index template with the given output paths.
func (builder *ReportBuilder) writeIndexTemplate(data *IndexTemplate) error {
	resolvedTemplate := new(bytes.Buffer)
	if err := builder.indexTemplate.Execute(resolvedTemplate, data); err != nil {
		return err
	}

	if err := os.WriteFile(builder.OutDir+"/index.html", resolvedTemplate.Bytes(), 0o644); err != nil {
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

	if err := os.MkdirAll(builder.OutDir+"/"+packageName, 0o777); err != nil {
		return err
	}

	if err := os.WriteFile(builder.OutDir+"/"+packageName+"/report.html", resolvedTemplate.Bytes(), 0o644); err != nil {
		return err
	}

	return nil
}

func (builder *ReportBuilder) findDirectories() ([]string, map[string]bool, map[string]bool, error) {
	destDirectories, err := builder.getAllSubdirectories(builder.DstPath)
	if err != nil {
		return nil, nil, nil, err
	}

	srcDirectories, err := builder.getAllSubdirectories(builder.SrcPath)
	if err != nil {
		return nil, nil, nil, err
	}

	res := make([]string, 0, len(srcDirectories)+len(destDirectories))
	srcMap := make(map[string]bool)
	dstMap := make(map[string]bool)
	for _, path := range srcDirectories {
		res = append(res, path)
		srcMap[path] = true
	}

	for _, path := range destDirectories {
		dstMap[path] = true
		if !srcMap[path] {
			res = append(res, path)
		}
	}

	slices.Sort(res)

	return res, srcMap, dstMap, nil
}
