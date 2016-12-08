package lgscript

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/mituoh/line-bot-gamebook/lgscript"
)

func TestParse(t *testing.T) {
	bs, err := ioutil.ReadFile("../app/test.lgscript")
	if err != nil {
		log.Fatal(err)
	}
	r := strings.NewReader(string(bs))

	parser := lgscript.NewParser(r)

	scripts, err := parser.Parse("*start")
	if err != nil {
		println(err.Error())
	}

	for _, script := range scripts {
		println(script.Text)
	}
}
