package main

import (
	"encoding/json"
	"fmt"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/types"
	"strconv"
	"time"
)

func main() {
	// init our logging
	archivist.Init("info", "stdout", "blafu")

	// init the gits
	gits.Init(types.PersistenceConfig{
		RotationEntriesMax:           1000000,
		Active:                       false,
		PersistenceChannelBufferSize: 10000000,
	})

	start := time.Now()
	//testSimpleRead()
	//testSingleJoinChild()
	//testBidirectionalJoin()
	//testBidrectionalJoinAndTurn()
	//testSimpleReadMultiPool()
	//testSimpleReadMultiPoolWithOrMatch()
	//testBidirectionalJoinWithCondition()
	//testDoubleDepthJoinChild()
	//testFilterValueByExcactMatch()
	//testFilterValueByGreaterThanMatch()
	//testFilterValueBySmallerThanMatch()
	//testFilterPropertyByExcactMatch()
	testSimpleReadWithReduce()
	fmt.Println("Time took ", time.Since(start))
}

func createTestDataLinked() {

	// create test data to query
	typeIDalpha, _ := gits.CreateEntityType("Alpha")
	typeIDbeta, _ := gits.CreateEntityType("Beta")
	typeIDdelta, _ := gits.CreateEntityType("Delta")
	typeIDgamma, _ := gits.CreateEntityType("Gamma")

	entityAlphaID, _ := gits.CreateEntity(types.StorageEntity{
		Type:    typeIDalpha,
		Value:   "alpha",
		Context: "uno",
	})

	entityBetaID, _ := gits.CreateEntity(types.StorageEntity{
		Type:    typeIDbeta,
		Value:   "beta",
		Context: "duo",
	})

	entityDeltaID, _ := gits.CreateEntity(types.StorageEntity{
		Type:    typeIDdelta,
		Value:   "delta",
		Context: "tres",
	})

	entityGammaID, _ := gits.CreateEntity(types.StorageEntity{
		Type:    typeIDgamma,
		Value:   "gamma",
		Context: "quattro",
	})

	//printData(gits.EntityStorage)

	gits.CreateRelation(typeIDalpha, entityAlphaID, typeIDbeta, entityBetaID, types.StorageRelation{
		SourceType: typeIDalpha,
		SourceID:   entityAlphaID,
		TargetType: typeIDbeta,
		TargetID:   entityBetaID,
	})

	gits.CreateRelation(typeIDbeta, entityBetaID, typeIDdelta, entityDeltaID, types.StorageRelation{
		SourceType: typeIDbeta,
		SourceID:   entityBetaID,
		TargetType: typeIDdelta,
		TargetID:   entityDeltaID,
	})

	gits.CreateRelation(typeIDalpha, entityAlphaID, typeIDgamma, entityGammaID, types.StorageRelation{
		SourceType: typeIDalpha,
		SourceID:   entityAlphaID,
		TargetType: typeIDgamma,
		TargetID:   entityGammaID,
	})
}

func testFilterValueByExcactMatch() {
	createTestDataLinearTypeNumericValue()
	qry := query.New().Read("Alpha").Match("Value", "==", "42")
	result := query.Execute(qry)
	printData(result)
}

func testFilterValueBySmallerThanMatch() {
	createTestDataLinearTypeNumericValue()
	qry := query.New().Read("Alpha").Match("Value", "<", "3")
	result := query.Execute(qry)
	printData(result)
}

func testFilterValueByGreaterThanMatch() {
	createTestDataLinearTypeNumericValue()
	qry := query.New().Read("Alpha").Match("Value", ">", "97")
	result := query.Execute(qry)
	printData(result)
}

func testFilterPropertyByExcactMatch() {
	createTestDataLinearTypeNumericPropertyTestValue()
	qry := query.New().Read("Alpha").Match("Properties.Test", "==", "42")
	result := query.Execute(qry)
	printData(result)
}

func createTestDataLinearTypeNumericValue() {
	typeIDalpha, _ := gits.CreateEntityType("Alpha")
	for i := 1; i <= 100; i++ {
		gits.CreateEntity(types.StorageEntity{
			Type:    typeIDalpha,
			Value:   strconv.Itoa(i),
			Context: "uno",
		})
	}
}

func createTestDataLinearTypeNumericPropertyTestValue() {
	typeIDalpha, _ := gits.CreateEntityType("Alpha")
	for i := 1; i <= 100; i++ {
		props := make(map[string]string)
		props["Test"] = strconv.Itoa(i)
		gits.CreateEntity(types.StorageEntity{
			Type:       typeIDalpha,
			Value:      "alpha",
			Context:    "uno",
			Properties: props,
		})
	}
}

func testBidirectionalJoin() {
	createTestDataLinked()
	qry := query.New().Read("Beta").Join(
		query.New().Read("Delta"),
	).RJoin(
		query.New().Read("Alpha"),
	)
	result := query.Execute(qry)
	printData(result)
}

func testBidrectionalJoinAndTurn() {
	createTestDataLinked()
	qry := query.New().Read("Beta").Join(
		query.New().Read("Delta"),
	).RJoin(
		query.New().Read("Alpha").Join(
			query.New().Read("Gamma").Read(),
		),
	)
	result := query.Execute(qry)
	printData(result)
}

func testSimpleReadMultiPool() {
	createTestDataLinked()
	qry := query.New().Read("Alpha", "Beta")
	result := query.Execute(qry)
	printData(result)
}

func testShowStephen() {
	createTestDataLinked()
	qry := query.New().Read("Signer").Match("Value", "=", "asdasdasdasd").RJoin(
		query.New().Read("CollectionOffer", "TokenOffer").Join(
			query.New().Read("Details"),
		),
	)
	result := query.Execute(qry)
	printData(result)
}

func testSimpleReadMultiPoolWithOrMatch() {
	createTestDataLinked()
	qry := query.New().Read("Alpha", "Beta").Match("Value", "==", "alpha").OrMatch("Value", "==", "beta")
	result := query.Execute(qry)
	printData(result)
}

func testBidirectionalJoinWithCondition() {
	createTestDataLinked()
	qry := query.New().Read("Beta").Join(
		query.New().Read("Delta"),
	).RJoin(
		query.New().Read("Alpha").Match("Context", "==", "uno").Match("Value", "==", "alpha"),
	)
	result := query.Execute(qry)
	printData(result)
}

func testSingleJoinChild() {
	createTestDataLinked()
	qry := query.New().Read("Alpha").Join(
		query.New().Read("Beta"),
	)
	result := query.Execute(qry)
	printData(result)
}

func testDoubleDepthJoinChild() {
	createTestDataLinked()
	qry := query.New().Read("Alpha").Join(
		query.New().Read("Beta").Join(
			query.New().Read("Delta"),
		),
	)
	result := query.Execute(qry)
	printData(result)
}

func testSimpleReadWithReduce() {
	createTestDataLinked()
	qry := query.New().Read("Alpha").Join(
		query.New().Reduce("Beta"),
	)
	result := query.Execute(qry)
	printData(result)
}

func testSimpleRead() {
	createTestDataLinked()
	qry := query.New().Read("Alpha")
	result := query.Execute(qry)
	printData(result)
}

func printData(data any) {
	t, _ := json.MarshalIndent(data, "", "\t")
	archivist.Info("Query Data Struct", string(t))
}
