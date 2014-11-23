package yaml

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

func Encode(obj interface{}) (string, error) {
	var buf bytes.Buffer
	encode(&buf, obj, "", false)
	return buf.String(), nil
}

func encode(buf *bytes.Buffer, obj interface{}, indent string, inList bool) {
	firstIndent := ""
	if inList {
		firstIndent = indent[2:] + "- "
	}
	switch o := obj.(type) {
	case string:
		buf.WriteString(firstIndent)
		buf.WriteString(encodeString(o, len(indent)))
		buf.WriteString("\n")
	case int, int32, int64:
		buf.WriteString(firstIndent)
		buf.WriteString(fmt.Sprintf("%d", o))
		buf.WriteString("\n")
	case float32, float64:
		buf.WriteString(firstIndent)
		buf.WriteString(fmt.Sprintf("%g", o))
		buf.WriteString("\n")
	case map[string]interface{}:
		indent2 := indent + "    "
		nonFirstIndent := indent
		if inList {
			nonFirstIndent = indent[2:] + "  "
			indent = firstIndent
		}
		for key, val := range o {
			s := encodeString(key, len(firstIndent))
			switch oo := val.(type) {
			case []interface{}:
				if len(oo) == 0 {
					buf.WriteString(indent + s + ": ")
				} else {
					buf.WriteString(indent + s + ":\n")
				}
			case map[string]interface{}:
				if len(oo) == 0 {
					buf.WriteString(indent + s + ": \n")
					continue
				} else {
					buf.WriteString(indent + s + ":\n")
				}
			default:
				buf.WriteString(indent + s + ": ")
			}
			encode(buf, val, indent2, false)
			indent = nonFirstIndent
		}
	case []interface{}:
		indent2 := indent + "  "
		if len(o) == 0 {
			buf.WriteString(firstIndent)
			buf.WriteString("[]\n")
		} else {
			for _, item := range o {
				encode(buf, item, indent2, true)
			}
		}
	default:
		typ := reflect.TypeOf(o)
		if typ.Kind() == reflect.Ptr{
			typ = typ.Elem()
		}
		if typ.Kind() == reflect.Struct {
			val := reflect.ValueOf(o)
			if val.Kind() == reflect.Ptr{
				val = val.Elem()
			}
			indent2 := indent + "    "
			nonFirstIndent := indent
			if inList {
				nonFirstIndent = indent[2:] + "  "
				indent = firstIndent
			}
			for i := 0; i < typ.NumField(); i++ {
				p := typ.Field(i)
				if !p.Anonymous {
					tag := p.Tag.Get("json")
					if tag != "-" { //avoid those marked to ignore
						if tag == "" {
							tag = p.Name
							//lower case preferred, but this is whan encoding/json does
						}
						optional := false
						k := strings.Index(tag, ",omitempty")
						if k > 0 {
							tag = tag[0:k]
						}
						s := encodeString(tag, len(firstIndent))
						value := val.Field(i).Interface()
						switch oo := value.(type) {
						case []interface{}:
							if len(oo) == 0 {
								buf.WriteString(indent + s + ": []")
								continue
							} else {
								buf.WriteString(indent + s + ":\n")
							}
						case map[string]interface{}:
							if len(oo) == 0 {
								buf.WriteString(indent + s + ": {}")
								continue
							} else {
								buf.WriteString(indent + s + ":\n")
							}
						default:
							typ2 := reflect.TypeOf(oo)
							if typ2.Kind() == reflect.Ptr {
								typ2 = typ2.Elem()
							}
							if typ2.Kind() == reflect.Struct {
								buf.WriteString(indent + s + ":\n")
							} else {
								buf.WriteString(indent + s + ": ")
							}
						}
						if optional && value == nil {
							//skip it
						} else {
							encode(buf, value, indent2, false)
							indent = nonFirstIndent
						}
					}
				}
			}
		} else {
			buf.WriteString(fmt.Sprintf("Whoops: %v", o))
		}
	}
}

func encodeString(s string, level int) string {
	buf := []byte{}
	needsQuote := false
	needsBlock := false
	for _, c := range s {
		switch c {
		case '\t', '\f', '\b', '\r':
			needsQuote = true
		case '\n':
			needsBlock = true
		}
	}
	needsIndent := false
	if needsQuote {
		buf = append(buf, '"')
	} else if needsBlock {
		buf = append(buf, '|')
		buf = append(buf, '\n')
		for i := 0; i < level; i++ {
			buf = append(buf, ' ')
		}
	}

	for _, c := range s {
		if needsIndent {
			needsIndent = false
			buf = append(buf, '\n')
			for i := 0; i < level; i++ {
				buf = append(buf, ' ')
			}
		}
		switch c {
		case '"':
			if needsQuote {
				buf = append(buf, '\\')
			}
			buf = append(buf, '"')
		case '\\':
			buf = append(buf, '\\')
			buf = append(buf, '\\')
		case '\n':
			if needsBlock {
				needsIndent = true
			} else {
				buf = append(buf, '\\')
				buf = append(buf, 'n')
			}
		case '\t':
			buf = append(buf, '\\')
			buf = append(buf, 't')
		case '\f':
			buf = append(buf, '\\')
			buf = append(buf, 'f')
		case '\b':
			buf = append(buf, '\\')
			buf = append(buf, 'b')
		case '\r':
			buf = append(buf, '\\')
			buf = append(buf, 'r')
			//to do: handle non-byte unicode by encoding as "\uhhhh"
		default:
			buf = append(buf, byte(c))
		}
	}
	if needsQuote {
		buf = append(buf, '"')
	}
	return string(buf)
}
