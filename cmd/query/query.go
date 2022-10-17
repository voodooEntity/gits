package main

import (
	"encoding/json"
	"fmt"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
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
	//testSimpleReadWithReduce()
	testUpdateEntityValue()
	//testDeleteEntityByTypeAndID()
	//testQueryLinkTo()
	//testQueryLinkFrom()
	//testQueryUnlink()
	//testQueryUnlinkReverse()
	//buildTestQueryJson()
	//buildTestQueryJson2()
	//buildTestQueryJson3()
	//buildTestQueryJsonGetQbQueries()
	//testOptionalQueryJoinFirstLevel()
	//testRequiredQueryJoinFirstLevelSuccess()
	//testRequiredQueryJoinFirstLevelFail()
	//testRequiredQueryJoinInDepthFail()
	//testRequiredQueryJoinInDepthSuccess()
	//testRequiredAndOptionalMixed()
	//testOrderByNumericValueAsc()
	//testOrderByNumericValueDesc()
	//testOrderByAlphabeticalValueAsc()
	//testOrderByAlphabeticalValueDesc()
	//testSpecificQueryContentCompareJs()
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
	qry := query.New().Read("Beta").To(
		query.New().Read("Delta"),
	).From(
		query.New().Read("Alpha"),
	)
	result := query.Execute(qry)
	printData(result)
}

