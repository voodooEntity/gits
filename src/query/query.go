package query

import (
	"sort"
	"strconv"
	"strings"

	"github.com/voodooEntity/gits/src/storage"

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
	query.setDirection(DIRECTION_CHILD)
	query.Required = true
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) From(query *Query) *Query {
	query.setDirection(DIRECTION_PARENT)
	query.Required = true
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) CanTo(query *Query) *Query {
	query.setDirection(DIRECTION_CHILD)
	query.Required = false
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) CanFrom(query *Query) *Query {
	query.setDirection(DIRECTION_PARENT)
	query.Required = false
	self.Map = append(self.Map, *query)
	return self
}

func (self *Query) Modify(properties ...string) *Query {
	self.Mode = append(self.Mode, properties)
	return self
}

func (self *Query) setDirection(direction int) *Query {
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

func (self *Query) TraverseOut(depth int) *Query {
	self.Mode = append(self.Mode, []string{"Traverse", strconv.Itoa(DIRECTION_CHILD), strconv.Itoa(depth)})
	return self
}

func (self *Query) TraverseIn(depth int) *Query {
	self.Mode = append(self.Mode, []string{"Traverse", strconv.Itoa(DIRECTION_PARENT), strconv.Itoa(depth)})
	return self
}

func (self *Query) Limit(amount int) *Query {
	self.Mode = append(self.Mode, []string{"Limit", strconv.Itoa(amount)})
	return self
}

func Execute(store *storage.Storage, query *Query) transport.Transport {
	if 0 == len(query.Pool) {
		return transport.Transport{}
	}

	mutexh := mutexhandler.New(store)
	if METHOD_READ == query.Method {
		mutexh.Apply(mutexhandler.EntityTypeRLock)
		mutexh.Apply(mutexhandler.EntityStorageRLock)
	} else {
		mutexh.Apply(mutexhandler.EntityTypeLock)
		mutexh.Apply(mutexhandler.EntityStorageLock)
	}

	if 0 < len(query.Map) {
		if METHOD_LINK == query.Method || METHOD_UNLINK == query.Method {
			mutexh.Apply(mutexhandler.RelationStorageLock)
		} else {
			mutexh.Apply(mutexhandler.RelationStorageRLock)
		}
	}

	var addressPairs [][4]int
	returnDataFlag := false
	linked := true
	linkAddresses := [2][][2]int{}
	linkAmount := 0

	if METHOD_READ == query.Method {
		returnDataFlag = true
	}
	if METHOD_LINK == query.Method {
		linked = false
	}

	baseMatchList, propertyMatchList := parseConditions(query)
	initialResultData, initialResultAddresses, initialAmount := store.GetEntitiesByQueryFilter(query.Pool, query.Conditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, returnDataFlag)

	ret := transport.Transport{
		Amount: 0,
	}

	if 0 == initialAmount {
		mutexh.Release()
		return ret
	}

	var finalFilteredAddresses [][2]int
	var tempEntitiesForRead []transport.TransportEntity

	if 0 < len(query.Map) {
		if linked { // Path for Read-with-joins, Update, Delete, Unlink
			collectAddressPairs := [][4]int{}
			for key, entityAddress := range initialResultAddresses {
				// Pass empty `addressPairs` for this call as it's for filtering/populating, not Unlink pair collection at this level
				childrenFromSubquery, parentsFromSubquery, tmpAddressPairsFromSub, subAmount := recursiveExecuteLinked(store, query.Map, entityAddress, [][4]int{})

				if query.HasRequiredSubQueries() && subAmount == 0 {
					continue
				}
				finalFilteredAddresses = append(finalFilteredAddresses, entityAddress)

				if query.Method == METHOD_UNLINK {
					// tmpAddressPairsFromSub contains pairs from subqueries linked to entityAddress
					collectAddressPairs = append(collectAddressPairs, tmpAddressPairsFromSub...)
				}

				if METHOD_READ == query.Method {
					currentEntityDataForRead := initialResultData[key]
					if 0 < len(childrenFromSubquery) {
						currentEntityDataForRead.ChildRelations = append(currentEntityDataForRead.ChildRelations, childrenFromSubquery...)
					}
					if 0 < len(parentsFromSubquery) {
						currentEntityDataForRead.ParentRelations = append(currentEntityDataForRead.ParentRelations, parentsFromSubquery...)
					}
					tempEntitiesForRead = append(tempEntitiesForRead, currentEntityDataForRead)
				}
			}
			if query.Method == METHOD_UNLINK {
				addressPairs = collectAddressPairs // Use the collected pairs
			}
			if METHOD_READ == query.Method {
				ret.Entities = tempEntitiesForRead
			}
			ret.Amount = len(finalFilteredAddresses)

		} else { // Path for METHOD_LINK (linked = false)
			finalFilteredAddresses = initialResultAddresses
			for _, targetQuery := range query.Map {
				tagretBaseMatchList, targetPopertyMatchList := parseConditions(&targetQuery)
				_, tmpLinkAddresses, tmpLinkAmount := store.GetEntitiesByQueryFilter(targetQuery.Pool, targetQuery.Conditions, tagretBaseMatchList[FILTER_ID], tagretBaseMatchList[FILTER_VALUE], tagretBaseMatchList[FILTER_CONTEXT], targetPopertyMatchList, false)
				if 0 < tmpLinkAmount {
					linkAddresses[targetQuery.Direction] = append(linkAddresses[targetQuery.Direction], tmpLinkAddresses...)
					linkAmount = linkAmount + tmpLinkAmount
				}
			}
			ret.Amount = initialAmount
		}
	} else { // No subqueries in query.Map
		finalFilteredAddresses = initialResultAddresses
		if METHOD_READ == query.Method {
			ret.Entities = initialResultData
		}
		ret.Amount = initialAmount
	}

	if query.Method == METHOD_UPDATE || query.Method == METHOD_DELETE || query.Method == METHOD_READ {
		if len(finalFilteredAddresses) == 0 {
			mutexh.Release()
			return transport.Transport{}
		}
		if query.Method == METHOD_UPDATE || query.Method == METHOD_DELETE {
			ret.Amount = len(finalFilteredAddresses)
		}
		// For METHOD_READ, ret.Amount is already len(finalFilteredAddresses) or initialAmount if no map.
	} else if query.Method == METHOD_UNLINK {
		if len(addressPairs) == 0 {
			mutexh.Release()
			return transport.Transport{}
		}
	} else if query.Method == METHOD_LINK {
		if len(finalFilteredAddresses) == 0 || linkAmount == 0 {
			ret.Amount = 0 // No sources or no targets means 0 links will be made.
		}
		// ret.Amount will be updated to affectedAmount later by the LINK case.
	}

	switch query.Method {
	case METHOD_UPDATE:
		if 0 < len(query.Values) && len(finalFilteredAddresses) > 0 {
			store.BatchUpdateAddressList(finalFilteredAddresses, query.Values)
		}
		ret.Amount = len(finalFilteredAddresses) // Ensure Amount reflects actual items considered for update
	case METHOD_DELETE:
		if len(finalFilteredAddresses) > 0 {
			store.BatchDeleteAddressList(finalFilteredAddresses)
		}
		ret.Amount = len(finalFilteredAddresses) // Ensure Amount reflects actual items considered for delete
	case METHOD_LINK:
		affectedAmount := 0
		if 0 < linkAmount && len(finalFilteredAddresses) > 0 {
			for direction, currentTargetAddresses := range linkAddresses {
				if 0 < len(currentTargetAddresses) {
					if DIRECTION_CHILD == direction {
						affectedAmount += store.LinkAddressLists(finalFilteredAddresses, currentTargetAddresses)
					} else {
						affectedAmount += store.LinkAddressLists(currentTargetAddresses, finalFilteredAddresses)
					}
				}
			}
		}
		ret.Amount = affectedAmount
	case METHOD_UNLINK:
		affectedAmount := 0
		if 0 < len(addressPairs) {
			for _, addressPair := range addressPairs {
				store.DeleteRelationUnsafe(addressPair[0], addressPair[1], addressPair[2], addressPair[3])
				affectedAmount++
			}
		}
		ret.Amount = affectedAmount
	case METHOD_READ:
		if direction, depth, traversed := isTraversed(*query); traversed {
			for id := range ret.Entities {
				store.TraverseEnrich(&(ret.Entities[id]), direction, depth)
			}
		}
	}

	if METHOD_READ == query.Method || ((query.Method == METHOD_UPDATE || query.Method == METHOD_DELETE) && returnDataFlag) {
		if (Order{}) != query.Sort {
			ret.Entities = sortResults(ret.Entities, query.Sort.Field, query.Sort.Direction, query.Sort.Mode)
		}
		limit := getLimitIfExists(*query)
		if -1 != limit {
			if len(ret.Entities) > limit {
				ret.Entities = ret.Entities[:limit]
				if METHOD_READ == query.Method { // Only adjust amount for READ if limited
					ret.Amount = len(ret.Entities)
				}
			}
		}
	}

	mutexh.Release()
	return ret
}

func recursiveExecuteLinked(store *storage.Storage, queries []Query, sourceAddress [2]int, addressPairListFromCaller [][4]int) ([]transport.TransportRelation, []transport.TransportRelation, [][4]int, int) {
	var retParents []transport.TransportRelation
	var retChildren []transport.TransportRelation
	var collectedAddressPairsForUnlink [][4]int // Pairs formed at this level of recursion

	overallSuccessfulPathsForThisLevel := 0

	for _, currentSubQuery := range queries {
		var fullyProcessedSubRelationsForCurrentQuery []transport.TransportRelation
		baseMatchList, propertyMatchList := parseConditions(&currentSubQuery)

		subQueryReturnDataFlag := false
		if METHOD_READ == currentSubQuery.Method {
			subQueryReturnDataFlag = true
		}

		resultSubData, resultSubAddresses, directMatchCount := store.GetEntitiesByQueryFilterAndSourceAddress(currentSubQuery.Pool, currentSubQuery.Conditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, sourceAddress, currentSubQuery.Direction, subQueryReturnDataFlag)

		if 0 == directMatchCount {
			if true == currentSubQuery.Required {
				return []transport.TransportRelation{}, []transport.TransportRelation{}, [][4]int{}, 0
			}
			continue
		}

		successfulPathsThroughCurrentSubQuery := 0

		if 0 < len(currentSubQuery.Map) { // currentSubQuery has nested children/parents
			for key, relatedEntityAddress := range resultSubAddresses {
				// Pass empty [][4]int{} for addressPairListFromCaller to nested calls,
				// as pair collection is per level for Unlink.
				nestedChildren, nestedParents, _, nestedSubAmount := recursiveExecuteLinked(store, currentSubQuery.Map, relatedEntityAddress, [][4]int{}) // nestedAddressPairs assigned to _

				if currentSubQuery.HasRequiredSubQueries() && nestedSubAmount == 0 {
					continue // This relatedEntityAddress failed its own required nested join.
				}

				successfulPathsThroughCurrentSubQuery++

				if subQueryReturnDataFlag {
					currentRelation := resultSubData[key]
					if 0 < len(nestedChildren) {
						currentRelation.Target.ChildRelations = append(currentRelation.Target.ChildRelations, nestedChildren...)
					}
					if 0 < len(nestedParents) {
						currentRelation.Target.ParentRelations = append(currentRelation.Target.ParentRelations, nestedParents...)
					}
					fullyProcessedSubRelationsForCurrentQuery = append(fullyProcessedSubRelationsForCurrentQuery, currentRelation)
				}
				// Collect pairs for Unlink: these are pairs formed by sourceAddress and relatedEntityAddress,
				// assuming this path (including nested) is valid.
				if DIRECTION_CHILD == currentSubQuery.Direction {
					collectedAddressPairsForUnlink = append(collectedAddressPairsForUnlink, [4]int{sourceAddress[0], sourceAddress[1], relatedEntityAddress[0], relatedEntityAddress[1]})
				} else {
					collectedAddressPairsForUnlink = append(collectedAddressPairsForUnlink, [4]int{relatedEntityAddress[0], relatedEntityAddress[1], sourceAddress[0], sourceAddress[1]})
				}
				// Note: nestedAddressPairs are not directly used here, they would have been handled by deeper Unlink if query was structured that way.
			}
		} else { // currentSubQuery has no nested children/parents
			successfulPathsThroughCurrentSubQuery = directMatchCount
			if subQueryReturnDataFlag {
				fullyProcessedSubRelationsForCurrentQuery = append(fullyProcessedSubRelationsForCurrentQuery, resultSubData...)
			}
			for _, relatedEntityAddress := range resultSubAddresses {
				if DIRECTION_CHILD == currentSubQuery.Direction {
					collectedAddressPairsForUnlink = append(collectedAddressPairsForUnlink, [4]int{sourceAddress[0], sourceAddress[1], relatedEntityAddress[0], relatedEntityAddress[1]})
				} else {
					collectedAddressPairsForUnlink = append(collectedAddressPairsForUnlink, [4]int{relatedEntityAddress[0], relatedEntityAddress[1], sourceAddress[0], sourceAddress[1]})
				}
			}
		}

		if successfulPathsThroughCurrentSubQuery == 0 && currentSubQuery.Required {
			return []transport.TransportRelation{}, []transport.TransportRelation{}, [][4]int{}, 0
		}

		overallSuccessfulPathsForThisLevel += successfulPathsThroughCurrentSubQuery

		if subQueryReturnDataFlag && 0 < len(fullyProcessedSubRelationsForCurrentQuery) {
			var appender *[]transport.TransportRelation
			if DIRECTION_CHILD == currentSubQuery.Direction {
				appender = &retChildren
			} else {
				appender = &retParents
			}
			start := len(*appender)
			*appender = append(*appender, fullyProcessedSubRelationsForCurrentQuery...)
			if direction, depth, ok := isTraversed(currentSubQuery); ok {
				for idx := start; idx < len(*appender); idx++ {
					store.TraverseEnrich(&((*appender)[idx].Target), direction, depth)
				}
			}
		}
	}
	return retChildren, retParents, collectedAddressPairsForUnlink, overallSuccessfulPathsForThisLevel
}

func parseConditions(query *Query) ([3][][]int, []map[string][]int) {
	baseMatchList := [3][][]int{{}, {}, {}}
	propertyMatchList := []map[string][]int{}
	for conditionGroupKey, conditionGroup := range query.Conditions {
		for _, filterGroup := range [3]int{FILTER_ID, FILTER_VALUE, FILTER_CONTEXT} {
			baseMatchList[filterGroup] = append(baseMatchList[filterGroup], []int{})
			baseMatchList[filterGroup][conditionGroupKey] = []int{}
		}
		propertyMatchList = append(propertyMatchList, map[string][]int{})
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
					propertyName := conditionValue[0][11:]
					if _, ok := propertyMatchList[conditionGroupKey][propertyName]; !ok {
						propertyMatchList[conditionGroupKey][propertyName] = []int{}
					}
					propertyMatchList[conditionGroupKey][propertyName] = append(propertyMatchList[conditionGroupKey][propertyName], conditionKey)
				}
			}
		}
	}
	return baseMatchList, propertyMatchList
}

