package main

import (
	"fmt"
	"github.com/metakeule/goh4"
	. "github.com/metakeule/goh4/attr"
	. "github.com/metakeule/goh4/tag"
	"github.com/metakeule/template"
	"strings"
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
	return Doc(inString).String()
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

func escapeDB(s string) string {
	in := strings.Replace(s, `$`, ``, -1)
	return "$i_n_p_u_t$" + in + "$i_n_p_u_t$"
}

func searchEscaperStart(in interface{}) (out string) {
	var inString string
	switch v := in.(type) {
	case goh4.Stringer:
		inString = v.String()
	case string:
		inString = v
	default:
		panic("unsupported type: " + fmt.Sprintf("%v (%T)", v, v))
	}
	return escapeDB("%" + inString)
}

func searchEscaperEnd(in interface{}) (out string) {
	var inString string
	switch v := in.(type) {
	case goh4.Stringer:
		inString = v.String()
	case string:
		inString = v
	default:
		panic("unsupported type: " + fmt.Sprintf("%v (%T)", v, v))
	}
	return escapeDB(inString + "%")
}

var dbTransformer = map[string]func(interface{}) string{
	"searchtext%": searchEscaperStart,
	"%searchtext": searchEscaperEnd,
}

var htmlTransformer = map[string]func(interface{}) string{
	"text": textEscaper,
	"html": htmlEscaper,
}

type Figure struct {
	FirstName string        `greet:"-"`
	LastName  string        `greet:"-"`
	Greeting  goh4.Stringer `greet:"-"`
	Width     int           `greet:"400"`
}

type figure struct {
	FirstName template.Placeholder `html-template:"text" db-template:"searchtext%"`
	LastName  template.Placeholder `html-template:"text" db-template:"searchtext%"`
	Width     template.Placeholder `html-template:"text"`
	Greeting  template.Placeholder `html-template:"html"`
}

var FigureHtml = &figure{}
var _ = template.FillStruct("html-template", htmlTransformer, FigureHtml)

var FigureDB = &figure{}
var _ = template.FillStruct("db-template", dbTransformer, FigureDB)

var PAGE, _ = template.New(
	DIV(
		H1(FigureHtml.FirstName),
		H2(FigureHtml.LastName, Width(FigureHtml.Width.String()+"px")),
		FigureHtml.Greeting,
	).String())

var QUERY, _ = template.New(`SELECT * FROM figure WHERE lastname LIKE ` + FigureDB.LastName.String())

func main() {
	bugs := Figure{
		FirstName: "Bugs",
		LastName:  "<Bunny>",
		Greeting:  P("How are you?"),
		Width:     200,
	}

	donald := Figure{
		FirstName: "Donald",
		LastName:  "<Duck>",
		Greeting:  P("Are you fine?"),
	}

	fmt.Println(PAGE.New().Merge(bugs, "greet", FigureHtml))
	fmt.Println(QUERY.New().Merge(bugs, "greet", FigureDB))

	fmt.Println(PAGE.New().Merge(donald, "greet", FigureHtml))
	fmt.Println(QUERY.New().Merge(donald, "greet", FigureDB))
}
