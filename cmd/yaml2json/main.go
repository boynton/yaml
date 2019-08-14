package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
//	"github.com/boynton/yaml"
	"gopkg.in/yaml.v2"
)

func main() {
	for _, fname := range(os.Args[1:]) {
//		dom, err := yaml.DecodeFile(fname)
		dom, err := DecodeFile(fname)
		if err != nil {
			log.Fatalf("Cannot parse %q: %v\n", fname, err)
		}
		b, err := json.MarshalIndent(dom, "", "    ")
		if err != nil {
			log.Fatalf("Cannot generate JSON for %q: %v\n", fname, err)
		}
		fmt.Println(string(b))
	}
}

//broken: go-yaml by defaults produces maps with arbitrary keys. JSON requires string keys.
func DecodeFile(fname string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}
	return mapToJSON(m)
}

func yamlToJSONValue(obj interface{}) (interface{}, error) {
	switch v := obj.(type) {
	case map[interface{}]interface{}:
		return mapToJSON(v)
	case []interface{}:
		return aryToJSON(v)
	case string, bool, int, float64:
		return obj, nil
	default:
	}
	return nil, fmt.Errorf("Not a valid JSON value: %v", obj)
}

func mapToJSON(src map[interface{}]interface{}) (map[string]interface{}, error) {
	dst := make(map[string]interface{})
	for k, v := range src {
		vv, err := yamlToJSONValue(v)
		if err != nil {
			return nil, err
		}
		kk := fmt.Sprintf("%v", k)
		dst[kk] = vv
	}
	return dst, nil
}

func aryToJSON(src []interface{}) ([]interface{}, error) {
	dst := make([]interface{}, 0, len(src))
	for _, v := range src {
		o, err := yamlToJSONValue(v)
		if err != nil {
			return nil, err
		}
		dst = append(dst, o)
	}
	return dst, nil
}