func sortResults(results []transport.TransportEntity, field string, direction int, mode int) []transport.TransportEntity {
	cl := func(i, j int) bool {
		sAlpha := results[i].GetFieldByString(field)
		sBeta := results[j].GetFieldByString(field)

		if ORDER_MODE_NUM == mode {
			iAlpha, erra := strconv.ParseInt(sAlpha, 10, 64)
			iBeta, errb := strconv.ParseInt(sBeta, 10, 64)
			if nil == erra && nil == errb {
				if ORDER_DIRECTION_ASC == direction && iAlpha < iBeta || ORDER_DIRECTION_DESC == direction && iAlpha > iBeta {
					return true
				}
			}
		} else {
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

func isTraversed(qry Query) (int, int, bool) {
	if nil != qry.Mode {
		for _, mode := range qry.Mode {
			tmpLen := len(mode)
			if 0 < tmpLen && "Traverse" == mode[0] {
				if 3 == tmpLen {
					direction, err := strconv.ParseInt(mode[1], 10, 64)
					if nil != err {
						return -1, -1, false
					}
					depth, err := strconv.ParseInt(mode[2], 10, 64)
					if nil != err {
						return -1, -1, false
					}
					return int(direction), int(depth), true
				}
			}
		}
	}
	return -1, -1, false
}

func getLimitIfExists(qry Query) int {
	if nil != qry.Mode {
		for _, mode := range qry.Mode {
			tmpLen := len(mode)
			if 0 < tmpLen && "Limit" == mode[0] {
				limit, err := strconv.Atoi(mode[1])
				if nil != err {
					return -1
				}
				return limit
			}
		}
	}
	return -1
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


POSTPROCESSING:
-> ORDER BY % ASC/DESC  [X]
-> TraverseOut          [X]
-> TraverseIn           [X]
-> LIMIT                [ ] // Implemented but not in this list?

*/
