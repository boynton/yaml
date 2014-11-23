yaml
====

A quick and dirty YAML parser in Go, with no external dependencies.

This library is not intended to embrace all features of YAML, just a large enough subset to be used for JSON <-> YAML conversions.

## Usage

    go get github.com/boynton/yaml

In your program, the decoded data looks like generic json data, objects are
`map[string]interface{}`, arrays are `[]interface{}`, and strings, numbers, and
bools map to the standard types.

    import "github.com/boynton/yaml"

From a file:

    yamlFilename := "test.yml"
    data, err := yaml.DecodeFile(yamlFilename)

From a byte array:
	
    yamlBytes, _ := ioutil.ReadFile(filename)
    data, err := yaml.DecodeBytes(yamlBytes)

From a string:
	
    yamlString, _ := "one:\n    foo: This is a test\n    bar: 23"
    data, err := yaml.DecodeString(yamlString)

From a reader:

    data, err := yaml.Decode(strings.NewReader(yamlString))


To encode:

    yamlString, err := yaml.Encode(data)

