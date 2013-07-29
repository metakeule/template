package template

import (
	"fmt"
	"github.com/metakeule/fastreplace"
	"github.com/metakeule/meta"
	"io"
	"strings"
)

type Template struct {
	fr     *fastreplace.FReplace
	Strict bool // panics if not all placeholders are set
}

type Replacer interface {
	Key() string   // the name of the placeholder, should be unique, no @@ should be inside the name
	Value() string // value might use escaping before returning the value
}

type Placeholder struct {
	name, val   string
	Transformer func(interface{}) string
}

func NewPlaceholder(name string) Placeholder { return Placeholder{name: name} }

func (ø Placeholder) Value() string  { return ø.val }
func (ø Placeholder) String() string { return "@@" + ø.name + "@@" }
func (ø Placeholder) Key() string    { return ø.name }
func (ø Placeholder) Setf(format string, val ...interface{}) Placeholder {
	return ø.Set(fmt.Sprintf(format, val...))
}
func (ø Placeholder) Set(val interface{}) Placeholder {
	if ø.Transformer != nil {
		return Placeholder{name: ø.name, val: ø.Transformer(val), Transformer: ø.Transformer}
	}
	return Placeholder{name: ø.name, val: fmt.Sprintf("%v", val)}
}

// set the value method on the pointer, to ensure, that the result of set was given to
// template.Replace

// the template string must have @@placeholdername@@ as placeholders
// and must not have two placeholders directly adjacent, e.g. @@ph1@@@@ph2@@
func New(templ string) (ø *Template, err error) {
	fr, e := fastreplace.NewString("@@", templ)
	if e != nil {
		err = e
		return
	}
	ø = &Template{fr: fr}
	return
}

func (ø *Template) Replace(replacements ...Replacer) Instancer {
	return ø.New().Replace(replacements...)
}

func (ø *Template) Merge(src interface{}, key string, placeholders interface{}) Instancer {
	return ø.New().Merge(src, key, placeholders)
}

type Instancer interface {
	Replace(replacements ...Replacer) Instancer
	String() string
	Bytes() []byte
	WriteTo(w io.Writer)
	Merge(src interface{}, key string, placeholders interface{}) Instancer
}

type Instance struct {
	inst fastreplace.Replacer
}

//ErrorOnMissingPlaceholders
type InstanceStrict struct {
	*Instance
	MissingPlaceholders map[string]bool
}

func (ø *InstanceStrict) String() string {
	// fmt.Println(ø.MissingPlaceholders)
	if len(ø.MissingPlaceholders) > 0 {
		s := []string{}
		for k, _ := range ø.MissingPlaceholders {
			s = append(s, k)
		}
		panic("Missing placeholders: " + strings.Join(s, ", "))
	}
	return ø.Instance.String()
}

func (ø *InstanceStrict) Replace(replacements ...Replacer) Instancer {
	for _, r := range replacements {
		delete(ø.MissingPlaceholders, r.Key())
		ø.inst.AssignString(r.Key(), r.Value())
	}
	return ø
}

func (ø *InstanceStrict) Merge(src interface{}, key string, placeholders interface{}) Instancer {
	ø.Instance.Merge(src, key, placeholders)
	return ø
}

func (ø *Instance) Replace(replacements ...Replacer) Instancer {
	for _, r := range replacements {
		ø.inst.AssignString(r.Key(), r.Value())
	}
	return ø
}

func (ø *Instance) String() string {
	return ø.inst.String()
}
func (ø *Instance) Bytes() []byte { return ø.inst.Bytes() }

func (ø *Instance) WriteTo(w io.Writer) {
	w.Write(ø.Bytes())
}

func (ø *Instance) Merge(src interface{}, key string, placeholders interface{}) Instancer {
	vals := StructVals(key, src)
	for field, v := range vals {
		var ph Placeholder
		meta.Struct.Get(placeholders, field, &ph)
		if ph.Key() != "" {
			ø.Replace(ph.Set(v))
		}
	}
	return ø
}

func (ø *Template) New() Instancer {
	inst := ø.fr.Instance()
	i := &Instance{inst}
	if ø.Strict {
		// fmt.Println("inst placeholders ", inst.Placeholders())
		is := &InstanceStrict{i, inst.Placeholders()}
		return is
	}
	return i
}

func StructVals(key string, structOrPointerToStruct interface{}) (r map[string]string) {
	r = map[string]string{}
	fv := meta.Struct.FinalValue(structOrPointerToStruct)
	tags := meta.Struct.Tags(structOrPointerToStruct)
	for field, v := range tags {
		if t := v.Get(key); t != "" {
			fvl := fv.FieldByName(field).Interface()
			if meta.IsDefault(fvl) {
				if t != "-" { // "-" signals no default value
					r[field] = t
				}
			} else {
				r[field] = fmt.Sprintf("%v", fvl)
			}
		}
	}
	return
}

func FillStruct(key string, transformer map[string]func(interface{}) string, ptrStruct interface{}) (notHandled map[string]string) {
	notHandled = map[string]string{}
	fn := func(field string, v interface{}) {
		tag := meta.Struct.Tag(ptrStruct, field)
		if tmpl := tag.Get(key); tmpl != "" {
			fn, ok := transformer[tmpl]
			if !ok {
				notHandled[field] = tmpl
				return
			}
			ph := NewPlaceholder(fmt.Sprintf("%T.%s", ptrStruct, field))
			ph.Transformer = fn
			meta.Struct.Set(ptrStruct, field, ph)
		}
	}
	meta.Struct.Each(ptrStruct, fn)
	return
}

func MustFillStruct(key string, transformer map[string]func(interface{}) string, ptrStruct interface{}) {
	unhandled := FillStruct(key, transformer, ptrStruct)
	if len(unhandled) > 0 {
		for k, v := range unhandled {
			panic(fmt.Sprintf("could not handle key %s with tag %s: no transformer\n", k, v))
		}
	}
}
