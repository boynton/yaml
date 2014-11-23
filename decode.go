package yaml

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

func DecodeFile(filename string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return DecodeBytes(data)
}

func DecodeBytes(in []byte) (map[string]interface{}, error) {
	return DecodeString(string(in))
}

func DecodeString(in string) (map[string]interface{}, error) {
	return Decode(strings.NewReader(in))
}

type DataDecoder struct {
	in   *bufio.Reader
	line *string
}

func Decode(in io.Reader) (map[string]interface{}, error) {
	br := bufio.NewReader(in)
	decoder := DataDecoder{br, nil}
	obj, err := decoder.decodeObject(0)
	if err != nil {
		return nil, err
	}
	switch v := obj.(type) {
	case map[string]interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("Unexpected return from decodeObject:", v)
	}
}

func (dd *DataDecoder) getLine() (string, error) {
	if dd.line != nil {
		s := *dd.line
		dd.line = nil
		return s, nil
	}
	s, err := dd.in.ReadString('\n')
	l := len(s)
	if err == io.EOF {
		if l > 0 {
			return s, nil
		}
	} else if err == nil {
		return s[0 : l-1], nil
	}
	return "", err
}

func (dd *DataDecoder) ungetLine(line string) {
	dd.line = &line
}

func stripComment(line string) string {
	i := strings.Index(line, "#")
	if i < 0 {
		return line
	}
	return line[0:i]
}

func (dd *DataDecoder) getNonEmptyLine() (string, error) {
	line, err := dd.getLine()
	for err == nil {
		line = stripComment(line)
		if len(line) > 0 {
			if line != "---" && line != "..." { //doc markers not implemented, jsut ignored
				return line, nil
			}
		}
		line, err = dd.getLine()
	}
	return "", err
}

func lineIndentLevel(line string) int {
	eol := len(line)
	for i := 0; i < eol; i++ {
		if line[i] != ' ' {
			return i
		}
	}
	return -1
}

func (dd *DataDecoder) decodeArray(level int) ([]interface{}, error) {
	obj := make([]interface{}, 0)
	//line, err := dd.getLine()
	line, err := dd.getNonEmptyLine()
	for err == nil {
		eol := len(line)
		lev := lineIndentLevel(line)
		if lev < 0 {
			return nil, fmt.Errorf("Badly formed array entry: " + line)
		}
		if lev < level {
			dd.ungetLine(line)
			break
		}
		if line[lev] != '-' {
			return nil, fmt.Errorf("Badly formed array entry: " + line)
		}
		itemStart := -1
		for i := lev + 1; i < eol; i++ {
			if line[i] != ' ' {
				itemStart = i
				break
			}
		}
		if itemStart < 0 {
			return nil, fmt.Errorf("Badly formed array entry: " + line)
		}
		val := line[itemStart:]
		if strings.Index(val, ":") < 0 {
			//no key, treat as a scalar
			obj = append(obj, val)
		} else {
			cleanLine := line[0:lev] + " " + line[lev+1:]
			dd.ungetLine(cleanLine)
			item, err2 := dd.decodeObject(itemStart)
			if err2 == io.EOF {
				obj = append(obj, item)
				return obj, nil
			} else if err2 != nil {
				return nil, err2
			}
			obj = append(obj, item)
		}
		line, err = dd.getNonEmptyLine()
	}
	if err == io.EOF {
		return obj, nil
	} else {
		return obj, err
	}
}

func (dd *DataDecoder) decodeObject(level int) (interface{}, error) {
	obj := make(map[string]interface{}, 0)
	line, err := dd.getNonEmptyLine()
	for err == nil {
		eol := len(line)
		lev := lineIndentLevel(line) //index of first nonspace char in line
		if lev == level {
			//a new collection. If it starts with '-', then it is an array, else object
			if line[lev] == '-' {
				dd.ungetLine(line)
				item, err2 := dd.decodeArray(lev)
				if err2 == io.EOF {
					return item, nil
				}
				return item, err2
			}
			//else an object
			var i, j, k int
			key := ""
			if line[lev] == '"' {
				//quoted key
				j = strings.Index(line[lev+1:], "\"")
				if j < 0 {
					return nil, fmt.Errorf("bad quoted key: " + line)
				}
				j += lev + 1
				key = line[lev+1 : j]
				colon := false
				for k := j + 1; k < eol; k++ {
					if line[k] != ' ' {
						if line[k] == ':' {
							j = k
							colon = true
						}
					}
				}
				if !colon {
					return nil, fmt.Errorf("missing colon after key: " + line)
				}
			} else {
				//if a colon is present, the key is between the lev and that position
				for j = lev; j < eol; j++ {
					if line[j] == ':' {
						key = line[lev:j]
						break
					}
				}
				if strings.Index(key, "\"") >= 0 {
					return nil, fmt.Errorf("malformed key: " + line)

				}
			}
			if key == "" {
				return nil, fmt.Errorf("missing colon after key: " + line)
			}
			k = -1
			for i = j + 1; i < eol; i++ {
				if line[i] != ' ' {
					k = i
					break
				}
			}
			var value interface{}
			if k < 0 {
				tmp, err := dd.getNonEmptyLine()
				//tmp, err := dd.getLine()
				if err != nil {
					return nil, err
				}
				eol = len(tmp)
				for i := 0; i < eol; i++ {
					if tmp[i] != ' ' {
						k = i
						break
					}
				}
				dd.ungetLine(tmp)
				value, err = dd.decodeObject(k)
				if err == io.EOF {
					obj[key] = value
					return obj, nil
				} else if err != nil {
					return nil, err
				}
				obj[key] = value
			} else {
				s := line[k:]
				if "|" == s {
					value, err = dd.decodeBlockString(lev)
				}
				obj[key] = dd.decodeScalar(s)
			}

		} else if lev < level {
			//pop out an object
			dd.ungetLine(line)
			break
		} else {
			return obj, nil
		}
		line, err = dd.getNonEmptyLine()
	}
	if err == io.EOF {
		return obj, nil
	}
	return obj, err
}

func (dd *DataDecoder) decodeScalar(s string) interface{} {
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	} else {
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f
		}
		return s
	}
}

func (dd *DataDecoder) decodeBlockString(lev int) (*string, error) {
	line, err := dd.getNonEmptyLine()
	if err != nil {
		return nil, err
	}
	eol := len(line)
	k := -1
	for i := 0; i < eol; i++ {
		if line[i] != ' ' {
			k = i
			break
		}
	}
	if k > lev {
		newlev := k
		svalue := line[newlev:]
		for {
			line, err = dd.getNonEmptyLine()
			if err != nil {
				if err == io.EOF {
					svalue = svalue + "\n"
					return &svalue, nil
				}
				return nil, err
			}
			eol = len(line)
			k = -1
			for i := 0; i < eol; i++ {
				if line[i] != ' ' {
					k = i
					break
				}
			}
			if k < newlev {
				svalue = svalue + "\n"
				dd.ungetLine(line)
				return &svalue, nil
			}
			svalue = svalue + "\n" + line[newlev:]
		}
	}
	dd.ungetLine(line)
	//no value, but can continue
	return nil, nil
}
