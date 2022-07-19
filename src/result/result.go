package result

type Result struct {
	Data   []ResultEntity
	Amount int
}

type ResultEntity struct {
	Type            string
	ID              int
	Value           string
	Context         string
	Version         int
	Properties      map[string]string
	ChildRelations  []ResultRelation
	ParentRelations []ResultRelation
}

type ResultRelation struct {
	Context    string
	Properties map[string]string
	Target     ResultEntity
}

func New() *Result {
	tmp := Result{}
	return &tmp
}

func (self *ResultEntity) Children() []ResultEntity {
	ret := []ResultEntity{}
	for _, resultRelation := range self.ChildRelations {
		ret = append(ret, resultRelation.Target)
	}
	return ret
}

func (self *ResultEntity) Parents() []ResultEntity {
	ret := []ResultEntity{}
	for _, resultRelation := range self.ParentRelations {
		ret = append(ret, resultRelation.Target)
	}
	return ret
}
