// language_catalog.gno — Mapeamento de extensões para linguagens
package compiler

import (
	"strings"
)

// ExtensionEntry define uma entrada no catálogo de extensões
type ExtensionEntry struct {
    Extensions []string // ex: [".py", ".pyw", ".pyx"]
    Language   string   // ex: "python"
    Family     string   // ex: "scripting", "systems", "web", "data", "config"
    Shebang    []string // ex: ["python", "python3"]
    Keywords   []string // palavras-chave únicas para detecção
    MimeType   string   // ex: "text/x-python"
}

// LanguageCatalog contém todas as extensões conhecidas
type LanguageCatalog struct {
    entries      map[string]*ExtensionEntry // extensão → entrada
    langIndex    map[string]*ExtensionEntry // linguagem → entrada
    mu           int
    extensionsCount int
    languagesCount  int
}

func NewLanguageCatalog() *LanguageCatalog {
    lc := &LanguageCatalog{
        entries: make(map[string]*ExtensionEntry),
        langIndex: make(map[string]*ExtensionEntry),
    }
    lc.populateCatalog()
    return lc
}

func (lc *LanguageCatalog) populateCatalog() {
    catalog := []*ExtensionEntry{
        // ===== LINGUAGENS DE SISTEMA =====
        {[]string{".c"}, "c", "systems", []string{}, []string{"#include", "int main", "void"}, "text/x-c"},
        {[]string{".cpp", ".cc", ".cxx", ".c++", ".C"}, "cpp", "systems", []string{}, []string{"#include", "std::", "cout", "class"}, "text/x-c++"},
        {[]string{".h", ".hpp", ".hxx"}, "c_header", "systems", []string{}, []string{"#ifndef", "#define", "typedef"}, "text/x-chdr"},
        {[]string{".rs"}, "rust", "systems", []string{}, []string{"fn ", "let ", "impl ", "struct "}, "text/x-rust"},
        {[]string{".go"}, "go", "systems", []string{}, []string{"package ", "func ", "go ", "defer "}, "text/x-go"},
        {[]string{".zig"}, "zig", "systems", []string{}, []string{"const ", "fn ", "pub "}, "text/x-zig"},
        {[]string{".nim"}, "nim", "systems", []string{}, []string{"proc ", "import ", "echo "}, "text/x-nim"},
        {[]string{".d"}, "d", "systems", []string{}, []string{"import ", "void ", "class "}, "text/x-d"},
        {[]string{".swift"}, "swift", "systems", []string{"swift"}, []string{"import ", "var ", "func ", "struct "}, "text/x-swift"},
        {[]string{".kt", ".kts"}, "kotlin", "systems", []string{}, []string{"fun ", "val ", "var ", "class "}, "text/x-kotlin"},
        {[]string{".java"}, "java", "systems", []string{}, []string{"public class", "import java", "System.out"}, "text/x-java"},
        {[]string{".scala", ".sc"}, "scala", "systems", []string{}, []string{"def ", "val ", "object ", "trait "}, "text/x-scala"},
        {[]string{".cs"}, "csharp", "systems", []string{}, []string{"using ", "namespace ", "class ", "void "}, "text/x-csharp"},
        {[]string{".fs", ".fsx"}, "fsharp", "systems", []string{}, []string{"let ", "module ", "type ", "match "}, "text/x-fsharp"},
        {[]string{".vb"}, "visualbasic", "systems", []string{}, []string{"Sub ", "Function ", "Dim ", "Module "}, "text/x-vb"},
        {[]string{".m", ".mm"}, "objectivec", "systems", []string{}, []string{"#import", "@interface", "@implementation"}, "text/x-objectivec"},

        // ===== SCRIPTING =====
        {[]string{".py", ".pyw", ".pyx", ".pxd", ".pyi"}, "python", "scripting", []string{"python", "python3", "python2"}, []string{"def ", "class ", "import ", "from "}, "text/x-python"},
        {[]string{".rb", ".rbw"}, "ruby", "scripting", []string{"ruby"}, []string{"def ", "class ", "require ", "end"}, "text/x-ruby"},
        {[]string{".pl", ".pm", ".t"}, "perl", "scripting", []string{"perl"}, []string{"sub ", "my ", "use ", "package "}, "text/x-perl"},
        {[]string{".php", ".phtml", ".php3", ".php4", ".php5", ".php7", ".php8"}, "php", "scripting", []string{}, []string{"<?php", "namespace ", "function ", "class "}, "text/x-php"},
        {[]string{".lua"}, "lua", "scripting", []string{"lua"}, []string{"function ", "local ", "require ", "end"}, "text/x-lua"},
        {[]string{".tcl"}, "tcl", "scripting", []string{"tclsh", "wish"}, []string{"proc ", "set ", "puts ", "package "}, "text/x-tcl"},
        {[]string{".r", ".R", ".Rprofile", ".Rhistory"}, "r", "scripting", []string{"Rscript"}, []string{"library(", "ggplot(", "data.frame", "<-"}, "text/x-r"},
        {[]string{".jl"}, "julia", "scripting", []string{"julia"}, []string{"function ", "struct ", "module ", "using "}, "text/x-julia"},
        {[]string{".groovy"}, "groovy", "scripting", []string{}, []string{"def ", "class ", "println ", "import "}, "text/x-groovy"},
        {[]string{".dart"}, "dart", "scripting", []string{"dart"}, []string{"import ", "class ", "void ", "final "}, "text/x-dart"},
        {[]string{".ex", ".exs"}, "elixir", "scripting", []string{"elixir"}, []string{"def ", "defmodule ", "do", "end"}, "text/x-elixir"},
        {[]string{".erl", ".hrl"}, "erlang", "scripting", []string{"escript"}, []string{"-module(", "-export(", "fun ", "end"}, "text/x-erlang"},
        {[]string{".clj", ".cljs", ".cljc", ".edn"}, "clojure", "scripting", []string{}, []string{"(defn ", "(def ", "(ns ", "(let "}, "text/x-clojure"},
        {[]string{".hs", ".lhs"}, "haskell", "scripting", []string{"runhaskell", "runghc"}, []string{"module ", "import ", "data ", "class "}, "text/x-haskell"},
        {[]string{".ml", ".mli"}, "ocaml", "scripting", []string{}, []string{"let ", "type ", "module ", "match "}, "text/x-ocaml"},
        {[]string{".fnl"}, "fennel", "scripting", []string{"fennel"}, []string{"(fn ", "(local ", "(let ", "(var "}, "text/x-fennel"},

        // ===== WEB =====
        {[]string{".js", ".mjs", ".cjs"}, "javascript", "web", []string{"node", "nodejs"}, []string{"function ", "const ", "let ", "var ", "=>"}, "text/javascript"},
        {[]string{".ts", ".tsx"}, "typescript", "web", []string{"ts-node", "tsx"}, []string{"interface ", "type ", "const ", "import "}, "text/typescript"},
        {[]string{".jsx"}, "jsx", "web", []string{}, []string{"import React", "<>", "export default"}, "text/jsx"},
        {[]string{".html", ".htm", ".xhtml"}, "html", "web", []string{}, []string{"<!DOCTYPE", "<html", "<head", "<body"}, "text/html"},
        {[]string{".css", ".scss", ".sass"}, "css", "web", []string{}, []string{"@import", "font-size", "margin:", "padding:"}, "text/css"},
        {[]string{".less"}, "less", "web", []string{}, []string{"@import", "@variable", "margin:", "padding:"}, "text/less"},
        {[]string{".vue"}, "vue", "web", []string{}, []string{"<template>", "<script>", "export default"}, "text/x-vue"},
        {[]string{".svelte"}, "svelte", "web", []string{}, []string{"<script>", "<style>", "export let"}, "text/x-svelte"},
        {[]string{".coffee"}, "coffeescript", "web", []string{"coffee"}, []string{"->", "=>", "class ", "extends"}, "text/x-coffeescript"},
        {[]string{".wasm"}, "wasm", "web", []string{}, []string{}, "application/wasm"},

        // ===== DADOS E CONFIGURAÇÃO =====
        {[]string{".json"}, "json", "data", []string{}, []string{}, "application/json"},
        {[]string{".yaml", ".yml"}, "yaml", "config", []string{}, []string{"apiVersion:", "kind:", "metadata:", "spec:"}, "text/yaml"},
        {[]string{".toml"}, "toml", "config", []string{}, []string{"[package]", "[dependencies]", "[build]"}, "text/toml"},
        {[]string{".xml", ".xsl", ".xsd", ".wsdl"}, "xml", "data", []string{}, []string{"<?xml", "<root", "<element", "xmlns"}, "text/xml"},
        {[]string{".csv", ".tsv"}, "csv", "data", []string{}, []string{}, "text/csv"},
        {[]string{".ini", ".cfg", ".conf", ".config"}, "ini", "config", []string{}, []string{"[section]", "key=", "key:"}, "text/x-ini"},
        {[]string{".env"}, "env", "config", []string{}, []string{"KEY=", "export ", "SECRET="}, "text/x-env"},
        {[]string{".properties"}, "properties", "config", []string{}, []string{"key=", "key:"}, "text/x-properties"},
        {[]string{".md", ".mdx", ".markdown"}, "markdown", "documentation", []string{}, []string{"# ", "## ", "```", "---"}, "text/markdown"},
        {[]string{".rst"}, "restructuredtext", "documentation", []string{}, []string{"===", "---", ".. ", "::"}, "text/x-rst"},
        {[]string{".tex", ".latex", ".sty", ".cls"}, "latex", "documentation", []string{}, []string{"\\documentclass", "\\begin", "\\section", "\\usepackage"}, "text/x-tex"},
        {[]string{".proto"}, "protobuf", "data", []string{}, []string{"syntax = \"proto", "message ", "service ", "rpc "}, "text/x-proto"},
        {[]string{".graphql", ".gql"}, "graphql", "data", []string{}, []string{"query ", "mutation ", "subscription ", "type "}, "text/x-graphql"},
        {[]string{".sql"}, "sql", "data", []string{}, []string{"SELECT ", "CREATE TABLE", "INSERT INTO", "ALTER TABLE"}, "text/x-sql"},
        {[]string{".sh", ".bash", ".zsh", ".ksh", ".fish"}, "shell", "scripting", []string{"bash", "sh", "zsh", "fish"}, []string{"#!/bin/", "export ", "echo ", "if ["}, "text/x-shellscript"},
        {[]string{".ps1", ".psm1", ".psd1"}, "powershell", "scripting", []string{"pwsh", "powershell"}, []string{"Write-Host", "Get-", "param(", "function "}, "text/x-powershell"},
        {[]string{".bat", ".cmd"}, "batch", "scripting", []string{}, []string{"@echo", "set ", "if ", "for "}, "text/x-batch"},
        {[]string{".dockerfile", "Dockerfile"}, "dockerfile", "config", []string{}, []string{"FROM ", "RUN ", "COPY ", "CMD "}, "text/x-dockerfile"},
        {[]string{".makefile", "Makefile", ".mk"}, "makefile", "config", []string{}, []string{"all:", "clean:", ".PHONY:", "CC="}, "text/x-makefile"},
        {[]string{".cmake", "CMakeLists.txt"}, "cmake", "config", []string{}, []string{"cmake_minimum_required", "project(", "add_executable"}, "text/x-cmake"},
        {[]string{".gradle"}, "gradle", "config", []string{}, []string{"plugins {", "dependencies {", "repositories {"}, "text/x-gradle"},
        {[]string{".pom", ".pom.xml"}, "maven", "config", []string{}, []string{"<project ", "<groupId>", "<artifactId>"}, "text/x-maven"},

        // ===== ASSEMBLY E BAIXO NÍVEL =====
        {[]string{".s", ".asm", ".nasm"}, "assembly", "systems", []string{}, []string{"mov ", "add ", "sub ", "jmp ", "call "}, "text/x-asm"},
        {[]string{".ll"}, "llvm", "systems", []string{}, []string{"define ", "declare ", "alloca", "ret "}, "text/x-llvm"},
        {[]string{".wat"}, "wast", "web", []string{}, []string{"(module", "(func", "(memory", "(export"}, "text/x-wat"},
        {[]string{".v"}, "verilog", "systems", []string{}, []string{"module ", "input ", "output ", "always"}, "text/x-verilog"},
        {[]string{".vhd", ".vhdl"}, "vhdl", "systems", []string{}, []string{"entity ", "architecture ", "port(", "signal "}, "text/x-vhdl"},
        {[]string{".sv"}, "systemverilog", "systems", []string{}, []string{"module ", "interface ", "program ", "class "}, "text/x-systemverilog"},
        {[]string{".sp"}, "spice", "systems", []string{}, []string{".SUBCKT", ".MODEL", "VIN", ".END"}, "text/x-spice"},
        {[]string{".nc"}, "netcdf", "data", []string{}, []string{}, "application/x-netcdf"},
        {[]string{".hdf", ".h5", ".hdf5"}, "hdf5", "data", []string{}, []string{}, "application/x-hdf5"},
        {[]string{".fits", ".fit"}, "fits", "data", []string{}, []string{}, "application/fits"},
        {[]string{".mat"}, "matlab_data", "data", []string{}, []string{}, "application/x-matlab-data"},
    }




    for _, entry := range catalog {
        lc.languagesCount++
        for _, ext := range entry.Extensions {
            lc.entries[strings.ToLower(ext)] = entry
            lc.extensionsCount++
        }
        lc.langIndex[entry.Language] = entry
    }
}
