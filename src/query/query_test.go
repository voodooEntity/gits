package query

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/voodooEntity/gits/src/storage"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
)

var testStorage *storage.Storage

func initStorage() {
	// archivist.Init("info", "stdout", "blafu")

	newStorage := storage.NewStorage()
	testStorage = newStorage
}

func createTestDataLinked() {

	// create test data to query
	typeIDalpha, _ := testStorage.CreateEntityType("Alpha")
	typeIDbeta, _ := testStorage.CreateEntityType("Beta")
	typeIDdelta, _ := testStorage.CreateEntityType("Delta")
	typeIDgamma, _ := testStorage.CreateEntityType("Gamma")

	entityAlphaID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:    typeIDalpha,
		Value:   "alpha",
		Context: "uno",
	})

	entityBetaID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:    typeIDbeta,
		Value:   "beta",
		Context: "duo",
	})

	entityDeltaID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:    typeIDdelta,
		Value:   "delta",
		Context: "tres",
	})

	entityGammaID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:    typeIDgamma,
		Value:   "gamma",
		Context: "quattro",
	})

	//printData(gits.GetDefault().EntityStorage)

	testStorage.CreateRelation(typeIDalpha, entityAlphaID, typeIDbeta, entityBetaID, types.StorageRelation{
		SourceType: typeIDalpha,
		SourceID:   entityAlphaID,
		TargetType: typeIDbeta,
		TargetID:   entityBetaID,
	})

	testStorage.CreateRelation(typeIDbeta, entityBetaID, typeIDdelta, entityDeltaID, types.StorageRelation{
		SourceType: typeIDbeta,
		SourceID:   entityBetaID,
		TargetType: typeIDdelta,
		TargetID:   entityDeltaID,
	})

	testStorage.CreateRelation(typeIDalpha, entityAlphaID, typeIDgamma, entityGammaID, types.StorageRelation{
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
	if 1 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func createTestDataLinearTypeNumericValue() {
	typeIDalpha, _ := testStorage.CreateEntityType("Alpha")
	for i := 1; i <= 100; i++ {
		testStorage.CreateEntity(types.StorageEntity{
			Type:    typeIDalpha,
			Value:   strconv.Itoa(i),
			Context: "uno",
		})
	}
}

func createTestDataLinearTypeNumericPropertyTestValue() {
	typeIDalpha, _ := testStorage.CreateEntityType("Alpha")
	for i := 1; i <= 100; i++ {
		props := make(map[string]string)
		props["Test"] = strconv.Itoa(i)
		testStorage.CreateEntity(types.StorageEntity{
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
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
	result := Execute(testStorage, qry)
	if 1 != len(result.Entities) {
		t.Error(result)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestUpdateEntityValue(t *testing.T) {
	initStorage()
	testStorage.CreateEntityType("Test")
	testStorage.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	qry := New().Read("Test")
	ret := Execute(testStorage, qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	qry = New().Update("Test").Match("Value", "==", "TestABC").Set("Value", "TestDEF").Set("Context", "asdasdasd")
	Execute(testStorage, qry)
	qry = New().Read("Test")
	ret = Execute(testStorage, qry)
	if 1 != len(ret.Entities) && "TestDEF" == ret.Entities[0].Value && "asdasdasd" == ret.Entities[0].Context {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestReadJoinMatchWithMultipleRequiredMatch(t *testing.T) {
	testStorage.MapTransportData(transport.TransportEntity{
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
	testStorage.MapTransportData(transport.TransportEntity{
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
	testStorage.MapTransportData(transport.TransportEntity{
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
	ret := Execute(testStorage, qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestFindValidTokenRequest(t *testing.T) {
	testStorage.MapTransportData(transport.TransportEntity{
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
	testStorage.MapTransportData(transport.TransportEntity{
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
	testStorage.MapTransportData(transport.TransportEntity{
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
		New().Read("Token").Match("Value", "==", "findme").Match("Context", "==", "TestABC"),
	)
	ret := Execute(testStorage, qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestReadMatchWithMultipleRequiredMatch(t *testing.T) {
	testStorage.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestDEF",
		Context: "TestABC",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestDEF",
	})
	qry := New().Read("Test").Match("Value", "==", "TestABC").Match("Context", "==", "TestABC")
	ret := Execute(testStorage, qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestDeleteEntityByTypeAndID(t *testing.T) {
	initStorage()
	testStorage.CreateEntityType("Test")
	entity := testStorage.MapTransportData(transport.TransportEntity{
		ID:      -1,
		Type:    "Test",
		Value:   "TestABC",
		Context: "TestABC",
	})
	qry := New().Read("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	ret := Execute(testStorage, qry)
	if 1 != len(ret.Entities) {
		t.Error(ret)
	}
	qry = New().Delete("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	Execute(testStorage, qry)
	qry = New().Read("Test").Match("ID", "==", strconv.Itoa(entity.ID))
	ret = Execute(testStorage, qry)
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
	testStorage.CreateEntityType("Test")
	testStorage.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestABC",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestDEF",
	})

	// print the testdata before linking
	qry := New().Read("Test")
	tmp := Execute(testStorage, qry)

	if 2 != len(tmp.Entities) || 0 != len(tmp.Entities[0].ChildRelations) || 0 != len(tmp.Entities[1].ChildRelations) {
		t.Error(tmp)
	}
	// link the datasets
	qry = New().Link("Test").Match("Value", "==", "TestABC").To(
		New().Find("Test").Match("Value", "==", "TestDEF"),
	)
	Execute(testStorage, qry)

	// now read out to approve we gotr the linked data
	qry = New().Read("Test").Match("Value", "==", "TestABC").To(
		New().Read("Test"),
	)
	ret := Execute(testStorage, qry)
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
	testStorage.CreateEntityType("Test")
	testStorage.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestABC",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		Type:  "Test",
		ID:    -1,
		Value: "TestDEF",
	})

	// print the testdata before linking
	qry := New().Read("Test")
	tmp := Execute(testStorage, qry)
	if 2 != len(tmp.Entities) {
		t.Error("missing results", tmp)
	}

	// link the datasets
	qry = New().Link("Test").Match("Value", "==", "TestABC").From(
		New().Find("Test").Match("Value", "==", "TestDEF"),
	)
	Execute(testStorage, qry)

	// now read out to approve we gotr the linked data
	qry = New().Read("Test").Match("Value", "==", "TestABC").From(
		New().Read("Test"),
	)
	ret := Execute(testStorage, qry)
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
	testStorage.CreateEntityType("TestA")
	testStorage.CreateEntityType("TestB")
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
	testStorage.MapTransportData(testdata)

	// read linked inserted data
	qry := New().Read("TestA").Match("Value", "==", "Something").To(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	tmp := Execute(testStorage, qry)
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
	Execute(testStorage, qry)

	// read linked inserted data
	qry = New().Read("TestA").Match("Value", "==", "Something").To(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	ret := Execute(testStorage, qry)
	if 0 != len(ret.Entities) {
		t.Error("there should be no result", tmp)
	}

	// make sure the entries have no links on either side
	qry = New().Read("TestA", "TestB")
	ret = Execute(testStorage, qry)
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
	testStorage.CreateEntityType("TestA")
	testStorage.CreateEntityType("TestB")
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
	testStorage.MapTransportData(testdata)

	// read linked inserted data
	qry := New().Read("TestA").Match("Value", "==", "Something").From(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	tmp := Execute(testStorage, qry)
	if 1 != len(tmp.Entities) || 1 != len(tmp.Entities[0].ParentRelations) {
		t.Error("testdata not existent further processing makes no sense", tmp)
	}

	// unlink the data
	qry = New().Unlink("TestA").Match("Value", "==", "Something").From(
		New().Find("TestB").Match("Value", "==", "Else"),
	)
	Execute(testStorage, qry)

	// read linked inserted data
	qry = New().Read("TestA").Match("Value", "==", "Something").From(
		New().Read("TestB").Match("Value", "==", "Else"),
	)
	ret := Execute(testStorage, qry)
	if 0 != len(ret.Entities) {
		t.Error("there should not be any result", ret)
	}
	qry = New().Read("TestA", "TestB")
	ret = Execute(testStorage, qry)
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

func TestRequiredAndOptionalMixedAlpha(t *testing.T) {
	initStorage()
	// create the testdata
	testdata := mapQbStructureMap()
	testStorage.MapTransportData(testdata)

	//  Test forced 2 depth marketplace
	qry := New().Read("Person").To(
		New().Read("Marketplace").To(
			New().Read("Marketplace"),
		).To(
			New().Read("Marketplace"),
		),
	)

	ret := Execute(testStorage, qry)
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
	testStorage.MapTransportData(testdata)

	qry := New().Read("Person").To(
		New().Read("Marketplace").CanTo(
			New().Read("Marketplace"),
		),
	)
	ret := Execute(testStorage, qry)
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
	testStorage.MapTransportData(testdata)
	qry := New().Read("Person").CanTo(
		New().Read("Shipping"),
	)
	ret := Execute(testStorage, qry)
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
	testStorage.MapTransportData(testdata)
	qry := New().Read("Person").To(
		New().Read("Marketplace"),
	)
	ret := Execute(testStorage, qry)

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
	testStorage.MapTransportData(testdata)
	qry := New().Read("Person").To(
		New().Read("Shipping"),
	)
	ret := Execute(testStorage, qry)

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
	testStorage.MapTransportData(testdata)
	qry := New().Read("Person").To(
		New().Read("Marketplace").To(
			New().Read("Person"),
		),
	)
	ret := Execute(testStorage, qry)

	if 0 != len(ret.Entities) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestLimitApplies(t *testing.T) {
	for i := 0; i < 100; i++ {
		testStorage.MapTransportData(transport.TransportEntity{
			ID:      -1,
			Value:   "Something" + strconv.Itoa(i),
			Context: "test",
			Type:    "Alpha",
		})
	}
	qry := New().Read("Alpha").Limit(4)
	ret := Execute(testStorage, qry)
	if 4 != len(ret.Entities) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestLimitButLessDatasets(t *testing.T) {
	for i := 0; i < 5; i++ {
		testStorage.MapTransportData(transport.TransportEntity{
			ID:      -1,
			Value:   "Something" + strconv.Itoa(i),
			Context: "test",
			Type:    "Alpha",
		})
	}
	qry := New().Read("Alpha").Limit(10)
	ret := Execute(testStorage, qry)
	if 5 != len(ret.Entities) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestRequiredQueryJoinInDepthSuccess(t *testing.T) {
	testdata := mapQbStructureMap()
	testStorage.MapTransportData(testdata)
	qry := New().Read("Person").To(
		New().Read("Marketplace").To(
			New().Read("Country"),
		),
	)
	ret := Execute(testStorage, qry)

	if 1 != len(ret.Entities) || 1 != len(ret.Entities[0].ChildRelations) || 2 != len(ret.Entities[0].ChildRelations[0].Target.ChildRelations) {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func printData(data any) {
	t, _ := json.MarshalIndent(data, "", "\t")
	fmt.Println("Query Data Struct", string(t))
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
	testStorage.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "2",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "333",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "1",
	})
}

func createTestDataOrderAlphabetical() {
	testStorage.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "Das",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "Zebra",
	})
	testStorage.MapTransportData(transport.TransportEntity{
		ID:    -1,
		Type:  "Something",
		Value: "auch",
	})
}

func TestOrderByNumericValueAsc(t *testing.T) {
	initStorage()
	createTestDataForOrderByValueNumeric()
	qry := New().Read("Something").Order("Value", ORDER_DIRECTION_ASC, ORDER_MODE_NUM)
	ret := Execute(testStorage, qry)

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
	ret := Execute(testStorage, qry)

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
	ret := Execute(testStorage, qry)

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
	ret := Execute(testStorage, qry)

	if 3 != len(ret.Entities) || "Zebra" != ret.Entities[0].Value || "Das" != ret.Entities[1].Value || "auch" != ret.Entities[2].Value {
		t.Error("wrong result format", ret)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestUpdateWithRequiredSubqueryFilter(t *testing.T) {
	initStorage()
	defer Cleanup()

	// Create entity types
	userTypeID, _ := testStorage.CreateEntityType("User")
	orderTypeID, _ := testStorage.CreateEntityType("Order")

	// Create UserA (should be updated)
	userA_ID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:       userTypeID,
		Value:      "UserA",
		Context:    "TestUpdateBug",
		Properties: map[string]string{"Status": "Active"},
	})
	order1_ID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:       orderTypeID,
		Value:      "Order101",
		Context:    "ForUserA",
		Properties: map[string]string{"Amount": "150"},
	})
	testStorage.CreateRelationUnsafe(userTypeID, userA_ID, orderTypeID, order1_ID, types.StorageRelation{
		SourceType: userTypeID, SourceID: userA_ID, TargetType: orderTypeID, TargetID: order1_ID,
	})

	// Create UserB (should NOT be updated)
	userB_ID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:       userTypeID,
		Value:      "UserB",
		Context:    "TestUpdateBug",
		Properties: map[string]string{"Status": "Active"},
	})
	order2_ID, _ := testStorage.CreateEntity(types.StorageEntity{
		Type:       orderTypeID,
		Value:      "Order102",
		Context:    "ForUserB",
		Properties: map[string]string{"Amount": "50"}, // Amount not > 100
	})
	testStorage.CreateRelationUnsafe(userTypeID, userB_ID, orderTypeID, order2_ID, types.StorageRelation{
		SourceType: userTypeID, SourceID: userB_ID, TargetType: orderTypeID, TargetID: order2_ID,
	})

	// Create UserC (should also NOT be updated, matches root but no orders)
	testStorage.CreateEntity(types.StorageEntity{
		Type:       userTypeID,
		Value:      "UserC",
		Context:    "TestUpdateBugNoOrders",
		Properties: map[string]string{"Status": "Active"},
	})

	// Update query: Update Users with Status "Active" who have an Order with Amount > 100
	updateQry := New().Update("User").
		Match("Properties.Status", "==", "Active").
		Set("Properties.Status", "Processed").
		To(New().Reduce("Order").Match("Properties.Amount", ">", "100"))

	updateResult := Execute(testStorage, updateQry)

	// Assertions
	// Check UserA
	userAResult := Execute(testStorage, New().Read("User").Match("Value", "==", "UserA"))
	if len(userAResult.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter: Expected 1 UserA, got %d", len(userAResult.Entities))
	} else if userAResult.Entities[0].Properties["Status"] != "Processed" {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter: Expected UserA Status to be 'Processed', got '%s'", userAResult.Entities[0].Properties["Status"])
	}

	// Check UserB
	userBResult := Execute(testStorage, New().Read("User").Match("Value", "==", "UserB"))
	if len(userBResult.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter: Expected 1 UserB, got %d", len(userBResult.Entities))
	} else if userBResult.Entities[0].Properties["Status"] != "Active" {
		// This is the crucial assertion for the bug
		t.Errorf("TestUpdateWithRequiredSubqueryFilter: Expected UserB Status to remain 'Active', got '%s'. BUG PRESENT.", userBResult.Entities[0].Properties["Status"])
	}

	// Check UserC
	userCResult := Execute(testStorage, New().Read("User").Match("Value", "==", "UserC"))
	if len(userCResult.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter: Expected 1 UserC, got %d", len(userCResult.Entities))
	} else if userCResult.Entities[0].Properties["Status"] != "Active" {
		// This also tests the bug: UserC matches root but fails subquery
		t.Errorf("TestUpdateWithRequiredSubqueryFilter: Expected UserC Status to remain 'Active', got '%s'. BUG PRESENT.", userCResult.Entities[0].Properties["Status"])
	}

	// Check the amount returned by the Update operation
	// If the bug exists, ret.Amount might be 1 (for UserA), but more users were updated.
	// If the fix is applied, ret.Amount should be 1.
	// The current code in query.go calculates ret.Amount correctly based on filtered entities.
	// The bug is that BatchUpdateAddressList uses the unfiltered list.
	// So, updateResult.Amount should reflect the count of *correctly* updatable entities.
	if updateResult.Amount != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter: Update operation reported affecting %d entities, expected 1", updateResult.Amount)
	}
}

func TestComplexNestedUnlinkConditioning(t *testing.T) {
	initStorage()

	checkVal := "iShouldNotGetChanged"
	checkField := "TestProperty"

	// first we map our testdata
	testStorage.MapTransportData(transport.TransportEntity{
		ID:    storage.MAP_FORCE_CREATE,
		Type:  "IP",
		Value: "127.0.0.1",
		Properties: map[string]string{
			checkField: checkVal,
		},
		ChildRelations: []transport.TransportRelation{
			transport.TransportRelation{
				Target: transport.TransportEntity{
					ID:    storage.MAP_FORCE_CREATE,
					Type:  "Port",
					Value: "80",
					ChildRelations: []transport.TransportRelation{
						transport.TransportRelation{
							Target: transport.TransportEntity{
								ID:    storage.MAP_FORCE_CREATE,
								Type:  "Banner",
								Value: "Apache 2.4",
							},
						},
						transport.TransportRelation{
							Target: transport.TransportEntity{
								ID:    storage.MAP_FORCE_CREATE,
								Type:  "Vhost",
								Value: "127.0.0.1",
								ChildRelations: []transport.TransportRelation{
									transport.TransportRelation{
										Target: transport.TransportEntity{
											ID:    storage.MAP_FORCE_CREATE,
											Type:  "Directory",
											Value: "/",
										},
									},
								},
							},
						},
					},
				},
			},
			transport.TransportRelation{
				Target: transport.TransportEntity{
					ID:    storage.MAP_FORCE_CREATE,
					Type:  "Port",
					Value: "22",
					ChildRelations: []transport.TransportRelation{
						transport.TransportRelation{
							Target: transport.TransportEntity{
								ID:    storage.MAP_FORCE_CREATE,
								Type:  "Software",
								Value: "sshd 2.3.4",
								ChildRelations: []transport.TransportRelation{
									transport.TransportRelation{
										Target: transport.TransportEntity{
											ID:    storage.MAP_FORCE_CREATE,
											Type:  "CVE",
											Value: "ASDASD123123",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	// first we check if we correctly can read it
	pre := New().Read("IP").Match("Properties."+checkField, "==", checkVal).To(
		New().Reduce("Port").Match("Value", "==", "80").To(
			New().Reduce("Banner").Match("Value", "==", "Apache 2.4"),
		).To(
			New().Reduce("Vhost").Match("Value", "==", "127.0.0.1").To(
				New().Reduce("Directory").Match("Value", "==", "/"),
			),
		),
	).To(
		New().Reduce("Port").Match("Value", "==", "22").To(
			New().Reduce("Software").Match("Value", "==", "sshd 2.3.4").To(
				New().Reduce("CVE").Match("Value", "==", "ASDASD123123"),
			),
		),
	)
	preRet := Execute(testStorage, pre)
	if len(preRet.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter Pre Failed : Expected 1 IP, got %d", len(preRet.Entities))
	}

	// now we create a "complex" query to check if it
	// still gets changed : first check , if on multi nested parallel condition 1 fails
	testA := New().Update("IP").Set("Properties."+checkField, "nope").To(
		New().Reduce("Port").Match("Value", "==", "80").To(
			New().Reduce("Banner").Match("Value", "==", "Apache 2.4"),
		).To(
			New().Reduce("Vhost").Match("Value", "==", "127.0.0.2").To(
				New().Reduce("Directory").Match("Value", "==", "/"),
			),
		),
	).To(
		New().Reduce("Port").Match("Value", "==", "22").To(
			New().Reduce("Software").Match("Value", "==", "sshd 2.3.4").To(
				New().Reduce("CVE").Match("Value", "==", "ASDASD123123"),
			),
		),
	)
	Execute(testStorage, testA)
	ret := Execute(testStorage, New().Read("IP").Match("Properties."+checkField, "==", checkVal))
	if len(ret.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter Check 1: Expected 1 IP, got %d", len(ret.Entities))
	}

	// second test, if we got on a multi level 1 canTo and 1 To on same level and see if the optional creates a fail success
	testB := New().Update("IP").Set("Properties."+checkField, "nope").To(
		New().Reduce("Port").Match("Value", "==", "81").To( // changed port
			New().Reduce("Banner").Match("Value", "==", "Apache 2.3"),
		).CanTo(
			New().Reduce("Vhost").Match("Value", "==", "127.0.0.1").To(
				New().Reduce("Directory").Match("Value", "==", "/"),
			),
		),
	).To(
		New().Reduce("Port").Match("Value", "==", "22").To(
			New().Reduce("Software").Match("Value", "==", "sshd 2.3.4").To(
				New().Reduce("CVE").Match("Value", "==", "ASDASD123123"),
			),
		),
	)
	Execute(testStorage, testB)
	ret = Execute(testStorage, New().Read("IP").Match("Properties."+checkField, "==", checkVal))
	if len(ret.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter Check 2: Expected 1 IP, got %d", len(ret.Entities))
	}

	// third test, if we got on fist join level 1 to and 1 canto and see if it behaves correctly
	testC := New().Update("IP").Set("Properties."+checkField, "nope").To(
		New().Reduce("Port").Match("Value", "==", "80").To(
			New().Reduce("Banner").Match("Value", "==", "Apache 2.4"),
		).To(
			New().Reduce("Vhost").Match("Value", "==", "127.0.0.1").To(
				New().Reduce("Directory").Match("Value", "==", "/phpmyadmin/"), // changed directory
			),
		),
	).CanTo(
		New().Reduce("Port").Match("Value", "==", "22").To(
			New().Reduce("Software").Match("Value", "==", "sshd 2.3.4").To(
				New().Reduce("CVE").Match("Value", "==", "ASDASD123123"),
			),
		),
	)
	Execute(testStorage, testC)
	ret = Execute(testStorage, New().Read("IP").Match("Properties."+checkField, "==", checkVal))
	if len(ret.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter Check 3: Expected 1 IP, got %d", len(ret.Entities))
	}

	// third test, if we got on fist join level 1 to and 1 canto and see if it behaves correctly
	testD := New().Read("IP").Match("Properties."+checkField, "==", checkVal).To(
		New().Reduce("Port").Match("Value", "==", "80").To(
			New().Reduce("Banner").Match("Value", "==", "Apache 2.4"),
		).To(
			New().Reduce("Vhost").Match("Value", "==", "127.0.0.1").To(
				New().Reduce("Directory").Match("Value", "==", "/phpmyadmin"), // changed directory
			),
		),
	).CanTo(
		New().Reduce("Port").Match("Value", "==", "22").To(
			New().Reduce("Software").Match("Value", "==", "sshd 2.3.4").To(
				New().Reduce("CVE").Match("Value", "==", "ASDASD123123"),
			),
		),
	)
	Execute(testStorage, testD)
	ret = Execute(testStorage, New().Read("IP").Match("Properties."+checkField, "==", checkVal))
	if len(ret.Entities) != 1 {
		t.Errorf("TestUpdateWithRequiredSubqueryFilter Check 4: Expected 1 IP, got %d", len(ret.Entities))
	}

	t.Cleanup(func() {
		Cleanup()
	})
}

func Cleanup() {
	testStorage.EntityStorage = make(map[int]map[int]types.StorageEntity)
	testStorage.EntityIDMax = make(map[int]int)
	testStorage.EntityTypes = make(map[int]string)
	testStorage.EntityRTypes = make(map[string]int)
	testStorage.EntityTypeIDMax = 0
	testStorage.RelationStorage = make(map[int]map[int]map[int]map[int]types.StorageRelation)
	testStorage.RelationRStorage = make(map[int]map[int]map[int]map[int]bool)
}
