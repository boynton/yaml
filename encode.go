package yaml

import (
	"bytes"
	"fmt"
	"strconv"
)

func Encode(obj map[string]interface{}) (string, error) {
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
		buf.WriteString(strconv.FormatInt(o.(int64), 10))
		buf.WriteString("\n")
	case float32, float64:
		buf.WriteString(firstIndent)
		buf.WriteString(strconv.FormatFloat(o.(float64), 'g', -1, 64))
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
			switch val.(type) {
			case map[string]interface{}, []interface{}:
				buf.WriteString(indent + s + ":\n")
			default:
				buf.WriteString(indent + s + ": ")
			}
			encode(buf, val, indent2, false)
			indent = nonFirstIndent
		}
	case []interface{}:
		indent2 := indent + "  "
		for _, item := range o {
			encode(buf, item, indent2, true)
		}
	default:
		buf.WriteString(fmt.Sprintf("Whoops: %v", o))
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
