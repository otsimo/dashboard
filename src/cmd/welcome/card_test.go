package main

import (
	"bytes"
	"testing"
)

func TestTemplate(t *testing.T) {
	config.TemplatesPath = "templates"
	ReadTemplates()

	if len(templates.Templates()) == 0 {
		t.Fatalf("failed to find templates")
	}
	data := map[string]interface{}{
		"FirstName": "Hello",
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "test.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	res := b.String()
	want := "Hello is here!!"
	if res != want {
		t.Fatalf("failed to execute template, want='%s' got='%s'", want, res)
	}
}
