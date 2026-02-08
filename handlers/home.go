package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

var (
	tmpl     *template.Template
	tmplOnce sync.Once
)

// HomeHandler renders the landing page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Parse templates once
	tmplOnce.Do(func() {
		tmplPath := filepath.Join("templates", "index.html")
		var err error
		tmpl, err = template.ParseFiles(tmplPath)
		if err != nil {
			log.Printf("CRITICAL: Error parsing template: %v", err)
			// In a real app, we might want to panic or handle this better,
			// but for now logging is sufficient as the next check handles nil tmpl
		}
	})

	if tmpl == nil {
		http.Error(w, "Internal Server Error (Template)", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
