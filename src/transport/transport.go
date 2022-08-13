package transport

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