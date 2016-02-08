package kupak

import (
	"fmt"

	"github.com/nu7hatch/gouuid"
)

type Status int

const (
	StatusError Status = iota
	StatusRunning
	StatusDeleting
)

type InstalledPak struct {
	Group      string
	Namespace  string
	PakUrl     string
	Properties map[string]string
	Objects    []interface{}
	Status     Status
}

type Manager struct {
}

func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

func (m *Manager) Installed(namespace string) ([]*InstalledPak, error) {
	return nil, nil
}

func (m *Manager) Instances(namespace string, pak *Pak) ([]*InstalledPak, error) {
	return nil, nil
}

func (m *Manager) Status(namespace string, instance string) (*InstalledPak, error) {
	return nil, nil
}

// Install a pak with given name
func (m *Manager) Install(pak *Pak, namespace string, properties map[string]interface{}) error {
	rawObjects, err := pak.ExecuteTemplates(properties)
	if err != nil {
		return err
	}
	group, err := uuid.NewV4()
	if err != nil {
		return err
	}
	labels := map[string]string{
		"kupak-group":   group.String(),
		"kupak-pak-url": pak.URL,
	}
	var objects []*Object
	for i := range rawObjects {
		object, err := NewObject(rawObjects[i])
		if err != nil {
			return err
		}
		md, err := object.Metadata()
		if err != nil {
			return err
		}
		mergedLabels := MergeStringMaps(md.Labels, labels)
		err = object.SetLabels(mergedLabels)
		if err != nil {
			return err
		}
		// TODO validation for replication controller - do not ignore
		if templateMd, err := object.TemplateMetadata(); err == nil {
			mergedLabels := MergeStringMaps(templateMd.Labels, labels)
			object.SetTemplateLabels(mergedLabels)
		}
		bytes, _ := object.Bytes()
		fmt.Println(string(bytes))
		fmt.Println("----\n----")
		objects = append(objects, object)
	}
	return nil
}

// DeleteInstance will delete a installed pak
func (m *Manager) DeleteInstance(namespace string, group string) ([]*InstalledPak, error) {
	return nil, nil
}