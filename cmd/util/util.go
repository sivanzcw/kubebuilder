/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gobuffalo/flect"
)

// WriteIfNotFound returns true if the file was created and false if it already exists
func WriteIfNotFound(path, templateName, templateValue string, data interface{}) bool {
	// Make sure the directory exists
	os.MkdirAll(filepath.Dir(path), 0700)

	// Don't create the doc.go if it exists
	if _, err := os.Stat(path); err == nil {
		return false
	} else if !os.IsNotExist(err) {
		log.Fatalf("Could not stat %s: %v", path, err)
	}

	return Write(path, templateName, templateValue, data)
}

func Write(path, templateName, templateValue string, data interface{}) bool {
	t := template.Must(template.New(templateName).Funcs(
		template.FuncMap{
			"title":  strings.Title,
			"lower":  strings.ToLower,
			"plural": flect.Pluralize,
		},
	).Parse(templateValue))

	var tmp bytes.Buffer
	err := t.Execute(&tmp, data)
	if err != nil {
		log.Fatalf("Failed to render template %s: %v", templateName, err)
	}

	content := tmp.Bytes()
	if filepath.Ext(path) == ".go" {
		content, err = format.Source(content)
		if err != nil {
			log.Fatalf("Failed to format template %s: %v", templateName, err)
		}
	}

	WriteString(path, string(content))

	return true
}

func WriteString(path, value string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		create(path)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", path, err)
	}
	defer f.Close()

	_, err = f.WriteString(value)
	if err != nil {
		log.Fatalf("Failed to write %s: %v", path, err)
	}
}

// GetCopyright will return the contents of the copyright file if it exists.
// if the file cannot be read, will return the empty string.
func GetCopyright(file string) string {
	if len(file) == 0 {
		file = filepath.Join("hack", "boilerplate.go.txt")
	}
	cr, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(cr)
}

func create(path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
}

func DoCmd(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	log.Printf("%s\n", strings.Join(c.Args, " "))
	err := c.Run()
	if err != nil {
		log.Fatalf("command failed %v", err)
	}
}

func IsNewVersion() bool {
	_, err := os.Stat("PROJECT")
	if err != nil {
		return false
	}

	return true
}

func ProjectExist() bool {
	return IsNewVersion()
}

func IsProjectNotInitialized() bool {
	dirs := []string{
		"cmd",
		"hack",
		"pkg",
		"vendor",
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			return false
		}
	}
	return true
}
