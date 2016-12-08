package lgscript

import (
	"io/ioutil"
	"log"
	"strings"
)

// Load return scripts
func Load(a string) ([]Script, error) {
	bs, err := ioutil.ReadFile("./scripts/aoi.lgscript")
	if err != nil {
		log.Fatal(err)
	}
	r := strings.NewReader(string(bs))

	parser := NewParser(r)

	scripts, err := parser.Parse(a)
	if err != nil {
		println(err.Error())
	}
	return scripts, nil
}
