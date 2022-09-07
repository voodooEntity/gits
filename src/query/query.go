package query

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/mutexhandler"
	"github.com/voodooEntity/gits/src/transport"
)

const (
	FILTER_ID      = 0
	FILTER_VALUE   = 1
	FILTER_CONTEXT = 2
)

const (
	DIRECTION_NONE   = -1
	DIRECTION_PARENT = 0
	DIRECTION_CHILD  = 1
)

const (
	METHOD_READ   = 1
	METHOD_REDUCE = 2
	METHOD_UPDATE = 3
	METHOD_UPSERT = 4
	METHOD_DELETE = 5
	METHOD_COUNT  = 6
	METHOD_LINK   = 7
	METHOD_UNLINK = 8
	METHOD_FIND   = 9
)

const (
	ORDER_DIRECTION_ASC  = 1
	ORDER_DIRECTION_DESC = 2
	ORDER_MODE_NUM       = 1
	ORDER_MODE_ALPHA     = 2
)

type Query struct {
	Method             int
	Pool               []string
	Conditions         [][][3]string
	Map                []Query
	Mode               [][]string
	Values             map[string]string
	currConditionGroup int
	Sort               Order
	Direction          int
	Required           bool
}

type Order struct {
	Direction int
	Mode      int
	Field     string
}

