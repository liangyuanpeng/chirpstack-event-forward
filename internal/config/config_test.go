package config

import (
	"os"
	"testing"
	"text/template"
)

const templ = `
name is {{.name}}
Company is {{.resources.Company}}
smartoilets-{{ .customerName }}/application/{{ .appid }}/device/{{ .devEUI }}/{{ .event }}
`

func Test_Templat(test *testing.T) {
	t := template.New("Person template")
	tem, err := t.Parse(templ)
	if err != nil {
		test.Fatal(err)
	}

	var tmp map[string]interface{} = map[string]interface{}{
		"name":         "macs",
		"customerName": "lan",
		"appid":        "3",
		"devEUI":       "asd123",
		"event":        "up",
	}
	tem.Execute(os.Stdout, tmp)
}
