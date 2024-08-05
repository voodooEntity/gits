package query

import (
	"encoding/json"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
	"strconv"
	"testing"
)

func initStorage() {
	archivist.Init("info", "stdout", "blafu")

	// init the gits
	gits.Init(types.PersistenceConfig{
		RotationEntriesMax:           1000000,
		Active:                       false,
		PersistenceChannelBufferSize: 10000000,
	})
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

func TestFilterValueByExcactMatch(t *testing.T) {
	initStorage()
	createTestDataLinearTypeNumericValue()
	qry := New().Read("Alpha").Match("Value", "==", "42")
	result := Execute(qry)
	if 1 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestFilterValueBySmallerThanMatch(t *testing.T) {
	initStorage()
	createTestDataLinearTypeNumericValue()
	qry := New().Read("Alpha").Match("Value", "<", "3")
	result := Execute(qry)
	if 2 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestFilterValueByGreaterThanMatch(t *testing.T) {
	initStorage()
	createTestDataLinearTypeNumericValue()
	qry := New().Read("Alpha").Match("Value", ">", "97")
	result := Execute(qry)
	if 3 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestFilterPropertyByExcactMatch(t *testing.T) {
	initStorage()
	createTestDataLinearTypeNumericPropertyTestValue()
	qry := New().Read("Alpha").Match("Properties.Test", "==", "42")
	result := Execute(qry)
	if 1 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
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

func TestBidirectionalJoin(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Beta").To(
		New().Read("Delta"),
	).From(
		New().Read("Alpha"),
	)
	result := Execute(qry)
	if 1 != len(result.Entities) || 1 != len(result.Entities[0].ChildRelations) || 1 != len(result.Entities[0].ParentRelations) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestBidrectionalJoinAndTurn(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Beta").To(
		New().Read("Delta"),
	).From(
		New().Read("Alpha").To(
			New().Read("Gamma").Read(),
		),
	)
	result := Execute(qry)
	if 1 != len(result.Entities) || 1 != len(result.Entities[0].ChildRelations) || 1 != len(result.Entities[0].ParentRelations[0].Target.ChildRelations) || 1 != len(result.Entities[0].ParentRelations) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestSimpleReadMultiPool(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Alpha", "Beta")
	result := Execute(qry)
	if 2 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestSimpleReadMultiPoolWithOrMatch(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Alpha", "Beta").Match("Value", "==", "alpha").OrMatch("Value", "==", "beta")
	result := Execute(qry)
	if 2 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestBidirectionalJoinWithCondition(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Beta").To(
		New().Read("Delta"),
	).From(
		New().Read("Alpha").Match("Context", "==", "uno").Match("Value", "==", "alpha"),
	)
	result := Execute(qry)
	if 1 != len(result.Entities) || 1 != len(result.Entities[0].ChildRelations) || 1 != len(result.Entities[0].ParentRelations) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestSingleJoinChild(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Alpha").To(
		New().Read("Beta"),
	)
	result := Execute(qry)
	if 1 != len(result.Entities) || 1 != len(result.Entities[0].ChildRelations) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestDoubleDepthJoinChild(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Alpha").To(
		New().Read("Beta").To(
			New().Read("Delta"),
		),
	)
	result := Execute(qry)
	if 1 != len(result.Entities) || 1 != len(result.Entities[0].ChildRelations) || 1 != len(result.Entities[0].ChildRelations[0].Target.ChildRelations) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestSimpleReadWithReduce(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Alpha").To(
		New().Reduce("Beta"),
	)
	result := Execute(qry)
	if 1 != len(result.Entities) && 0 == len(result.Entities[0].ChildRelations) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestSimpleRead(t *testing.T) {
	initStorage()
	createTestDataLinked()
	qry := New().Read("Alpha")
	result := Execute(qry)
	if 1 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestUpdateEntityValue(t *testing.T) {
	initStorage()
	gits.CreateEntityType("Test")
	gits.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	qry := New().Read("Test")
	ret := Execute(qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	qry = New().Update("Test").Match("Value", "==", "TestABC").Set("Value", "TestDEF").Set("Context", "asdasdasd")
	Execute(qry)
	qry = New().Read("Test")
	ret = Execute(qry)
	if 1 != len(ret.Entities) && "TestDEF" == ret.Entities[0].Value && "asdasdasd" == ret.Entities[0].Context {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestReadJoinMatchWithMultipleRequiredMatch(t *testing.T) {
	gits.MapTransportData(transport.TransportEntity{
		ID:      0,
		Type:    "Parent",
		Value:   "dad",
		Context: "dad",
		ChildRelations: []transport.TransportRelation{{
			Target: transport.TransportEntity{
				ID:      -1,
				Type:    "Test",
				Value:   "TestABC",
				Context: "TestABC",
			}},
		},
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:      0,
		Type:    "Parent",
		Value:   "dad",
		Context: "dad",
		ChildRelations: []transport.TransportRelation{{
			Target: transport.TransportEntity{
				ID:      -1,
				Type:    "Test",
				Value:   "TestDEF",
				Context: "TestABC",
			}},
		},
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:      0,
		Type:    "Parent",
		Value:   "dad",
		Context: "dad",
		ChildRelations: []transport.TransportRelation{{
			Target: transport.TransportEntity{
				ID:      -1,
				Type:    "Test",
				Value:   "TestABC",
				Context: "TestDEF",
			}},
		},
	})
	qry := New().Read("Parent").To(
		New().Read("Test").Match("Value", "==", "TestABC").Match("Context", "==", "TestABC"),
	)
	ret := Execute(qry)
	archivist.Info("data return", ret)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestFindValidTokenRequest(t *testing.T) {
	gits.MapTransportData(transport.TransportEntity{
		ID:    0,
		Type:  "User",
		Value: "testuser",
		ChildRelations: []transport.TransportRelation{{
			Target: transport.TransportEntity{
				ID:         -1,
				Type:       "Token",
				Value:      "findme",
				Context:    "TestABC",
				Properties: map[string]string{"time": "300"},
			}},
		},
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:      0,
		Type:    "Parent",
		Value:   "dad",
		Context: "dad",
		ChildRelations: []transport.TransportRelation{{
			Target: transport.TransportEntity{
				ID:         -1,
				Type:       "Token",
				Value:      "ishouldnotbefound",
				Context:    "TestABC",
				Properties: map[string]string{"time": "300"},
			}},
		},
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:      0,
		Type:    "Parent",
		Value:   "dad",
		Context: "dad",
		ChildRelations: []transport.TransportRelation{{
			Target: transport.TransportEntity{
				ID:         -1,
				Type:       "Token",
				Value:      "inthedarknessihide",
				Context:    "TestABC",
				Properties: map[string]string{"time": "300"},
			}},
		},
	})
	qry := New().Read("User").Match("Value", "==", "testuser").To(
		New().Read("Token").Match("Value", "==", "findme").Match("Properties.Context", "==", "TestABC"),
	)
	ret := Execute(qry)
	archivist.Info("data return", ret)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestReadMatchWithMultipleRequiredMatch(t *testing.T) {
	gits.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestDEF",
		Context: "TestABC",
	})
	gits.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestDEF",
	})
	qry := New().Read("Test").Match("Value", "==", "TestABC").Match("Context", "==", "TestABC")
	ret := Execute(qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestDeleteEntityByTypeAndID(t *testing.T) {
	initStorage()
	gits.CreateEntityType("Test")
	entity := gits.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	qry := New().Read("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	ret := Execute(qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	qry = New().Delete("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	Execute(qry)
	qry = New().Read("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	ret = Execute(qry)
	if 0 != len(ret.Entities) {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestQueryLinkTo(t *testing.T) {
	initStorage()
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
	qry := New().Read("Test")
	tmp := Execute(qry)

	if 2 != len(tmp.Entities) || 0 != len(tmp.Entities[0].ChildRelations) || 0 != len(tmp.Entities[1].ChildRelations) {
		t.Error(tmp)
	}
	// link the datasets
	qry = New().Link("Test").Match("Value", "==", "TestABC").To(
		New().Find("Test").Match("Value", "==", "TestDEF"),
	)
	Execute(qry)

	// now read out to approve we gotr the linked data
	qry = New().Read("Test").Match("Value", "==", "TestABC").To(
		New().Read("Test"),
	)
	ret := Execute(qry)
	if 1 != len(ret.Entities) || 1 != len(ret.Entities[0].ChildRelations) {
		t.Error("missing results", ret)
	}
	if "TestABC" != ret.Entities[0].Value || "TestDEF" != ret.Entities[0].ChildRelations[0].Target.Value {
		t.Error("incorrect link direction", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestQueryLinkFrom(t *testing.T) {
	initStorage()
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
	qry := New().Read("Test")
	tmp := Execute(qry)
	if 2 != len(tmp.Entities) {
		t.Error("missing results", tmp)
	}

	// link the datasets
	qry = New().Link("Test").Match("Value", "==", "TestABC").From(
		New().Find("Test").Match("Value", "==", "TestDEF"),
	)
	Execute(qry)

	// now read out to approve we gotr the linked data
	qry = New().Read("Test").Match("Value", "==", "TestABC").From(
		New().Read("Test"),
	)
	ret := Execute(qry)
	if 1 != len(ret.Entities) || 1 != len(ret.Entities[0].ParentRelations) {
		t.Error("missing results", ret)
	}
	if "TestABC" != ret.Entities[0].Value || "TestDEF" != ret.Entities[0].ParentRelations[0].Target.Value {
		t.Error("incorrect link direction", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestQueryUnlink(t *testing.T) {
	initStorage()
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
	qry := New().Read("TestA").Match("Value", "==", "Something").To(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	tmp := Execute(qry)
	if 1 != len(tmp.Entities) || 1 != len(tmp.Entities[0].ChildRelations) {
		t.Error("missing results", tmp)
	}
	if "Something" != tmp.Entities[0].Value || "Else" != tmp.Entities[0].ChildRelations[0].Target.Value {
		t.Error("incorrect link direction", tmp)
	}

	// unlink the data
	qry = New().Unlink("TestA").Match("Value", "==", "Something").To(
		New().Find("TestB").Match("Value", "==", "Else"),
	)
	Execute(qry)

	// read linked inserted data
	qry = New().Read("TestA").Match("Value", "==", "Something").To(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	ret := Execute(qry)
	if 0 != len(ret.Entities) {
		t.Error("there should be no result", tmp)
	}

	// make sure the entries have no links on either side
	qry = New().Read("TestA", "TestB")
	ret = Execute(qry)
	if 2 != len(ret.Entities) {
		t.Error("there should be 2 results", tmp)
	}
	if 0 < len(ret.Entities[0].ChildRelations) || 0 < len(ret.Entities[1].ChildRelations) || 0 < len(ret.Entities[0].ParentRelations) || 0 < len(ret.Entities[1].ParentRelations) {
		t.Error("there should be no relations", tmp)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestQueryUnlinkReverse(t *testing.T) {
	initStorage()
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
	qry := New().Read("TestA").Match("Value", "==", "Something").From(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	tmp := Execute(qry)
	if 1 != len(tmp.Entities) || 1 != len(tmp.Entities[0].ParentRelations) {
		t.Error("testdata not existent further processing makes no sense", tmp)
	}

	// unlink the data
	qry = New().Unlink("TestA").Match("Value", "==", "Something").From(
		New().Find("TestB").Match("Value", "==", "Else"),
	)
	Execute(qry)

	// read linked inserted data
	qry = New().Read("TestA").Match("Value", "==", "Something").From(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	ret := Execute(qry)
	if 0 != len(ret.Entities) {
		t.Error("there should not be any result", ret)
	}
	qry = New().Read("TestA", "TestB")
	ret = Execute(qry)
	if 2 != len(ret.Entities) {
		t.Error("there should be 2 entries", ret)
	}
	if 0 != len(ret.Entities[0].ChildRelations) || 0 != len(ret.Entities[0].ParentRelations) || 0 != len(ret.Entities[0].ChildRelations) || 0 != len(ret.Entities[0].ParentRelations) {
		t.Error("there are relations that shouldn exist", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func review_TestbuildTestQueryJson(t *testing.T) {
	initStorage()
	qry := New().Read("IP").To(
		New().Read("Port").To(
			New().Read("Software"),
		),
	)
	printData(qry)
}

func review_TestbuildTestQueryJson2(t *testing.T) {
	initStorage()
	qry := New().Read("IP").To(
		New().Read("Port").To(
			New().Read("Software").To(
				New().Read("Vhost"),
			),
		),
	)
	printData(qry)
}

func review_TestbuildTestQueryJson3(t *testing.T) {
	qry := New().Read("IP").To(
		New().Read("Port").To(
			New().Read("Software").To(
				New().Read("Vhost"),
			),
		).To(
			New().Read("Software"),
		),
	)
	printData(qry)
}

func review_TestbuildTestQueryJsonGetQbQueries(t *testing.T) {
	initStorage()
	//
	archivist.Info("Get all marketplaces implemented by Max Mustermann from person")
	qry := New().Read("Person").Match("Value", "==", "Max Mustermann").To(
		New().Read("Marketplace").Match("Properties.IsAbstract", "==", "false"),
	).To(
		New().Read("Marketplace").To(
			New().Read("Marketplace"),
		),
	)
	//printData(qry)

	//
	archivist.Info("Get all marketplaces shipping to germany")
	qry = New().Read("Marketplace").To(
		New().Reduce("Country").Match("Value", "==", "Germany"),
	)
	//printData(qry)

	//
	archivist.Info("Get all marketplaces ")
	qry = New().Read("Person").Match("Value", "==", "Max Mustermann").To(
		New().Read("Marketplace"),
	).To(
		New().Read("Marketplace").To(
			New().Read("Marketplace"),
		),
	)
	printData(qry)

	archivist.Info("Get Person that implemented marketplace")
	qry = New().Read("Marketplace").From(
		New().Read("Person"),
	)
	//printData(qry)
}

func TestRequiredAndOptionalMixedAlpha(t *testing.T) {
	initStorage()
	// create the testdata
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)

	//  Test forced 2 depth marketplace
	qry := New().Read("Person").To(
		New().Read("Marketplace").To(
			New().Read("Marketplace"),
		).To(
			New().Read("Marketplace"),
		),
	)

	ret := Execute(qry)
	if 1 != len(ret.Entities) || 1 != len(ret.Entities[0].ChildRelations) ||
		4 != len(ret.Entities[0].ChildRelations[0].Target.ChildRelations) {
		t.Error("missing results", ret)
	}

	t.Cleanup(func() {
		Cleanup()
	})
}

func TestRequiredAndOptionalMixedBeta(t *testing.T) {
	initStorage()
	// create the testdata
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)

	qry := New().Read("Person").To(
		New().Read("Marketplace").CanTo(
			New().Read("Marketplace"),
		),
	)
	ret := Execute(qry)
	if 1 != len(ret.Entities) || 2 != len(ret.Entities[0].ChildRelations) {
		t.Error("missing results", ret)
	}

	// check specific child relation exists, since order isnt clear it has to be tested
	// we can only assue that base data exists on our precheck
	testKey := 1
	if ret.Entities[0].ChildRelations[0].Target.Value == "Wortmann" {
		testKey = 0
	}
	if 2 != len(ret.Entities[0].ChildRelations[testKey].Target.ChildRelations) {
		t.Error("missing results", ret)
	}

	t.Cleanup(func() {
		Cleanup()
	})
}

func TestOptionalQueryJoinFirstLevel(t *testing.T) {
	initStorage()
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)
	qry := New().Read("Person").CanTo(
		New().Read("Shipping"),
	)
	ret := Execute(qry)
	if 1 != len(ret.Entities) || 0 < len(ret.Entities[0].ChildRelations) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestRequiredQueryJoinFirstLevelSuccess(t *testing.T) {
	initStorage()
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)
	qry := New().Read("Person").To(
		New().Read("Marketplace"),
	)
	ret := Execute(qry)

	if 1 != len(ret.Entities) || 2 != len(ret.Entities[0].ChildRelations) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestRequiredQueryJoinFirstLevelFail(t *testing.T) {
	initStorage()
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)
	qry := New().Read("Person").To(
		New().Read("Shipping"),
	)
	ret := Execute(qry)

	if 0 != len(ret.Entities) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestRequiredQueryJoinInDepthFail(t *testing.T) {
	initStorage()
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test required first level join  - - - - - - - - -")
	qry := New().Read("Person").To(
		New().Read("Marketplace").To(
			New().Read("Person"),
		),
	)
	ret := Execute(qry)

	if 0 != len(ret.Entities) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestRequiredQueryJoinInDepthSuccess(t *testing.T) {
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test required first level join  - - - - - - - - -")
	qry := New().Read("Person").To(
		New().Read("Marketplace").To(
			New().Read("Country"),
		),
	)
	ret := Execute(qry)

	if 1 != len(ret.Entities) || 1 != len(ret.Entities[0].ChildRelations) || 2 != len(ret.Entities[0].ChildRelations[0].Target.ChildRelations) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestRequiredQueryJoinWithRelationValuesSuccess(t *testing.T) {
	testdata := mapQbStructureMap()
	gits.MapTransportData(testdata)
	archivist.Info(" - - - - - - - - - Test required first level join  - - - - - - - - -")
	qry := New().Read("Person").To(
		New().Read("Marketplace").Match("Value", "==", "Gabor").To(
			New().Read("Country").Match("Value", "==", "Germany"),
		),
	)
	ret := Execute(qry)

	if 1 != len(ret.Entities) || 1 != len(ret.Entities[0].ChildRelations) || 1 != len(ret.Entities[0].ChildRelations[0].Target.ChildRelations) {
		t.Error("wrong result format", ret)
	}

	archivist.Info("Result data", ret)

	t.Cleanup(func() {
		Cleanup()
	})
}

func printData(data any) {
	t, _ := json.MarshalIndent(data, "", "\t")
	archivist.Info("Query Data Struct", string(t))
}

func mapQbStructureMap() transport.TransportEntity {
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
	//printData(testdata)
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

func TestOrderByNumericValueAsc(t *testing.T) {
	initStorage()
	createTestDataForOrderByValueNumeric()
	qry := New().Read("Something").Order("Value", ORDER_DIRECTION_ASC, ORDER_MODE_NUM)
	ret := Execute(qry)

	if 3 != len(ret.Entities) || "1" != ret.Entities[0].Value || "2" != ret.Entities[1].Value || "333" != ret.Entities[2].Value {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestOrderByNumericValueDesc(t *testing.T) {
	initStorage()
	createTestDataForOrderByValueNumeric()
	qry := New().Read("Something").Order("Value", ORDER_DIRECTION_DESC, ORDER_MODE_NUM)
	ret := Execute(qry)

	if 3 != len(ret.Entities) || "333" != ret.Entities[0].Value || "2" != ret.Entities[1].Value || "1" != ret.Entities[2].Value {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestOrderByAlphabeticalValueAsc(t *testing.T) {
	initStorage()
	createTestDataOrderAlphabetical()
	qry := New().Read("Something").Order("Value", ORDER_DIRECTION_ASC, ORDER_MODE_ALPHA)
	ret := Execute(qry)

	if 3 != len(ret.Entities) || "auch" != ret.Entities[0].Value || "Das" != ret.Entities[1].Value || "Zebra" != ret.Entities[2].Value {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestOrderByAlphabeticalValueDesc(t *testing.T) {
	initStorage()
	createTestDataOrderAlphabetical()
	qry := New().Read("Something").Order("Value", ORDER_DIRECTION_DESC, ORDER_MODE_ALPHA)
	ret := Execute(qry)

	if 3 != len(ret.Entities) || "Zebra" != ret.Entities[0].Value || "Das" != ret.Entities[1].Value || "auch" != ret.Entities[2].Value {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func Cleanup() {
	gits.EntityStorage = make(map[int]map[int]types.StorageEntity)
	gits.EntityIDMax = make(map[int]int)
	gits.EntityTypes = make(map[int]string)
	gits.EntityRTypes = make(map[string]int)
	gits.EntityTypeIDMax = 0
	gits.RelationStorage = make(map[int]map[int]map[int]map[int]types.StorageRelation)
	gits.RelationRStorage = make(map[int]map[int]map[int]map[int]bool)
}
