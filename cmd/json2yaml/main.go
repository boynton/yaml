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
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Fatalf("Cannot read file %q: %v\n", fname, err)
		}
		m := make(map[string]interface{})
		err = json.Unmarshal(data, &m)
		if err != nil {
			log.Fatalf("Cannot parse JSON file %q: %v\n", fname, err)
		}
		b, err := yaml.Marshal(m)
		if err != nil {
			log.Fatalf("Cannot generate YAML for %q: %v\n", fname, err)
		}
		fmt.Println(string(b))
	}
}
