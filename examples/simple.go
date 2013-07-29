package main

import (
	"fmt"
	"github.com/metakeule/goh4"
	"github.com/metakeule/goh4/tag"
	// "github.com/metakeule/meta"
	"github.com/metakeule/template"
)

func textEscaper(in interface{}) (out string) {
	inString := ""
	switch v := in.(type) {
	case goh4.Stringer:
		inString = v.String()
	case string:
		inString = v
	default:
		inString = fmt.Sprintf("%v", v)
	}
	return tag.Doc(inString).String()
}

func htmlEscaper(in interface{}) (out string) {
	switch v := in.(type) {
	case goh4.Stringer:
		return v.String()
	case string:
		return v
	default:
		panic("unsupported type: " + fmt.Sprintf("%v (%T)", v, v))
	}
}

func Html(name string) (t template.Placeholder) {
	t = template.NewPlaceholder(name)
	t.Transformer = htmlEscaper
	return
}
func Text(name string) (t template.Placeholder) {
	t = template.NewPlaceholder(name)
	t.Transformer = textEscaper
	return
}

func main() {
	person := Text("person")
	greeting := Html("greeting")

	t, _ := template.New(tag.Doc(
		tag.H1("Hi, ", person),
		greeting,
	).String())

	fmt.Println(
		t.New().Replace(
			person.Set("Bugs <Bunny>"),
			greeting.Set(tag.P("How are you?").String()),
		))
}
