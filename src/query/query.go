package query

import (
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/result"
	"strings"
)

const (
	FILTER_ID      = 0
	FILTER_VALUE   = 1
	FILTER_CONTEXT = 2
)

const (
	DIRECTION_PARENT = -1
	DIRECTION_NONE   = 0
	DIRECTION_CHILD  = 1
)

const (
	METHOD_READ   = 1
	METHOD_REDUCE = 2
	METHOD_CREATE = 3
	METHOD_UPDATE = 4
	METHOD_DELETE = 5
)

type Query struct {
	Method             int
	Pool               []string
	Conditions         [][][3]string
	Map                []Query
	Mode               [][]string
	Values             map[string]string
	currConditionGroup int
	Direction          int
}

func New() *Query {
	tmp := Query{
		Conditions:         [][][3]string{},
		currConditionGroup: 0,
		Direction:          DIRECTION_NONE,
	}
	return &tmp
}

func (self *Query) Read(etype ...string) *Query {
	self.Method = METHOD_READ
	if 0 != len(etype) {
		for _, entry := range etype {
			self.Pool = append(self.Pool, entry)
		}
	}
	return self
}

func (self *Query) Reduce(etype ...string) *Query {
	self.Method = METHOD_REDUCE
	if 0 != len(etype) {
		for _, entry := range etype {
			self.Pool = append(self.Pool, entry)
		}
	}
	return self
}

func (self *Query) Update(etype ...string) *Query {
	self.Method = METHOD_UPDATE
	if 0 != len(etype) {
		for _, entry := range etype {
			self.Pool = append(self.Pool, entry)
		}
	}
	return self
}

func (self *Query) Delete(etype ...string) *Query {
	self.Method = METHOD_DELETE
	if 0 != len(etype) {
		for _, entry := range etype {
			self.Pool = append(self.Pool, entry)
		}
	}
	return self
}

func (self *Query) Match(alpha string, operator string, beta string) *Query {
	if 0 == len(self.Conditions) {
		self.Conditions = make([][][3]string, 1)
	}
	self.Conditions[self.currConditionGroup] = append(self.Conditions[self.currConditionGroup], [3]string{alpha, operator, beta})
	return self
}

func (self *Query) OrMatch(alpha string, operator string, beta string) *Query {
	self.currConditionGroup++
	self.Conditions = append(self.Conditions, make([][3]string, 1))
	self.Match(alpha, operator, beta)
	return self
}

func (self *Query) Join(query *Query) *Query {
	query.SetDirection(DIRECTION_CHILD)
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) RJoin(query *Query) *Query {
	query.SetDirection(DIRECTION_PARENT)
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) Modify(properties ...string) *Query {
	self.Mode = append(self.Mode, properties)
	return self
}

func (self *Query) SetDirection(direction int) *Query {
	self.Direction = direction
	return self
}

func (self *Query) Set(key string, value string) *Query {
	self.Values[key] = value
	return self
}

func Execute(query *Query) result.Result {
	// if there are no filters something must be terribly wrong ### review this since we may have update/delete/create actions without filters
	if 0 == len(query.Pool) {
		return result.Result{}
	}

	// we are in the most outer layer so we gonne lock here,
	// also dispatch what type of mutex we need. if we only read
	// we can work with a read lock, everything else will need a
	// full lock
	if METHOD_READ == query.Method {
		gits.EntityTypeMutex.RLock()
		gits.EntityStorageMutex.RLock()
	} else {
		gits.EntityStorageMutex.Lock()
	}

	// do we have any potential joins? if yes we need to read lock the relation storage
	// ### maybe add lock for link() method later
	unlockRelationStorage := false
	if 0 < len(query.Map) {
		gits.RelationStorageMutex.RLock()
		unlockRelationStorage = true
	}

	// parse the conditions into our 2 neccesary groups
	baseMatchList, propertyMatchList := parseConditions(query)

	// do we need to return the data itself?
	returnDataFlag := false
	if METHOD_READ == query.Method {
		returnDataFlag = true
	}

	// now we need to fetch the list of entities fitting to our filters
	//var addressList [][2]int
	resultData, resultAddresses, amount := gits.GetEntitiesByQueryFilter(query.Pool, query.Conditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, returnDataFlag)

	// wo we have any hits?
	if 0 == amount {
		// no hits , are we in wrap?
		return result.Result{}
	}

	// do we have child queries to execute recursive?
	ret := result.Result{
		Amount: 0,
	}

	if 0 < len(query.Map) {
		for key, entityAddress := range resultAddresses {
			children, parents, amount := recursiveExecute(query.Map, entityAddress)
			if 0 < len(children) {
				resultData[key].ChildRelations = append(resultData[key].ChildRelations, children...)
			}
			if 0 < len(parents) {
				resultData[key].ParentRelations = append(resultData[key].ParentRelations, parents...)
			}
			// do we have any data to add?
			// if true == add { ### refactor add flag usage
			if 0 < amount {
				ret.Data = append(ret.Data, resultData[key])
				ret.Amount++
			}
		}
	} else {
		ret.Data = resultData
		ret.Amount = amount
	}

	if 0 < ret.Amount {
		// now we need to dispatch based on method what we gonne do
		switch query.Method {
		case METHOD_READ:
			gits.EntityTypeMutex.RUnlock()
			gits.EntityStorageMutex.RUnlock()
			if unlockRelationStorage {
				gits.RelationStorageMutex.RUnlock()
			}
			return ret
		case METHOD_UPDATE:
		case METHOD_DELETE:
		case METHOD_CREATE:
		}
	}

	// finally check if we are in the wrapping query and unlock everything
	if METHOD_READ == query.Method {
		gits.EntityStorageMutex.RUnlock()
	} else {
		gits.EntityStorageMutex.Unlock()
	}
	return result.Result{}
}

