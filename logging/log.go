package logevent

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"text/template"
)
var std  = log.New(os.Stderr, "", log.LstdFlags)

func Event(name string, data interface{}) error {
	tmpl := template.Must(
		template.ParseGlob("events/**/*.tmpl")
	)
	if (tmpl.Lookup(name + ":off") != nil){
		return nil
	}
	eventTmpl := tmpl.Lookup(name)
	if (eventTmpl != nil){
		log.Printf("- Expected %s template to exist, please define it with {{define \"%s\"}}{{end}}", name, name)
		log.Printf(debug.Stack())
		return nil
	}
	return eventTmpl.Execute(std, data)
}
