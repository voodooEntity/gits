package transport

import (
	"strconv"
	"strings"
)

type Transport struct {
	Entities  []TransportEntity
	Relations []TransportRelation
	Amount    int
}

type TransportEntity struct {
	Type            string
	ID              int
	Value           string
	Context         string
	Version         int
	Properties      map[string]string
	ChildRelations  []TransportRelation
	ParentRelations []TransportRelation
}

type TransportRelation struct {
	Context    string
	Properties map[string]string
	Target     TransportEntity
	SourceType string
	SourceID   int
	TargetType string
	TargetID   int
	Version    int
}

func New() *Transport {
	tmp := Transport{}
	return &tmp
}

func (self *TransportEntity) Children() []TransportEntity {
	ret := []TransportEntity{}
	for _, resultRelation := range self.ChildRelations {
		ret = append(ret, resultRelation.Target)
	}
	return ret
}

func (self *TransportEntity) Parents() []TransportEntity {
	ret := []TransportEntity{}
	for _, resultRelation := range self.ParentRelations {
		ret = append(ret, resultRelation.Target)
	}
	return ret
}

func (self *TransportEntity) GetFieldByString(field string) string {
	switch field {
	case "ID":
		return strconv.Itoa(self.ID)
	case "Context":
		return self.Context
	case "Value":
		return self.Value
	default:
		if -1 != strings.Index(field, "Properties") {
			// ### we need to prepare the map here if it doesn't exist
			property := field[11:]
			if nil != self.Properties {
				if val, ok := self.Properties[property]; ok {
					return val
				}
			}
		}
	}
	return ""
}