func testBidrectionalJoinAndTurn() {
	createTestDataLinked()
	qry := query.New().Read("Beta").To(
		query.New().Read("Delta"),
	).From(
		query.New().Read("Alpha").To(
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
	qry := query.New().Read("Signer").Match("Value", "=", "asdasdasdasd").From(
		query.New().Read("CollectionOffer", "TokenOffer").To(
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
	qry := query.New().Read("Beta").To(
		query.New().Read("Delta"),
	).From(
		query.New().Read("Alpha").Match("Context", "==", "uno").Match("Value", "==", "alpha"),
	)
	result := query.Execute(qry)
	printData(result)
}

func testSingleJoinChild() {
	createTestDataLinked()
	qry := query.New().Read("Alpha").To(
		query.New().Read("Beta"),
	)
	result := query.Execute(qry)
	printData(result)
}

func testDoubleDepthJoinChild() {
	createTestDataLinked()
	qry := query.New().Read("Alpha").To(
		query.New().Read("Beta").To(
			query.New().Read("Delta"),
		),
	)
	result := query.Execute(qry)
	printData(result)
}

func testSimpleReadWithReduce() {
	createTestDataLinked()
	qry := query.New().Read("Alpha").To(
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

func testUpdateEntityValue() {
	gits.CreateEntityType("Test")
	gits.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	qry := query.New().Read("Test")
	ret := query.Execute(qry)
	printData(ret)
	qry = query.New().Update("Test").Match("Value", "==", "TestABC").Set("Value", "TestDEF").Set("Context", "asdasdasd")
	query.Execute(qry)
	qry = query.New().Read("Test")
	ret = query.Execute(qry)
	printData(ret)
}

func testDeleteEntityByTypeAndID() {
	gits.CreateEntityType("Test")
	entity := gits.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	qry := query.New().Read("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	ret := query.Execute(qry)
	printData(ret)
	qry = query.New().Delete("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	query.Execute(qry)
	qry = query.New().Read("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	ret = query.Execute(qry)
	printData(ret)
}

func testQueryLinkTo() {
	// create testdata
	gits.CreateEntityType("Test")
	gits.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestABC",
	})
	gits.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestDEF",
	})

	// print the testdata before linking
	qry := query.New().Read("Test")
	printData(query.Execute(qry))

	// link the datasets
	qry = query.New().Link("Test").Match("Value", "==", "TestABC").To(
		query.New().Find("Test").Match("Value", "==", "TestDEF"),
	)
	query.Execute(qry)

	// now read out to approve we gotr the linked data
	qry = query.New().Read("Test").Match("Value", "==", "TestABC").To(
		query.New().Read("Test"),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func testQueryLinkFrom() {
	// create testdata
	gits.CreateEntityType("Test")
	gits.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestABC",
	})
	gits.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestDEF",
	})

	// print the testdata before linking
	qry := query.New().Read("Test")
	printData(query.Execute(qry))

	// link the datasets
	qry = query.New().Link("Test").Match("Value", "==", "TestABC").From(
		query.New().Find("Test").Match("Value", "==", "TestDEF"),
	)
	query.Execute(qry)

	// now read out to approve we gotr the linked data
	qry = query.New().Read("Test").Match("Value", "==", "TestABC").From(
		query.New().Read("Test"),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func testQueryUnlink() {
	gits.CreateEntityType("TestA")
	gits.CreateEntityType("TestB")
	testdata := transport.TransportEntity{
		ID:    -1,
		Type:  "TestA",
		Value: "Something",
		ChildRelations: []transport.TransportRelation{
			{
				Context: "Something",
				Target: transport.TransportEntity{
					ID:    -1,
					Type:  "TestB",
					Value: "Else",
				},
			},
		},
	}
	gits.MapTransportData(testdata)

	// read linked inserted data
	qry := query.New().Read("TestA").Match("Value", "==", "Something").To(
		query.New().Read("TestB").Match("Value", "==", "Else"),
	)
	printData(query.Execute(qry))

	// unlink the data
	qry = query.New().Unlink("TestA").Match("Value", "==", "Something").To(
		query.New().Find("TestB").Match("Value", "==", "Else"),
	)
	query.Execute(qry)

	// read linked inserted data
	qry = query.New().Read("TestA").Match("Value", "==", "Something").To(
		query.New().Read("TestB").Match("Value", "==", "Else"),
	)
	printData(query.Execute(qry))

}

func testQueryUnlinkReverse() {
	gits.CreateEntityType("TestA")
	gits.CreateEntityType("TestB")
	testdata := transport.TransportEntity{
		ID:    -1,
		Type:  "TestB",
		Value: "Else",
		ChildRelations: []transport.TransportRelation{
			{
				Context: "Something",
				Target: transport.TransportEntity{
					ID:    -1,
					Type:  "TestA",
					Value: "Something",
				},
			},
		},
	}
	gits.MapTransportData(testdata)

	// read linked inserted data
	qry := query.New().Read("TestA").Match("Value", "==", "Something").From(
		query.New().Read("TestB").Match("Value", "==", "Else"),
	)
	printData(query.Execute(qry))

	// unlink the data
	qry = query.New().Unlink("TestA").Match("Value", "==", "Something").From(
		query.New().Find("TestB").Match("Value", "==", "Else"),
	)
	query.Execute(qry)

	// read linked inserted data
	qry = query.New().Read("TestA").Match("Value", "==", "Something").From(
		query.New().Read("TestB").Match("Value", "==", "Else"),
	)
	printData(query.Execute(qry))

}

func buildTestQueryJson() {
	qry := query.New().Read("IP").To(
		query.New().Read("Port").To(
			query.New().Read("Software"),
		),
	)
	printData(qry)
}

func buildTestQueryJson2() {
	qry := query.New().Read("IP").To(
		query.New().Read("Port").To(
			query.New().Read("Software").To(
				query.New().Read("Vhost"),
			),
		),
	)
	printData(qry)
}

func buildTestQueryJson3() {
	qry := query.New().Read("IP").To(
		query.New().Read("Port").To(
			query.New().Read("Software").To(
				query.New().Read("Vhost"),
			),
		).To(
			query.New().Read("Software"),
		),
	)
	printData(qry)
}

func buildTestQueryJsonGetQbQueries() {
	//
	archivist.Info("Get all marketplaces implemented by Max Mustermann from person")
	qry := query.New().Read("Person").Match("Value", "==", "Max Mustermann").To(
		query.New().Read("Marketplace").Match("Properties.IsAbstract", "==", "false"),
	).To(
		query.New().Read("Marketplace").To(
			query.New().Read("Marketplace"),
		),
	)
	//printData(qry)

	//
	archivist.Info("Get all marketplaces shipping to germany")
	qry = query.New().Read("Marketplace").To(
		query.New().Reduce("Country").Match("Value", "==", "Germany"),
	)
	//printData(qry)

	//
	archivist.Info("Get all marketplaces ")
	qry = query.New().Read("Person").Match("Value", "==", "Max Mustermann").To(
		query.New().Read("Marketplace"),
	).To(
		query.New().Read("Marketplace").To(
			query.New().Read("Marketplace"),
		),
	)
	printData(qry)

	archivist.Info("Get Person that implemented marketplace")
	qry = query.New().Read("Marketplace").From(
		query.New().Read("Person"),
	)
	//printData(qry)
}

func testRequiredAndOptionalMixed() {
	// create the testdata
	testdata := testQbStructureMap()
	gits.MapTransportData(testdata)

	archivist.Info(" - - - - - - - - - - - - -  Test forced 2 depth marketplace - - - - - - - - - - - - - -")
	//qry := query.New().Read("Person").To(
	//	query.New().Read("Marketplace").To(
	//		query.New().Read("Marketplace"),
	//	).To(
	//		query.New().Read("Marketplace"),
	//	),
	//)
	//ret := query.Execute(qry)
	//printData(ret)
	archivist.Info(" - - - - - - - - - Test forced 1 depth marketplace and 2nmd depth optional  - - - - - - - - -")
	qry := query.New().Read("Person").To(
		query.New().Read("Marketplace").CanTo(
			query.New().Read("Marketplace"),
		),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func testOptionalQueryJoinFirstLevel() {
	testdata := testQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test optional first level join  - - - - - - - - -")
	qry := query.New().Read("Person").CanTo(
		query.New().Read("Shipping"),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func testRequiredQueryJoinFirstLevelSuccess() {
	testdata := testQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test required first level join  - - - - - - - - -")
	qry := query.New().Read("Person").To(
		query.New().Read("Marketplace"),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func testRequiredQueryJoinFirstLevelFail() {
	testdata := testQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test required first level join  - - - - - - - - -")
	qry := query.New().Read("Person").To(
		query.New().Read("Shipping"),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func testRequiredQueryJoinInDepthFail() {
	testdata := testQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test required first level join  - - - - - - - - -")
	qry := query.New().Read("Person").To(
		query.New().Read("Marketplace").To(
			query.New().Read("Person"),
		),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func testRequiredQueryJoinInDepthSuccess() {
	testdata := testQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test required first level join  - - - - - - - - -")
	qry := query.New().Read("Person").To(
		query.New().Read("Marketplace").To(
			query.New().Read("Country"),
		),
	)
	ret := query.Execute(qry)
	printData(ret)
}

func printData(data any) {
	t, _ := json.MarshalIndent(data, "", "\t")
	archivist.Info("Query Data Struct", string(t))
}

func testQbStructureMap() transport.TransportEntity {
	archivist.Info("Print testdata")
	// create testdata
	testdata := transport.TransportEntity{
		ID:    -1,
		Type:  "Person",
		Value: "Max Mustermann",
		ChildRelations: []transport.TransportRelation{
			{
				Context: "Implemented",
				Target: transport.TransportEntity{
					ID:         -1,
					Type:       "Marketplace",
					Value:      "Gabor",
					Properties: map[string]string{"IsAbstract": "false"},
					ChildRelations: []transport.TransportRelation{
						{
							Context: "ShipsTo",
							Target: transport.TransportEntity{
								ID:    -1,
								Type:  "Country",
								Value: "Germany",
							},
						}, {
							Context: "ShipsTo",
							Target: transport.TransportEntity{
								ID:    -1,
								Type:  "Country",
								Value: "France",
							},
						},
					},
				},
			}, {
				Context: "Implemented",
				Target: transport.TransportEntity{
					ID:    -1,
					Type:  "Marketplace",
					Value: "Wortmann",
					ChildRelations: []transport.TransportRelation{
						{
							Context: "Defers",
							Target: transport.TransportEntity{
								ID:         -1,
								Type:       "Marketplace",
								Value:      "Marco Tozzi",
								Properties: map[string]string{"Golive": "1.1.1001"},
								ChildRelations: []transport.TransportRelation{
									{
										Context: "ShipsTo",
										Target: transport.TransportEntity{
											ID:    -1,
											Type:  "Country",
											Value: "Germany",
										},
									}, {
										Context: "ShipsTo",
										Target: transport.TransportEntity{
											ID:    -1,
											Type:  "Country",
											Value: "Austria",
										},
									},
								},
							},
						}, {
							Context: "Defers",
							Target: transport.TransportEntity{
								ID:         -1,
								Type:       "Marketplace",
								Value:      "Tamaris",
								Properties: map[string]string{"Golive": "1.1.2001"},
								ChildRelations: []transport.TransportRelation{
									{
										Context: "ShipsTo",
										Target: transport.TransportEntity{
											ID:    -1,
											Type:  "Country",
											Value: "Germany",
										},
									}, {
										Context: "ShipsTo",
										Target: transport.TransportEntity{
											ID:    -1,
											Type:  "Country",
											Value: "Austria",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	printData(testdata)
	return testdata
}

func createTestDataForOrderByValueNumeric() {
	gits.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "2",
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "333",
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "1",
	})
}

func createTestDataOrderAlphabetical() {
	gits.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "Das",
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "Zebra",
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "auch",
	})
}

func testOrderByNumericValueAsc() {
	createTestDataForOrderByValueNumeric()
	qry := query.New().Read("Something").Order("Value", query.ORDER_DIRECTION_ASC, query.ORDER_MODE_NUM)
	ret := query.Execute(qry)
	printData(ret)
}

func testOrderByNumericValueDesc() {
	createTestDataForOrderByValueNumeric()
	qry := query.New().Read("Something").Order("Value", query.ORDER_DIRECTION_DESC, query.ORDER_MODE_NUM)
	ret := query.Execute(qry)
	fmt.Printf("%+v", ret)
	printData(ret)
}

func testOrderByAlphabeticalValueAsc() {
	createTestDataOrderAlphabetical()
	qry := query.New().Read("Something").Order("Value", query.ORDER_DIRECTION_ASC, query.ORDER_MODE_ALPHA)
	ret := query.Execute(qry)
	printData(ret)
}

func testOrderByAlphabeticalValueDesc() {
	createTestDataOrderAlphabetical()
	qry := query.New().Read("Something").Order("Value", query.ORDER_DIRECTION_DESC, query.ORDER_MODE_ALPHA)
	ret := query.Execute(qry)
	printData(ret)
}

func testSpecificQueryContentCompareJs() {
	qry := query.New().Read("Note").Match("Value", "contain", "something").OrMatch("Property.Text", "contain", "something").OrMatch("Property.Date", "contain", "something")
	printData(qry)
}
