package main

import (
	"fmt"
	"github.com/metakeule/goh4"
	"github.com/metakeule/goh4/tag"
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

func Merge(ø interface{}, key string, t *template.Instance) *template.Instance {
	textVals := template.StructVals(key+".text", ø)
	for k, v := range textVals {
		t.Replace(Text(k).Set(v))
	}

	htmlVals := template.StructVals(key+".html", ø)
	for k, v := range htmlVals {
		t.Replace(Html(k).Set(v))
	}
	return t
}

type Person struct {
	FirstName string `greet.text:"-"`
	LastName  string `greet.text:"-"`
	Greeting  string `greet.html:"-"`
	Width     int    `greet.text:"400"`
}

func main() {
	//person := Text("person")
	//greeting := Html("greeting")

	t, _ := template.New(
		tag.Doc(
			tag.H1("Hi, ", "@@FirstName@@", " ", "@@LastName@@",
				tag.ATTR("width", "@@Width@@px")),
			"@@Greeting@@",
		).String(),
	)

	fmt.Println(
		Merge(
			Person{
				FirstName: "Bugs",
				LastName:  "<Bunny>",
				Greeting:  tag.P("How are you?").String(),
			},
			"greet",
			t.New()))
}
