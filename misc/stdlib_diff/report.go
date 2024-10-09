package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
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
	PathToReport   string
	PackageName    string
	WasFound       bool
	Subdirectories []LinkToReport
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

	realSrcPath, err := getRealPath(srcPath)
	if err != nil {
		return nil, err
	}

	realDstPath, err := getRealPath(dstPath)
	if err != nil {
		return nil, err
	}

	realOutPath, err := getRealPath(outDir)
	if err != nil {
		return nil, err
	}
	return &ReportBuilder{
		// Trim suffix / in order to standardize paths accept path with or without `/`
		SrcPath:         strings.TrimSuffix(realSrcPath, `/`),
		DstPath:         strings.TrimSuffix(realDstPath, `/`),
		OutDir:          strings.TrimSuffix(realOutPath, `/`),
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
		if err := builder.ExecuteDiffTemplate(directory); err != nil {
			return err
		}
		report := LinkToReport{
			PathToReport:   "./" + directory.Path + "/report.html",
			PackageName:    directory.Path,
			WasFound:       directory.FoundInDest,
			Subdirectories: make([]LinkToReport, 0),
		}
		for _, subDirectory := range directory.Children {
			if err := builder.ExecuteDiffTemplate(subDirectory); err != nil {
				return err
			}
			report.Subdirectories = append(report.Subdirectories, LinkToReport{
				PathToReport: "./" + subDirectory.Path + "/report.html",
				PackageName:  subDirectory.Path,
				WasFound:     subDirectory.FoundInDest,
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

	return builder.writePackageTemplate(data, directory.Path)
}

type Directory struct {
	Path        string
	FoundInDest bool
	Children    []*Directory
}

// listSrcDirectories retrieves a list of directories in the source path.
func (builder *ReportBuilder) listSrcDirectories() ([]*Directory, error) {
	destDirectories, err := builder.getDstDirectories()
	if err != nil {
		return nil, err
	}

	notfoundInDest := []string{}
	directories := make(map[string]*Directory)
	res := make([]*Directory, 0)
	err = filepath.WalkDir(builder.SrcPath, func(path string, dirEntry fs.DirEntry, err error) error {
		if path == builder.SrcPath {
			return nil
		}

		folderName := strings.TrimPrefix(path, builder.SrcPath+"/")

		// skip directories that are not found in the destination
		for _, prefix := range notfoundInDest {
			if strings.HasPrefix(folderName, prefix) {
				return nil
			}
		}

		if err != nil {
			return err
		}

		if !dirEntry.IsDir() {
			return nil
		}

		newDir := &Directory{
			Path:        folderName,
			FoundInDest: destDirectories[folderName],
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

		if !destDirectories[dirEntry.Name()] {
			notfoundInDest = append(notfoundInDest, folderName)
		}
		return nil
	})

	return res, err
}
func isRootFolder(path string) bool {
	return !strings.Contains(path, "/")
}
func getRootFolder(path string) string {
	return strings.Split(path, "/")[0]
}
func (builder *ReportBuilder) getDstDirectories() (map[string]bool, error) {
	directories := make(map[string]bool)
	err := filepath.WalkDir(builder.DstPath, func(path string, dirEntry fs.DirEntry, err error) error {
		if dirEntry.IsDir() {
			folderName := strings.TrimPrefix(path, builder.DstPath+"/")
			directories[folderName] = true
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

// getRealPath will check if the directory is a symbolic link and resolve if path before returning it
func getRealPath(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	if info.Mode()&fs.ModeSymlink != 0 {
		// File is symbolic link, no need to resolve
		link, err := os.Readlink(path)
		if err != nil {
			return "", fmt.Errorf("can't resolve symbolic link: %w", err)
		}
		return link, nil
	}

	return path, nil
}