func New() *Query {
	tmp := Query{
		Conditions:         [][][3]string{},
		currConditionGroup: 0,
		Direction:          DIRECTION_NONE,
		Values:             make(map[string]string),
		Required:           true,
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

func (self *Query) Find(etype ...string) *Query {
	self.Method = METHOD_FIND
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

func (self *Query) Link(etype ...string) *Query {
	self.Method = METHOD_LINK
	if 0 != len(etype) {
		for _, entry := range etype {
			self.Pool = append(self.Pool, entry)
		}
	}
	return self
}

func (self *Query) Unlink(etype ...string) *Query {
	self.Method = METHOD_UNLINK
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

func (self *Query) To(query *Query) *Query {
	query.SetDirection(DIRECTION_CHILD)
	query.Required = true
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) From(query *Query) *Query {
	query.SetDirection(DIRECTION_PARENT)
	query.Required = true
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) CanTo(query *Query) *Query {
	query.SetDirection(DIRECTION_CHILD)
	query.Required = false
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) CanFrom(query *Query) *Query {
	query.SetDirection(DIRECTION_PARENT)
	query.Required = false
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

func (self *Query) Order(field string, direction int, mode int) *Query {
	self.Sort = Order{
		Direction: direction,
		Mode:      mode,
		Field:     field,
	}
	return self
}

func Execute(query *Query) transport.Transport {
	// no type pool = something is very wrong
	if 0 == len(query.Pool) {
		return transport.Transport{}
	}

	// prepare mutex handler
	mutexh := mutexhandler.New()

	// we are in the most outer layer so we gonne lock here,
	// also dispatch what type of mutex we need. if we only read
	// we can work with a read lock, everything else will need a
	// full lock
	if METHOD_READ == query.Method {
		mutexh.Apply(mutexhandler.EntityTypeRLock)
		mutexh.Apply(mutexhandler.EntityStorageRLock)
	} else {
		mutexh.Apply(mutexhandler.EntityTypeLock)
		mutexh.Apply(mutexhandler.EntityStorageLock)
	}

	// do we have any potential joins? if yes we need to read lock the relation storage
	if 0 < len(query.Map) {
		// if its link method or unlink method we need to write lock the relation storage
		if METHOD_LINK == query.Method || METHOD_UNLINK == query.Method {
			mutexh.Apply(mutexhandler.RelationStorageLock)
		} else {
			mutexh.Apply(mutexhandler.RelationStorageRLock)
		}
	}

	// some dispatches for special query variables
	// do we need to return the data itself?
	var addressPairs [][4]int
	returnDataFlag := false
	linked := true
	linkAddresses := [2][][2]int{}
	linkAmount := 0
	if METHOD_READ == query.Method {
		returnDataFlag = true
	}
	// is it a link query ? needs to be handled different
	if METHOD_LINK == query.Method {
		linked = false
	}

	// parse the conditions into our 2 neccesary groups
	baseMatchList, propertyMatchList := parseConditions(query)

	// now we need to fetch the list of entities fitting to our filters
	resultData, resultAddresses, amount := gits.GetEntitiesByQueryFilter(query.Pool, query.Conditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, returnDataFlag)

	// prepare transport data
	ret := transport.Transport{
		Amount: 0,
	}

	// wo we have any hits?
	if 0 == amount {
		// no hits
		mutexh.Release()
		return ret
	}

	// do we have child queries to execute recursive?
	if 0 < len(query.Map) {
		// do we work with linked data? this gonne be the main case
		if linked {
			for key, entityAddress := range resultAddresses {
				// recursive execute our actions
				children, parents, tmpAddressPairs, amount := recursiveExecuteLinked(query.Map, entityAddress, addressPairs)
				// append the current addresspairs since it can be used for unlinking or other funny stuff
				addressPairs = append(addressPairs, tmpAddressPairs...) // i dont like it but ok for now ### todo overthink
				// now if given store children and parents entities
				if 0 < len(children) {
					resultData[key].ChildRelations = append(resultData[key].ChildRelations, children...)
				}
				if 0 < len(parents) {
					resultData[key].ParentRelations = append(resultData[key].ParentRelations, parents...)
				}

				// are there any results?
				if 0 < amount || !query.HasRequiredSubQueries() {
					ret.Amount++
					// do we have any data to add? (and it is a read)
					if METHOD_READ == query.Method {
						ret.Entities = append(ret.Entities, resultData[key])
					}
				}
			}
		} else { // unlinked data - for now the only case for this is the METHOD_LINK method so we gonne hard handle it that way ###todo maybe expand it on need to have unlinked joins (dont see any case rn)
			for _, targetQuery := range query.Map {
				tagretBaseMatchList, targetPopertyMatchList := parseConditions(&targetQuery)
				_, tmpLinkAddresses, tmpLinkAmount := gits.GetEntitiesByQueryFilter(targetQuery.Pool, targetQuery.Conditions, tagretBaseMatchList[FILTER_ID], tagretBaseMatchList[FILTER_VALUE], tagretBaseMatchList[FILTER_CONTEXT], targetPopertyMatchList, false)
				if 0 < tmpLinkAmount {
					linkAddresses[targetQuery.Direction] = append(linkAddresses[targetQuery.Direction], tmpLinkAddresses...)
					linkAmount = linkAmount + tmpLinkAmount
				}
			}
			ret.Amount = amount
		}
	} else {
		ret.Entities = resultData
		ret.Amount = amount
	}

	if 0 < ret.Amount {
		// now we need to dispatch based on method what we gonne do
		switch query.Method {
		case METHOD_UPDATE:
			// if we got any results and values to update given fire Batch update
			if 0 < len(query.Values) {
				gits.BatchUpdateAddressList(resultAddresses, query.Values)
			}
			mutexh.Release()
		case METHOD_DELETE:
			gits.BatchDeleteAddressList(resultAddresses)
		case METHOD_LINK:
			affectedAmount := 0
			if 0 < linkAmount {
				for direction, tmpLinkAddresses := range linkAddresses {
					if 0 < len(tmpLinkAddresses) {
						if DIRECTION_CHILD == direction {
							affectedAmount += gits.LinkAddressLists(resultAddresses, tmpLinkAddresses)
						} else { // else it must be towards parent so we flip params
							affectedAmount += gits.LinkAddressLists(tmpLinkAddresses, resultAddresses)
						}
					}
				}
			}
			ret.Amount = affectedAmount
		case METHOD_UNLINK:
			affectedAmount := 0
			if 0 < len(addressPairs) {
				for _, addressPair := range addressPairs {
					gits.DeleteRelationUnsafe(addressPair[0], addressPair[1], addressPair[2], addressPair[3])
					affectedAmount++
				}
			}
			ret.Amount = affectedAmount
		}
		if (Order{}) != query.Sort {
			fmt.Println("yes we are sorting")
			ret.Entities = sortResults(ret.Entities, query.Sort.Field, query.Sort.Direction, query.Sort.Mode)
		}
		//sortResults
		mutexh.Release()
		return ret
	}

	// if there were no results we still need to unlock all the mutex
	mutexh.Release()
	return transport.Transport{}
}

func recursiveExecuteLinked(queries []Query, sourceAddress [2]int, addressPairList [][4]int) ([]transport.TransportRelation, []transport.TransportRelation, [][4]int, int) {
	var retParents []transport.TransportRelation
	var retChildren []transport.TransportRelation
	i := 0
	for _, query := range queries {
		var tmpRet []transport.TransportRelation
		// parse the conditions into our 2 neccesary groups
		baseMatchList, propertyMatchList := parseConditions(&query)

		// do we need to return the data itself?
		returnDataFlag := false
		if METHOD_READ == query.Method {
			returnDataFlag = true
		}

		// get data from subquery
		resultData, resultAddresses, amount := gits.GetEntitiesByQueryFilterAndSourceAddress(query.Pool, query.Conditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, sourceAddress, query.Direction, returnDataFlag)

		// if we got no returns
		if 0 == amount {
			// we check if there had to be some
			if true == query.Required {
				// empty return since we got no hits on a required subquery
				return []transport.TransportRelation{}, []transport.TransportRelation{}, [][4]int{}, 0
			}
			// if not we just continue
			continue
		}
		// since we got data we gonne get recursive from here
		if 0 < len(query.Map) {
			for key, entityAddress := range resultAddresses {
				// further execute and store data on return
				children, parents, tmpAddressList, amount := recursiveExecuteLinked(query.Map, entityAddress, addressPairList)
				if DIRECTION_CHILD == query.Direction {
					tmpAddressList = append(tmpAddressList, [4]int{sourceAddress[0], sourceAddress[1], entityAddress[0], entityAddress[1]})
				} else {
					tmpAddressList = append(tmpAddressList, [4]int{entityAddress[0], entityAddress[1], sourceAddress[0], sourceAddress[1]})
				}

				addressPairList = append(addressPairList, tmpAddressList...)
				if 0 < len(children) {
					resultData[key].Target.ChildRelations = append(resultData[key].Target.ChildRelations, children...)
				}
				if 0 < len(parents) {
					resultData[key].Target.ParentRelations = append(resultData[key].Target.ParentRelations, children...)
				}
				if 0 < amount || !query.HasRequiredSubQueries() {
					tmpRet = append(tmpRet, resultData[key])
					i++
				}
			}
		} else {
			// there must be a smarter way for the following problem:
			for _, entityAddress := range resultAddresses {
				if DIRECTION_CHILD == query.Direction {
					addressPairList = append(addressPairList, [4]int{sourceAddress[0], sourceAddress[1], entityAddress[0], entityAddress[1]})
				} else {
					addressPairList = append(addressPairList, [4]int{entityAddress[0], entityAddress[1], sourceAddress[0], sourceAddress[1]})
				}
			}
			// - - - - - - - - - - - - - - - - - -
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
	return retChildren, retParents, addressPairList, i
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
					// ### we need to prepare the map here if it doesn't exist
					propertyMatchList[conditionGroupKey][conditionValue[0][11:]] = append(propertyMatchList[conditionGroupKey][conditionValue[0][11:]], conditionKey)
				}
			}
		}
	}
	return baseMatchList, propertyMatchList
}

func sortResults(results []transport.TransportEntity, field string, direction int, mode int) []transport.TransportEntity {
	cl := func(i, j int) bool {
		// get the values
		sAlpha := results[i].GetFieldByString(field)
		sBeta := results[j].GetFieldByString(field)

		// if mode is numeric we need to int cast the values
		if ORDER_MODE_NUM == mode {
			iAlpha, erra := strconv.ParseInt(sAlpha, 10, 64)
			iBeta, errb := strconv.ParseInt(sBeta, 10, 64)
			if nil == erra && nil == errb {
				if ORDER_DIRECTION_ASC == direction && iAlpha < iBeta || ORDER_DIRECTION_DESC == direction && iAlpha > iBeta {
					return true
				}
			}
		} else { // alphabetical search
			sLowerAlpha := strings.ToLower(sAlpha)
			sLowerBeta := strings.ToLower(sBeta)
			if sLowerAlpha == sLowerBeta {
				if ORDER_DIRECTION_ASC == direction && sAlpha < sBeta || ORDER_DIRECTION_DESC == direction && sAlpha > sBeta {
					return true
				}
			} else {
				if ORDER_DIRECTION_ASC == direction && sLowerAlpha < sLowerBeta || ORDER_DIRECTION_DESC == direction && sLowerAlpha > sLowerBeta {
					return true
				}
			}
		}
		return false
	}
	sort.Slice(results, cl)
	return results
}

func (self *Query) HasRequiredSubQueries() bool {
	for _, qry := range self.Map {
		if true == qry.Required {
			return true
		}
	}
	return false
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
/**
Methods:
-> READ     [x]
-> REDUCE   [x]
-> UPDATE   [x]
-> DELETE   [x]
-> LINK     [X]
-> UNLINK   [X]


Filter:
-> Value       [X]
-> Context     [X]
-> Property    [X]
-> ID          [X]
-> Type        [X]


Compare Operators:
-> equals           [X]
-> prefix           [X]
-> suffix           [X]
-> substring        [X]
-> >=               [X]
-> <=               [X]
-> =                [X]
-> in               [X]


AFTERPROCESSING:
-> ORDER BY % ASC/DESC  [X]

SPECIAL:
-> TRAVERSE    [ ]
-> RTRAVERSE   [ ]
*/
