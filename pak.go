package kupak

import (
	"bytes"
	"errors"
	"text/template"

	"gopkg.in/yaml.v2"
)

func validateProperties(properties []Property) error {
	nameMap := make(map[string]bool)
	for i := range properties {
		if _, has := nameMap[properties[i].Name]; has {
			return errors.New("Duplicated property")
		}
		// validating types
		switch properties[i].Type {
		case "int":
		case "number":
		case "string":
			// TODO validate the default value and other type specification
			_ = "ok"
		default:
			return errors.New("Specified type is not valid")
		}
	}
	return nil
}

func (p *Pak) fetchAndMakeTemplates(baseUrl string) error {
	p.Templates = make([]*template.Template, len(p.ResourceUrls))
	for i := range p.ResourceUrls {
		url := joinUrl(baseUrl, p.ResourceUrls[i])
		data, err := fetchUrl(url)
		if err != nil {
			return err
		}
		t := template.New(p.ResourceUrls[i])
		t.Delims("$(", ")")
		resourceTemplate, err := t.Parse(string(data))
		if err != nil {
			return err
		}
		p.Templates[i] = resourceTemplate
	}
	return nil
}

func (p *Pak) ExecuteTemplates(values map[string]interface{}) ([][]byte, error) {
	// TODO validate values
	// TODO copy values
	outputs := make([][]byte, len(p.Templates))
	for i := range p.Templates {
		buffer := &bytes.Buffer{}
		if err := p.valuesWithDefaults(values); err != nil {
			return nil, err
		}
		if err := p.Templates[i].Execute(buffer, values); err != nil {
			return nil, err
		}
		outputs[i] = buffer.Bytes()
	}
	return outputs, nil
}

func (p *Pak) valuesWithDefaults(values map[string]interface{}) error {
	for i := range p.Properties {
		if _, ok := values[p.Properties[i].Name]; !ok {
			values[p.Properties[i].Name] = p.Properties[i].Default
		}
	}
	return nil
}

func PakFromUrl(url string) (*Pak, error) {
	data, err := fetchUrl(url)
	if err != nil {
		return nil, err
	}
	pak := Pak{}
	if err := yaml.Unmarshal(data, &pak); err != nil {
		return nil, err
	}
	if err := validateProperties(pak.Properties); err != nil {
		return nil, err
	}
	if err := pak.fetchAndMakeTemplates(url); err != nil {
		return nil, err
	}
	println(pak.Templates[0])
	return &pak, nil
}