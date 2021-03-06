package pak

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"strconv"
	"text/template"

	"github.com/cafebazaar/kupak/logging"
	"github.com/cafebazaar/kupak/pkg/util"

	"github.com/ghodss/yaml"
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
		case "bool":
		case "string":
			// TODO validate the default value and other type specification
			_ = "ok"
		default:
			return errors.New("Specified type is not valid")
		}
	}
	return nil
}

// ID returns unique id for this pak
func (p *Pak) ID() string {
	md5er := md5.New()
	io.WriteString(md5er, p.URL)
	return fmt.Sprintf("%x", md5er.Sum(nil))
}

// receives a directory in which the resources can be found
func (p *Pak) fetchAndMakeTemplates(baseURL string) error {
	p.Templates = make([]*template.Template, len(p.ResourceURLs))
	for i := range p.ResourceURLs {
		url := util.JoinURL(baseURL, p.ResourceURLs[i])
		if !util.Relative(p.ResourceURLs[i]) {
			url = p.ResourceURLs[i]
		}

		data, err := util.FetchURL(url)
		if err != nil {
			return err
		}
		t := template.New(p.ResourceURLs[i])
		t.Delims("$(", ")")
		resourceTemplate, err := t.Parse(string(data))
		if err != nil {
			return err
		}
		p.Templates[i] = resourceTemplate
	}
	return nil
}

// ValidateValues validates given values with corresponding properties
// given values should be contain defaults - use MergeValuesWithDefaults before
// passing values to this function
func (p *Pak) ValidateValues(values map[string]interface{}) error {
	// check all required values are given and their values are ok
	for i := range p.Properties {
		v, has := values[p.Properties[i].Name]
		if !has {
			return errors.New("required property '" + p.Properties[i].Name + "' is not specified")
		}

		ok := false
		switch p.Properties[i].Type {
		case "string":
			_, ok = v.(string)
		case "int":
			_, ok = v.(int)
		case "bool":
			_, ok = v.(bool)
		}
		if !ok {
			return fmt.Errorf("value \"%v\" for property \"%s\" is not correct", v, p.Properties[i].Name)
		}
	}
	return nil
}

// ExecuteTemplates generate resources of a pak with given values
func (p *Pak) ExecuteTemplates(values map[string]interface{}) ([][]byte, error) {
	// merge default values
	err := p.AddDefaultValues(values)
	if err != nil {
		return nil, err
	}

	p.normalizeValues(values)

	err = p.ValidateValues(values)
	if err != nil {
		return nil, err
	}
	outputs := make([][]byte, len(p.Templates))
	for i := range p.Templates {
		buffer := &bytes.Buffer{}

		if err := p.Templates[i].Execute(buffer, values); err != nil {
			return nil, err
		}
		outputs[i] = buffer.Bytes()
	}
	return outputs, nil
}

// AddDefaultValues add default values for properties that not exists in values
func (p *Pak) AddDefaultValues(values map[string]interface{}) error {
	for i := range p.Properties {
		_, ok := values[p.Properties[i].Name]
		if !ok {
			values[p.Properties[i].Name] = p.Properties[i].Default
		}
	}
	return nil
}

// normalizeValues tries to convert default values to their specified type
// like yes, true, 1, ... to boolean true
// this should be called before ValidateValues and after AddDefaultValues
func (p *Pak) normalizeValues(values map[string]interface{}) {
	for i := range p.Properties {
		v, _ := values[p.Properties[i].Name]
		value := fmt.Sprintf("%v", v)
		switch p.Properties[i].Type {
		case "int":
			n, err := strconv.Atoi(value)
			if err == nil {
				values[p.Properties[i].Name] = n
			}
		case "bool":
			b, err := util.StringToBool(value)
			if err == nil {
				values[p.Properties[i].Name] = b
			}
		case "string":
			fallthrough
		default:
			values[p.Properties[i].Name] = v
		}
	}
}

// FromURL reads a pak.yaml file and fetches all the resources files
func FromURL(url string) (*Pak, error) {
	data, err := util.FetchURL(url)
	if err != nil {
		return nil, err
	}
	pak := Pak{}
	if err := yaml.Unmarshal(data, &pak); err != nil {
		logging.Error("Failed to unmarshal pak .yaml file : " + url)
		return nil, err
	}
	if err := validateProperties(pak.Properties); err != nil {
		logging.Error("Failed validating pak properties : " + url)
		return nil, err
	}
	if err := pak.fetchAndMakeTemplates(util.AddressParentNode(url)); err != nil {
		logging.Log("Error making pak templates : " + url)
		return nil, err
	}
	if logging.Verbose {
		logging.Log("Successfully fetched pak file : " + url)
	}
	return &pak, nil
}
