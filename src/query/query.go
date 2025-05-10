package query

import (
	"sort"
	"strconv"
	"strings"

	"github.com/voodooEntity/gits/src/storage"

	"github.com/voodooEntity/gits/src/mutexhandler"
	"github.com/voodooEntity/gits/src/query/cond" // Added import for cond package
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
	RootFilter         cond.Condition // New field for complex conditions

	// Fields for enhanced Link/Unlink
	LinkRelationContext           string
	LinkRelationProperties        map[string]string
	UnlinkRelationContextFilter   string
	UnlinkRelationPropertyFilters []RelationPropertyFilter
}

// RelationPropertyFilter defines a filter condition for a relation's property.
type RelationPropertyFilter struct {
	Key      string
	Operator string
	Value    string
}

type Order struct {
	Direction int
	Mode      int
	Field     string
}

func New() *Query {
	tmp := Query{
		Conditions:                    [][][3]string{},
		currConditionGroup:            0,
		Direction:                     DIRECTION_NONE,
		Values:                        make(map[string]string),
		Required:                      true,
		LinkRelationProperties:        make(map[string]string),    // Initialize map for Link properties
		UnlinkRelationPropertyFilters: []RelationPropertyFilter{}, // Initialize slice for Unlink property filters
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

// Filter sets the root condition for the query using the new complex condition model.
// If this method is used, any conditions set by Match() or OrMatch() will be ignored.
func (self *Query) Filter(condition cond.Condition) *Query {
	self.RootFilter = condition
	// By design, RootFilter takes precedence. Execute logic will handle this.
	// No need to clear self.Conditions here, as Execute will check RootFilter first.
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

// WithRelationContext sets the context for relations created by a Link query.
func (self *Query) WithRelationContext(context string) *Query {
	if self.Method != METHOD_LINK {
		// Optionally log a warning or simply ignore if not a LINK query
		return self
	}
	self.LinkRelationContext = context
	return self
}

// WithRelationProperty sets a single property for relations created by a Link query.
func (self *Query) WithRelationProperty(key string, value string) *Query {
	if self.Method != METHOD_LINK {
		return self
	}
	// Ensure the map is initialized (it is in New(), but good practice for builder methods)
	if self.LinkRelationProperties == nil {
		self.LinkRelationProperties = make(map[string]string)
	}
	self.LinkRelationProperties[key] = value
	return self
}

// WithRelationProperties sets multiple properties for relations created by a Link query.
// This will merge with any existing properties set by WithRelationProperty.
func (self *Query) WithRelationProperties(properties map[string]string) *Query {
	if self.Method != METHOD_LINK {
		return self
	}
	if self.LinkRelationProperties == nil {
		self.LinkRelationProperties = make(map[string]string)
	}
	for k, v := range properties {
		self.LinkRelationProperties[k] = v
	}
	return self
}

// MatchingRelationContext sets a context filter for relations targeted by an Unlink query.
func (self *Query) MatchingRelationContext(context string) *Query {
	if self.Method != METHOD_UNLINK {
		// Optionally log a warning or simply ignore if not an UNLINK query
		return self
	}
	self.UnlinkRelationContextFilter = context
	return self
}

// MatchingRelationProperty adds a property filter for relations targeted by an Unlink query.
func (self *Query) MatchingRelationProperty(key string, operator string, value string) *Query {
	if self.Method != METHOD_UNLINK {
		return self
	}
	// Ensure the slice is initialized (it is in New(), but good practice)
	if self.UnlinkRelationPropertyFilters == nil {
		self.UnlinkRelationPropertyFilters = []RelationPropertyFilter{}
	}
	self.UnlinkRelationPropertyFilters = append(self.UnlinkRelationPropertyFilters, RelationPropertyFilter{Key: key, Operator: operator, Value: value})
	return self
}

func Execute(store *storage.Storage, query *Query) transport.Transport {
	// no type pool = something is very wrong
	if 0 == len(query.Pool) {
		return transport.Transport{}
	}

	// prepare mutex handler
	mutexh := mutexhandler.New(store)

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
	returnDataFlag := false
	if METHOD_READ == query.Method {
		returnDataFlag = true
	}
	// linked, linkAddresses, linkAmount, and addressPairs were removed as they are handled by new variables
	// or specific logic paths.

	// parse the conditions into our 2 neccesary groups
	var baseMatchList [3][][]int
	var propertyMatchList []map[string][]int
	var legacyConditions [][][3]string

	if query.RootFilter == nil {
		legacyConditions = query.Conditions
		baseMatchList, propertyMatchList = parseConditions(query)
	} else {
		// Ensure these are empty or nil if RootFilter is used,
		// so storage layer knows to use RootFilter.
		// Or, the storage layer function signature will explicitly take RootFilter
		// and ignore these if RootFilter is non-nil.
		// For now, pass them; storage will be adapted.
		// parseConditions would use query.Conditions, so if RootFilter is set,
		// we effectively want to pass empty legacy conditions.
		// However, GetEntitiesByQueryFilter will be modified to accept RootFilter
		// and prioritize it.
	}

	// now we need to fetch the list of entities fitting to our filters
	// Signature of GetEntitiesByQueryFilter will be adapted in storage.go
	// to accept query.RootFilter and prioritize it if non-nil.
	var resultData []transport.TransportEntity
	var resultAddresses [][2]int
	var amount int

	if query.RootFilter == nil {
		// Use legacy conditions, pass nil for new RootFilter parameter
		resultData, resultAddresses, amount = store.GetEntitiesByQueryFilter(query.Pool, nil, legacyConditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, returnDataFlag)
	} else {
		// Use new RootFilter, pass nil for legacy condition parameters
		resultData, resultAddresses, amount = store.GetEntitiesByQueryFilter(query.Pool, query.RootFilter, nil, nil, nil, nil, nil, returnDataFlag)
	}

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

	performOperation := true // General flag, true by default

	// Specifically for UPDATE with required joins: all-or-nothing validation
	if query.Method == METHOD_UPDATE && query.HasRequiredSubQueries() && 0 < len(query.Map) {
		for _, entityAddress := range resultAddresses {
			// Call recursiveExecuteLinked primarily to get subAmount for validation.
			// Pass an empty/fresh addressPairList ([][4]int{}) to avoid side effects on any shared slice
			// if recursiveExecuteLinked modifies the slice it receives.
			_, _, _, subAmount := recursiveExecuteLinked(store, query.Map, entityAddress, [][4]int{})
			if subAmount == 0 { // subAmount is 0 if a required join within query.Map fails for this entityAddress
				performOperation = false
				break
			}
		}
		if !performOperation {
			// If validation failed for an UPDATE query, no entities will be processed.
			ret.Amount = 0 // Signify no operation performed
			mutexh.Release()
			return ret // Update aborted
		}
	}

	// Proceed with join processing for READ, LINK, UNLINK, or for UPDATE if validation passed
	// This section will populate ret.Entities for READ, addressPairs for UNLINK, linkAddresses for LINK
	// and set ret.Amount based on entities that satisfy their joins.

	finalEntitiesForRead := []transport.TransportEntity{}
	finalAddressPairsForUnlink := [][4]int{}   // For METHOD_UNLINK
	finalLinkAddressesForLink := [2][][2]int{} // For METHOD_LINK, [direction][entityAddresses]
	countOfEntitiesPassingJoins := 0

	if 0 < len(query.Map) { // If there are joins
		if query.Method == METHOD_LINK { // Special handling for METHOD_LINK
			// Populate finalLinkAddressesForLink based on targetQuery matches
			// This is the original logic for METHOD_LINK target collection
			for _, targetQuery := range query.Map {
				var tBaseMatchList [3][][]int
				var tPropertyMatchList []map[string][]int
				var tLegacyConditions [][][3]string
				if targetQuery.RootFilter == nil {
					tLegacyConditions = targetQuery.Conditions
					tBaseMatchList, tPropertyMatchList = parseConditions(&targetQuery)
				}

				var tResultData []transport.TransportEntity
				var tmpLinkAddresses [][2]int
				var tmpLinkAmount int
				if targetQuery.RootFilter == nil {
					tResultData, tmpLinkAddresses, tmpLinkAmount = store.GetEntitiesByQueryFilter(targetQuery.Pool, nil, tLegacyConditions, tBaseMatchList[FILTER_ID], tBaseMatchList[FILTER_VALUE], tBaseMatchList[FILTER_CONTEXT], tPropertyMatchList, false)
				} else {
					tResultData, tmpLinkAddresses, tmpLinkAmount = store.GetEntitiesByQueryFilter(targetQuery.Pool, targetQuery.RootFilter, nil, nil, nil, nil, nil, false)
				}
				_ = tResultData // Explicitly ignore if not used, as per original logic

				if 0 < tmpLinkAmount {
					finalLinkAddressesForLink[targetQuery.Direction] = append(finalLinkAddressesForLink[targetQuery.Direction], tmpLinkAddresses...)
					// linkAmount was a local var in original, its sum contributes to final link operation
				}
			}
			// For METHOD_LINK, the number of primary entities to link *from* is the initial `amount`.
			// The actual number of links made will be calculated later.
			countOfEntitiesPassingJoins = amount
		} else { // For READ, UPDATE (if validated and performOperation is true), UNLINK
			for key, entityAddress := range resultAddresses {
				// Pass a fresh/empty addressPairList to recursiveExecuteLinked to avoid side effects,
				// as it might modify the slice it receives. tmpAddressPairs will be specific to this entityAddress.
				children, parents, tmpAddressPairs, subAmount := recursiveExecuteLinked(store, query.Map, entityAddress, [][4]int{})

				if query.HasRequiredSubQueries() && subAmount == 0 {
					// This entity failed a required join. Skip it for READ/UNLINK.
					// For UPDATE, this case should have been caught by the `performOperation` check earlier,
					// so we wouldn't be in this loop if performOperation was false.
					continue
				}

				// This entity passed its required joins (or had no required ones).
				countOfEntitiesPassingJoins++
				if query.Method == METHOD_READ {
					// resultData[key] corresponds to entityAddress.
					currentEntityData := resultData[key] // Use the data fetched initially
					if 0 < len(children) {
						currentEntityData.ChildRelations = append(currentEntityData.ChildRelations, children...)
					}
					if 0 < len(parents) {
						currentEntityData.ParentRelations = append(currentEntityData.ParentRelations, parents...)
					}
					finalEntitiesForRead = append(finalEntitiesForRead, currentEntityData)
				}
				if query.Method == METHOD_UNLINK {
					finalAddressPairsForUnlink = append(finalAddressPairsForUnlink, tmpAddressPairs...)
				}
			}
		}
	} else { // No joins
		countOfEntitiesPassingJoins = amount
		if query.Method == METHOD_READ {
			finalEntitiesForRead = resultData
		}
	}

	// Set ret.Amount based on processing.
	// For UPDATE that passed validation, countOfEntitiesPassingJoins will be original `amount`.
	// For READ/UNLINK, it's entities that passed joins.
	// For LINK, it's initial `amount` of source entities.
	ret.Amount = countOfEntitiesPassingJoins
	if query.Method == METHOD_READ {
		ret.Entities = finalEntitiesForRead
	}

	// Perform actual data modification or finalize read results
	if 0 < ret.Amount { // If any entities are left to operate on or return
		switch query.Method {
		case METHOD_UPDATE:
			// The `performOperation` check already handled the all-or-nothing for required joins.
			// If we are here, it means either no required joins, or all entities met them.
			// `resultAddresses` is the list of *all initially matched entities*. This is correct.
			if 0 < len(query.Values) {
				store.BatchUpdateAddressList(resultAddresses, query.Values)
				// ret.Amount for UPDATE should reflect the number of entities attempted to update,
				// which is the original `amount` if validation passed.
				// countOfEntitiesPassingJoins would be `amount` in this scenario.
				ret.Amount = amount // Ensure ret.Amount reflects all entities if update proceeded.
			}
		case METHOD_DELETE:
			// Original behavior: operates on all initially matched `resultAddresses`.
			// Task is specific to UPDATE, so DELETE logic remains unchanged unless specified.
			store.BatchDeleteAddressList(resultAddresses)
			// ret.Amount for DELETE should reflect number of entities deleted.
			// If joins were to filter deletions, this would need adjustment.
			// For now, it's `countOfEntitiesPassingJoins` (which is `amount` if no joins).
		case METHOD_LINK:
			affectedAmount := 0
			// Use finalLinkAddressesForLink collected earlier.
			// resultAddresses are the source entities for the links.
			if 0 < len(finalLinkAddressesForLink[DIRECTION_CHILD])+len(finalLinkAddressesForLink[DIRECTION_PARENT]) {
				for direction, tmpLinkAddresses := range finalLinkAddressesForLink {
					if 0 < len(tmpLinkAddresses) {
						if DIRECTION_CHILD == direction {
							affectedAmount += store.LinkAddressLists(resultAddresses, tmpLinkAddresses, query.LinkRelationContext, query.LinkRelationProperties)
						} else { // else it must be towards parent so we flip params
							affectedAmount += store.LinkAddressLists(tmpLinkAddresses, resultAddresses, query.LinkRelationContext, query.LinkRelationProperties)
						}
					}
				}
			}
			ret.Amount = affectedAmount // Actual number of links created.
		case METHOD_UNLINK:
			affectedAmount := 0
			if 0 < len(finalAddressPairsForUnlink) { // Use the collected pairs
				for _, addressPair := range finalAddressPairsForUnlink {
					// Check if relation matches context and property filters before deleting
					relationShouldBeDeleted := true

					// Fetch the relation to check its context and properties
					// addressPair: [sourceType, sourceID, targetType, targetID]
					relation, err := store.GetRelationUnsafe(addressPair[0], addressPair[1], addressPair[2], addressPair[3])
					if err != nil {
						// Relation doesn't exist or error fetching, skip.
						// This might happen if another concurrent operation deleted it.
						continue
					}

					// Apply UnlinkRelationContextFilter
					if query.UnlinkRelationContextFilter != "" {
						if relation.Context != query.UnlinkRelationContextFilter {
							relationShouldBeDeleted = false
						}
					}

					// Apply UnlinkRelationPropertyFilters if relationShouldBeDeleted is still true
					if relationShouldBeDeleted && len(query.UnlinkRelationPropertyFilters) > 0 {
						for _, propFilter := range query.UnlinkRelationPropertyFilters {
							propValue, propExists := relation.Properties[propFilter.Key]
							if !propExists {
								relationShouldBeDeleted = false // Property to filter on doesn't exist on relation
								break
							}
							// Use storage.match (or a similar helper if storage.match is not directly accessible/suitable)
							// For now, assuming storage.match can be used or adapted.
							// We need to make s.match accessible or replicate its logic here.
							// For simplicity, let's assume we have a way to call it:
							// match(valueFromRelation, operatorFromFilter, valueFromFilter)
							// This part needs careful implementation of the match logic.
							// Let's create a local helper or assume store.Match is available.
							// For now, direct comparison for "==" operator as placeholder:
							if !store.Match(propValue, propFilter.Operator, propFilter.Value) { // Assuming store.Match exists and is accessible
								relationShouldBeDeleted = false
								break
							}
						}
					}

					if relationShouldBeDeleted {
						store.DeleteRelationUnsafe(addressPair[0], addressPair[1], addressPair[2], addressPair[3])
						affectedAmount++
					}
				}
			}
			ret.Amount = affectedAmount // Actual number of relations unlinked.
		case METHOD_READ:
			// ret.Entities is already set. Apply traversal.
			if direction, depth, traversed := isTraversed(*query); traversed {
				for id := range ret.Entities { // Iterate over the actual entities being returned
					store.TraverseEnrich(&(ret.Entities[id]), direction, depth)
				}
			}
		}

		// Post-processing for READ queries
		if query.Method == METHOD_READ {
			if (Order{}) != query.Sort {
				ret.Entities = sortResults(ret.Entities, query.Sort.Field, query.Sort.Direction, query.Sort.Mode)
			}

			limit := getLimitIfExists(*query)
			if -1 != limit {
				if len(ret.Entities) > limit {
					ret.Entities = ret.Entities[:limit]
				}
			}
		}

		mutexh.Release()
		return ret
	}

	// if ret.Amount is 0 (either no initial hits, or update aborted, or no entities passed joins for read/unlink)
	mutexh.Release()
	return transport.Transport{} // Return empty transport
}

func recursiveExecuteLinked(store *storage.Storage, queries []Query, sourceAddress [2]int, addressPairList [][4]int) ([]transport.TransportRelation, []transport.TransportRelation, [][4]int, int) {
	var retParents []transport.TransportRelation
	var retChildren []transport.TransportRelation
	i := 0
	for _, currentQuery := range queries { // Renamed query to currentQuery to avoid conflict
		var tmpRet []transport.TransportRelation
		// parse the conditions into our 2 necessary groups
		var baseMatchList [3][][]int
		var propertyMatchList []map[string][]int
		var legacyConditions [][][3]string

		if currentQuery.RootFilter == nil {
			legacyConditions = currentQuery.Conditions
			baseMatchList, propertyMatchList = parseConditions(&currentQuery)
		}

		// do we need to return the data itself?
		returnDataFlag := false
		if METHOD_READ == currentQuery.Method { // Use currentQuery
			returnDataFlag = true
		}

		// get data from subquery
		// Signature of GetEntitiesByQueryFilterAndSourceAddress will be adapted in storage.go
		var resultData []transport.TransportRelation
		var resultAddresses [][2]int
		var amount int

		if currentQuery.RootFilter == nil {
			resultData, resultAddresses, amount = store.GetEntitiesByQueryFilterAndSourceAddress(currentQuery.Pool, nil, legacyConditions, baseMatchList[FILTER_ID], baseMatchList[FILTER_VALUE], baseMatchList[FILTER_CONTEXT], propertyMatchList, sourceAddress, currentQuery.Direction, returnDataFlag)
		} else {
			resultData, resultAddresses, amount = store.GetEntitiesByQueryFilterAndSourceAddress(currentQuery.Pool, currentQuery.RootFilter, nil, nil, nil, nil, nil, sourceAddress, currentQuery.Direction, returnDataFlag)
		}

		// if we got no returns
		if 0 == amount {
			// we check if there had to be some
			if true == currentQuery.Required { // Fixed: undefined 'query' to 'currentQuery'
				// empty return since we got no hits on a required subquery
				return []transport.TransportRelation{}, []transport.TransportRelation{}, [][4]int{}, 0
			}
			// if not we just continue
			continue
		}
		// since we got data we gonne get recursive from here
		if 0 < len(currentQuery.Map) { // Use currentQuery
			collectAddressList := [][4]int{}
			for key, entityAddress := range resultAddresses {
				// further execute and store data on return
				children, parents, tmpAddressList, subAmount := recursiveExecuteLinked(store, currentQuery.Map, entityAddress, addressPairList) // Pass currentQuery.Map
				// Determine relation direction based on currentQuery.Direction
				if DIRECTION_CHILD == currentQuery.Direction {
					tmpAddressList = append(tmpAddressList, [4]int{sourceAddress[0], sourceAddress[1], entityAddress[0], entityAddress[1]})
				} else {
					tmpAddressList = append(tmpAddressList, [4]int{entityAddress[0], entityAddress[1], sourceAddress[0], sourceAddress[1]})
				}

				collectAddressList = append(collectAddressList, tmpAddressList...)
				if 0 < len(children) {
					resultData[key].Target.ChildRelations = append(resultData[key].Target.ChildRelations, children...)
				}
				if 0 < len(parents) {
					resultData[key].Target.ParentRelations = append(resultData[key].Target.ParentRelations, parents...)
				}
				if subAmount > 0 || !currentQuery.HasRequiredSubQueries() { // Check subAmount and use currentQuery
					// only add results if we actually are returning data
					if returnDataFlag {
						tmpRet = append(tmpRet, resultData[key])
					}
					i++ // Increment main counter i based on successful processing of sub-query results
				}
			}
			addressPairList = append(addressPairList, collectAddressList...)
		} else {
			// there must be a smarter way for the following problem:
			for _, entityAddress := range resultAddresses {
				if DIRECTION_CHILD == currentQuery.Direction { // Use currentQuery
					addressPairList = append(addressPairList, [4]int{sourceAddress[0], sourceAddress[1], entityAddress[0], entityAddress[1]})
				} else {
					addressPairList = append(addressPairList, [4]int{entityAddress[0], entityAddress[1], sourceAddress[0], sourceAddress[1]})
				}
			}
			// - - - - - - - - - - - - - - - - - -
			i += amount // Increment main counter i by the number of direct results
			tmpRet = append(tmpRet, resultData...)
		}

		// if we got any results we add them
		tmpRetLen := len(tmpRet)
		if 0 < tmpRetLen {
			var appender *[]transport.TransportRelation
			if DIRECTION_CHILD == currentQuery.Direction { // Use currentQuery
				appender = &retChildren
			} else {
				appender = &retParents
			}
			start := len(*appender)
			*appender = append(*appender, tmpRet...)
			if direction, depth, ok := isTraversed(currentQuery); ok { // Use currentQuery
				for idx := start; idx < start+tmpRetLen; idx++ { // Use different loop variable idx
					store.TraverseEnrich(&((*appender)[idx].Target), direction, depth)
				}
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

func isTraversed(qry Query) (int, int, bool) {
	if nil != qry.Mode {
		for _, mode := range qry.Mode {
			tmpLen := len(mode)
			if 0 < tmpLen && "Traverse" == mode[0] {
				if 3 == tmpLen {
					direction, err := strconv.ParseInt(mode[1], 10, 64)
					if nil != err {
						// archivist.Info("Invalid traverse direction given. Skipping") ###todo overthink if false should be err and we return that info somehoow
						return -1, -1, false
					}
					depth, err := strconv.ParseInt(mode[2], 10, 64)
					if nil != err {
						// archivist.Info("Invalid traverse depth given. Skipping") ###todo overthink if false should be err and we return that info somehoow
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
-> LIMIT                [ ]



*/