func recursiveExecute(queries []Query, sourceAddress [2]int) ([]result.ResultRelation, []result.ResultRelation, int) {
	var retParents []result.ResultRelation
	var retChildren []result.ResultRelation
	i := 0
	for _, query := range queries {
		var tmpRet []result.ResultRelation
		// parse the conditions into our 2 neccesary groups
		baseMatchList, propertyMatchList := parseConditions(&query)

		// do we need to return the data itself?
		returnDataFlag := false
		if METHOD_READ == query.Method {
			returnDataFlag = true
		}

		// get data from subquery
		resultData, resultAddresses, amount := gits.GetEntitiesByQueryFilterAndSourceAddress(query.Pool, query.Conditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, sourceAddress, query.Direction, returnDataFlag)
		// if we got no returns we continue
		if 0 == amount {
			continue
		}
		// since we got data we gonne get recursive from here
		if 0 < len(query.Map) {
			for key, entityAddress := range resultAddresses {
				children, parents, amount := recursiveExecute(query.Map, entityAddress)
				if 0 < len(children) {
					resultData[key].Target.ChildRelations = append(resultData[key].Target.ChildRelations, children...)
				}
				if 0 < len(parents) {
					resultData[key].Target.ParentRelations = append(resultData[key].Target.ParentRelations, children...)
				}
				if 0 < amount {
					tmpRet = append(tmpRet, resultData[key])
					i++
				}
			}
		} else {
			i = amount
			tmpRet = append(tmpRet, resultData...)
		}
		// if we got any results we add them
		if 0 < len(tmpRet) {
			// add the results to either child direction list
			if DIRECTION_CHILD == query.Direction {
				retChildren = append(retChildren, tmpRet...)
			} else {
				// or we assume its DIRECTION_PARENT if not child
				retParents = append(retParents, tmpRet...)
			}
		}

	}
	return retChildren, retParents, i
}

func parseConditions(query *Query) ([3][][]int, []map[string][]int) {
	// now we need to identify what we are searching for
	//  0 => ID, 1 => Value , 2 => Context
	baseMatchList := [3][][]int{{}, {}, {}}
	propertyMatchList := []map[string][]int{}
	for conditionGroupKey, conditionGroup := range query.Conditions {
		// sub allocate arrays for each condition group to make sure we dont have missing entries
		// slices u know...
		// first for the base filters
		for _, filterGroup := range [3]int{FILTER_ID, FILTER_VALUE, FILTER_CONTEXT} {
			baseMatchList[filterGroup] = append(baseMatchList[filterGroup], []int{})
			baseMatchList[filterGroup][conditionGroupKey] = []int{}
		}
		// than for the property filter
		propertyMatchList = append(propertyMatchList, map[string][]int{})
		// than we actually parse the conditions
		for conditionKey, conditionValue := range conditionGroup {
			switch conditionValue[0] {
			case "ID":
				baseMatchList[FILTER_ID][conditionGroupKey] = append(baseMatchList[FILTER_ID][conditionGroupKey], conditionKey)
			case "Value":
				baseMatchList[FILTER_VALUE][conditionGroupKey] = append(baseMatchList[FILTER_VALUE][conditionGroupKey], conditionKey)
			case "Context":
				baseMatchList[FILTER_CONTEXT][conditionGroupKey] = append(baseMatchList[FILTER_CONTEXT][conditionGroupKey], conditionKey)
			default:
				if -1 != strings.Index(conditionValue[0], "Properties") {
					// ### we nmeed to prepare the map here if it doesnt exist
					propertyMatchList[conditionGroupKey][conditionValue[0][11:]] = append(propertyMatchList[conditionGroupKey][conditionValue[0][11:]], conditionKey)
				}
			}
		}
	}
	return baseMatchList, propertyMatchList
}
