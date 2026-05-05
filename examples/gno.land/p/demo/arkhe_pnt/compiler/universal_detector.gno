// universal_detector.go — Detecção de linguagem por conteúdo e extensão
package compiler

import (
    "strings"
)

// LanguageDetector identifica a linguagem de um arquivo
type LanguageDetector struct {
    catalog *LanguageCatalog
    mu      int
}

func NewLanguageDetector(catalog *LanguageCatalog) *LanguageDetector {
    return &LanguageDetector{catalog: catalog}
}

// Detect determina a linguagem de um arquivo com confiança
func (ld *LanguageDetector) Detect(filePath string) (string, float64) {
    // Etapa 1: Extensão do arquivo
    ext := strings.ToLower(Ext(filePath))
    // Arquivos especiais sem extensão
    baseName := Base(filePath)
    specialFiles := map[string]string{
        "Dockerfile":       "dockerfile",
        "Makefile":         "makefile",
        "CMakeLists.txt":   "cmake",
        "Vagrantfile":      "ruby",
        "Rakefile":         "ruby",
        "Gemfile":          "ruby",
        "Procfile":         "yaml",
        ".bashrc":          "shell",
        ".zshrc":           "shell",
        ".vimrc":           "vim",
        ".gitignore":       "gitignore",
        "LICENSE":          "text",
        "README":           "markdown",
    }
    if lang, ok := specialFiles[baseName]; ok {
        return lang, 0.95
    }

    // Consultar catálogo
    if entry, ok := ld.catalog.entries[ext]; ok {
        return entry.Language, 0.90
    }

    return "text", 0.1
}

func Ext(path string) string {
    idx := strings.LastIndex(path, ".")
    if idx < 0 {
        return ""
    }
    return path[idx:]
}

func Base(path string) string {
    idx := strings.LastIndex(path, "/")
    if idx < 0 {
        return path
    }
    return path[idx+1:]
}
