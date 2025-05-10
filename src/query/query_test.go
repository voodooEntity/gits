package query

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/voodooEntity/gits/src/query/cond" // Import for new cond package
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

// --- Test Suite for Enhanced Link/Unlink ---

func createTestDataForLinkUnlink() {
	// Entity Types
	typeIDUser, _ := testStorage.CreateEntityType("User")
	typeIDGroup, _ := testStorage.CreateEntityType("Group")
	typeIDProject, _ := testStorage.CreateEntityType("Project")

	// Entities
	_, _ = testStorage.CreateEntity(types.StorageEntity{Type: typeIDUser, Value: "User1"})       // ID 1
	_, _ = testStorage.CreateEntity(types.StorageEntity{Type: typeIDUser, Value: "User2"})       // ID 2
	_, _ = testStorage.CreateEntity(types.StorageEntity{Type: typeIDGroup, Value: "GroupA"})     // ID 1
	_, _ = testStorage.CreateEntity(types.StorageEntity{Type: typeIDGroup, Value: "GroupB"})     // ID 2
	_, _ = testStorage.CreateEntity(types.StorageEntity{Type: typeIDProject, Value: "ProjectX"}) // ID 1
	_, _ = testStorage.CreateEntity(types.StorageEntity{Type: typeIDProject, Value: "ProjectY"}) // ID 2

	// Initial relations for testing Unlink
	// User1 -> GroupA (Context: "member", Properties: {"role": "admin", "status": "active"})
	testStorage.CreateRelationUnsafe(typeIDUser, 1, typeIDGroup, 1, types.StorageRelation{
		SourceType: typeIDUser, SourceID: 1, TargetType: typeIDGroup, TargetID: 1,
		Context: "member", Properties: map[string]string{"role": "admin", "status": "active"},
	})
	// User1 -> GroupB (Context: "viewer", Properties: {"role": "guest"})
	testStorage.CreateRelationUnsafe(typeIDUser, 1, typeIDGroup, 2, types.StorageRelation{
		SourceType: typeIDUser, SourceID: 1, TargetType: typeIDGroup, TargetID: 2,
		Context: "viewer", Properties: map[string]string{"role": "guest"},
	})
	// User2 -> ProjectX (Context: "contributor", Properties: {"permission": "write", "active": "true"})
	testStorage.CreateRelationUnsafe(typeIDUser, 2, typeIDProject, 1, types.StorageRelation{
		SourceType: typeIDUser, SourceID: 2, TargetType: typeIDProject, TargetID: 1,
		Context: "contributor", Properties: map[string]string{"permission": "write", "active": "true"},
	})
}

func TestLinkWithContextAndProperties_Success(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink()

	// Link User2 to GroupA with context "manager" and property "level":"senior"
	linkQry := New().Link("User").Match("Value", "==", "User2").
		To(New().Find("Group").Match("Value", "==", "GroupA")).
		WithRelationContext("manager").
		WithRelationProperty("level", "senior")
	result := Execute(testStorage, linkQry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1, got %d", result.Amount)
	}

	// Verify the relation
	readQry := New().Read("User").Match("Value", "==", "User2").
		To(New().Read("Group").Match("Value", "==", "GroupA"))
	readResult := Execute(testStorage, readQry)

	if readResult.Amount != 1 || len(readResult.Entities[0].ChildRelations) != 1 {
		t.Fatalf("Expected User2 to be linked to GroupA, got: %+v", readResult)
	}
	relation := readResult.Entities[0].ChildRelations[0]
	if relation.Context != "manager" {
		t.Errorf("Expected relation context 'manager', got '%s'", relation.Context)
	}
	if val, ok := relation.Properties["level"]; !ok || val != "senior" {
		t.Errorf("Expected relation property 'level':'senior', got '%+v'", relation.Properties)
	}

	t.Cleanup(func() { Cleanup() })
}

func TestLinkWithMultipleProperties_Success(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink()

	// Link User2 to GroupB with multiple properties
	linkQry := New().Link("User").Match("Value", "==", "User2").
		To(New().Find("Group").Match("Value", "==", "GroupB")).
		WithRelationContext("team_lead").
		WithRelationProperties(map[string]string{"department": "engineering", "start_date": "2023-01-01"})
	result := Execute(testStorage, linkQry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1, got %d", result.Amount)
	}

	readQry := New().Read("User").Match("Value", "==", "User2").
		To(New().Read("Group").Match("Value", "==", "GroupB"))
	readResult := Execute(testStorage, readQry)
	if readResult.Amount != 1 || len(readResult.Entities[0].ChildRelations) != 1 {
		t.Fatalf("Expected User2 to be linked to GroupB, got: %+v", readResult)
	}
	relation := readResult.Entities[0].ChildRelations[0]
	if relation.Context != "team_lead" {
		t.Errorf("Expected relation context 'team_lead', got '%s'", relation.Context)
	}
	if relation.Properties["department"] != "engineering" || relation.Properties["start_date"] != "2023-01-01" {
		t.Errorf("Expected relation properties not found, got '%+v'", relation.Properties)
	}
	t.Cleanup(func() { Cleanup() })
}

func TestUnlinkMatchingContext_Success(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink() // User1 -> GroupA (member), User1 -> GroupB (viewer)

	// Unlink User1 from GroupA where context is "member"
	unlinkQry := New().Unlink("User").Match("Value", "==", "User1").
		To(New().Find("Group").Match("Value", "==", "GroupA")).
		MatchingRelationContext("member")
	result := Execute(testStorage, unlinkQry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1 for unlinked relation, got %d", result.Amount)
	}

	// Verify User1 is no longer linked to GroupA
	readQryGA := New().Read("User").Match("Value", "==", "User1").
		To(New().Read("Group").Match("Value", "==", "GroupA"))
	readResultGA := Execute(testStorage, readQryGA)
	if readResultGA.Amount != 0 { // Should be 0 as the link is removed
		t.Errorf("User1 should no longer be linked to GroupA, but got %d results", readResultGA.Amount)
	}

	// Verify User1 is still linked to GroupB
	readQryGB := New().Read("User").Match("Value", "==", "User1").
		To(New().Read("Group").Match("Value", "==", "GroupB"))
	readResultGB := Execute(testStorage, readQryGB)
	if readResultGB.Amount != 1 {
		t.Errorf("User1 should still be linked to GroupB, got %d results", readResultGB.Amount)
	}
	t.Cleanup(func() { Cleanup() })
}

func TestUnlinkMatchingContext_FailNotMatching(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink() // User1 -> GroupA (member)

	// Attempt to Unlink User1 from GroupA with non-matching context "non_member"
	unlinkQry := New().Unlink("User").Match("Value", "==", "User1").
		To(New().Find("Group").Match("Value", "==", "GroupA")).
		MatchingRelationContext("non_member")
	result := Execute(testStorage, unlinkQry)

	if result.Amount != 0 { // No relation should be unlinked
		t.Errorf("Expected Amount to be 0 as context does not match, got %d", result.Amount)
	}

	// Verify User1 is still linked to GroupA
	readQry := New().Read("User").Match("Value", "==", "User1").
		To(New().Read("Group").Match("Value", "==", "GroupA"))
	readResult := Execute(testStorage, readQry)
	if readResult.Amount != 1 {
		t.Errorf("User1 should still be linked to GroupA, got %d results", readResult.Amount)
	}
	t.Cleanup(func() { Cleanup() })
}

func TestUnlinkMatchingProperty_Success(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink() // User1 -> GroupA (role:admin), User1 -> GroupB (role:guest)

	// Unlink User1 from GroupA where property role == "admin"
	unlinkQry := New().Unlink("User").Match("Value", "==", "User1").
		To(New().Find("Group").Match("Value", "==", "GroupA")).
		MatchingRelationProperty("role", "==", "admin")
	result := Execute(testStorage, unlinkQry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1 for unlinked relation, got %d", result.Amount)
	}

	// Verify User1 is no longer linked to GroupA
	readQryGA := New().Read("User").Match("Value", "==", "User1").
		To(New().Read("Group").Match("Value", "==", "GroupA"))
	readResultGA := Execute(testStorage, readQryGA)
	if readResultGA.Amount != 0 {
		t.Errorf("User1 should no longer be linked to GroupA, got %d results", readResultGA.Amount)
	}
	t.Cleanup(func() { Cleanup() })
}

func TestUnlinkMatchingProperty_FailNotMatching(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink() // User1 -> GroupA (role:admin)

	// Attempt to Unlink User1 from GroupA with non-matching property role == "user"
	unlinkQry := New().Unlink("User").Match("Value", "==", "User1").
		To(New().Find("Group").Match("Value", "==", "GroupA")).
		MatchingRelationProperty("role", "==", "user")
	result := Execute(testStorage, unlinkQry)

	if result.Amount != 0 {
		t.Errorf("Expected Amount to be 0 as property does not match, got %d", result.Amount)
	}
	t.Cleanup(func() { Cleanup() })
}

func TestUnlinkMatchingMultipleProperties_Success(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink() // User1 -> GroupA (role:admin, status:active)

	unlinkQry := New().Unlink("User").Match("Value", "==", "User1").
		To(New().Find("Group").Match("Value", "==", "GroupA")).
		MatchingRelationProperty("role", "==", "admin").
		MatchingRelationProperty("status", "==", "active")
	result := Execute(testStorage, unlinkQry)
	if result.Amount != 1 {
		t.Errorf("Expected 1 relation to be unlinked, got %d", result.Amount)
	}
	t.Cleanup(func() { Cleanup() })
}

func TestUnlinkMatchingMultipleProperties_FailOneNotMatching(t *testing.T) {
	initStorage()
	createTestDataForLinkUnlink() // User1 -> GroupA (role:admin, status:active)

	unlinkQry := New().Unlink("User").Match("Value", "==", "User1").
		To(New().Find("Group").Match("Value", "==", "GroupA")).
		MatchingRelationProperty("role", "==", "admin").
		MatchingRelationProperty("status", "==", "inactive") // This one fails
	result := Execute(testStorage, unlinkQry)
	if result.Amount != 0 {
		t.Errorf("Expected 0 relations to be unlinked, got %d", result.Amount)
	}
	t.Cleanup(func() { Cleanup() })
}

// --- End of Test Suite for Enhanced Link/Unlink ---

// --- Test Suite for Update Query with Required Joins ---

// Helper function to create test data for update-join scenarios
func createTestDataForUpdateJoins() {
	// Entity Types
	// Main entities to be updated: "UpdateTarget"
	// Related entities: "RequiredChild", "RequiredParent", "OptionalChild"
	typeIDUpdateTarget, _ := testStorage.CreateEntityType("UpdateTarget")
	typeIDRequiredChild, _ := testStorage.CreateEntityType("RequiredChild")
	typeIDRequiredParent, _ := testStorage.CreateEntityType("RequiredParent")
	typeIDOptionalChild, _ := testStorage.CreateEntityType("OptionalChild")

	// --- Scenario 1: Target1 (should pass required To join) ---
	// UpdateTarget1 -> RequiredChild1
	ut1ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDUpdateTarget, Value: "UT1", Properties: map[string]string{"Status": "Initial"}})
	rc1ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDRequiredChild, Value: "RC1"})
	testStorage.CreateRelation(typeIDUpdateTarget, ut1ID, typeIDRequiredChild, rc1ID, types.StorageRelation{})

	// --- Scenario 2: Target2 (should FAIL required To join) ---
	// UpdateTarget2 (no child)
	ut2ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDUpdateTarget, Value: "UT2", Properties: map[string]string{"Status": "Initial"}})
	_ = ut2ID // use ut2ID if needed later, for now it's just created

	// --- Scenario 3: Target3 (should pass required From join) ---
	// RequiredParent1 -> UpdateTarget3
	ut3ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDUpdateTarget, Value: "UT3", Properties: map[string]string{"Status": "Initial"}})
	rp1ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDRequiredParent, Value: "RP1"})
	testStorage.CreateRelation(typeIDRequiredParent, rp1ID, typeIDUpdateTarget, ut3ID, types.StorageRelation{})

	// --- Scenario 4: Target4 (should pass multiple required joins and optional) ---
	// RequiredParent2 -> UpdateTarget4 -> RequiredChild2
	// UpdateTarget4 -> OptionalChild1 (optional)
	ut4ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDUpdateTarget, Value: "UT4", Properties: map[string]string{"Status": "Initial"}})
	rp2ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDRequiredParent, Value: "RP2"})
	rc2ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDRequiredChild, Value: "RC2"})
	oc1ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDOptionalChild, Value: "OC1"})
	testStorage.CreateRelation(typeIDRequiredParent, rp2ID, typeIDUpdateTarget, ut4ID, types.StorageRelation{}) // RP2 -> UT4
	testStorage.CreateRelation(typeIDUpdateTarget, ut4ID, typeIDRequiredChild, rc2ID, types.StorageRelation{})  // UT4 -> RC2
	testStorage.CreateRelation(typeIDUpdateTarget, ut4ID, typeIDOptionalChild, oc1ID, types.StorageRelation{})  // UT4 -> OC1

	// --- Scenario 5: Target5 (all matching initially, but one will fail a multi-join) ---
	// UpdateTarget5 -> RequiredChild3
	// RequiredParent3 -> UpdateTarget5 (This one is MISSING for UT6)
	ut5ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDUpdateTarget, Value: "UT5", Properties: map[string]string{"Status": "Initial"}})
	rc3ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDRequiredChild, Value: "RC3"})
	rp3ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDRequiredParent, Value: "RP3"})
	testStorage.CreateRelation(typeIDUpdateTarget, ut5ID, typeIDRequiredChild, rc3ID, types.StorageRelation{})  // UT5 -> RC3
	testStorage.CreateRelation(typeIDRequiredParent, rp3ID, typeIDUpdateTarget, ut5ID, types.StorageRelation{}) // RP3 -> UT5

	// UpdateTarget6 -> RequiredChild4 (This one is OK)
	// RequiredParent4 -> UpdateTarget6 (This relation will be MISSING to cause failure for the batch)
	ut6ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDUpdateTarget, Value: "UT6", Properties: map[string]string{"Status": "Initial"}})
	rc4ID, _ := testStorage.CreateEntity(types.StorageEntity{Type: typeIDRequiredChild, Value: "RC4"})
	testStorage.CreateRelation(typeIDUpdateTarget, ut6ID, typeIDRequiredChild, rc4ID, types.StorageRelation{}) // UT6 -> RC4
	// Missing: RP -> UT6
}

func TestUpdateWithRequiredToJoin_Success(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// UT1 has RequiredChild1. Update should proceed.
	qry := New().Update("UpdateTarget").Match("Value", "==", "UT1").
		To(New().Read("RequiredChild").Match("Value", "==", "RC1")).
		Set("Properties.Status", "Updated")
	result := Execute(testStorage, qry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1, got %d", result.Amount)
	}

	// Verify UT1 was updated
	readQry := New().Read("UpdateTarget").Match("Value", "==", "UT1")
	readResult := Execute(testStorage, readQry)
	if readResult.Amount != 1 || readResult.Entities[0].Properties["Status"] != "Updated" {
		t.Errorf("UT1 was not updated as expected. Status: %s", readResult.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

func TestUpdateWithRequiredToJoin_Fail(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// UT1 has child, UT2 does not. Batch update should fail.
	// Query matches UT1 and UT2 initially.
	qry := New().Update("UpdateTarget").Match("Context", "==", ""). // Matches all UT initially if context is empty or not set
									To(New().Read("RequiredChild")). // UT2 will fail this
									Set("Properties.Status", "UpdatedBatchFail")
	result := Execute(testStorage, qry)

	if result.Amount != 0 {
		t.Errorf("Expected Amount to be 0 because UT2 misses required join, got %d", result.Amount)
	}

	// Verify NEITHER UT1 nor UT2 was updated
	readQryUT1 := New().Read("UpdateTarget").Match("Value", "==", "UT1")
	readResultUT1 := Execute(testStorage, readQryUT1)
	if readResultUT1.Entities[0].Properties["Status"] != "Initial" {
		t.Errorf("UT1 should not have been updated. Status: %s", readResultUT1.Entities[0].Properties["Status"])
	}

	readQryUT2 := New().Read("UpdateTarget").Match("Value", "==", "UT2")
	readResultUT2 := Execute(testStorage, readQryUT2)
	if readResultUT2.Entities[0].Properties["Status"] != "Initial" {
		t.Errorf("UT2 should not have been updated. Status: %s", readResultUT2.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

func TestUpdateWithRequiredFromJoin_Success(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// UT3 has RequiredParent1. Update should proceed.
	qry := New().Update("UpdateTarget").Match("Value", "==", "UT3").
		From(New().Read("RequiredParent").Match("Value", "==", "RP1")).
		Set("Properties.Status", "Updated")
	result := Execute(testStorage, qry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1, got %d", result.Amount)
	}

	readQry := New().Read("UpdateTarget").Match("Value", "==", "UT3")
	readResult := Execute(testStorage, readQry)
	if readResult.Amount != 1 || readResult.Entities[0].Properties["Status"] != "Updated" {
		t.Errorf("UT3 was not updated as expected. Status: %s", readResult.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

func TestUpdateWithRequiredFromJoin_Fail(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// UT2 has no parent. If query targets UT2 and requires a parent, it should fail.
	// Let's make a query that targets UT2 and UT3. UT2 will fail.
	qry := New().Update("UpdateTarget").Match("Value", "in", "UT2,UT3"). // Matches UT2 and UT3
										From(New().Read("RequiredParent")). // UT2 will fail this
										Set("Properties.Status", "UpdatedBatchFail")
	result := Execute(testStorage, qry)

	if result.Amount != 0 {
		t.Errorf("Expected Amount to be 0 because UT2 misses required From join, got %d", result.Amount)
	}

	// Verify UT3 was NOT updated
	readQryUT3 := New().Read("UpdateTarget").Match("Value", "==", "UT3")
	readResultUT3 := Execute(testStorage, readQryUT3)
	if readResultUT3.Entities[0].Properties["Status"] != "Initial" {
		t.Errorf("UT3 should not have been updated. Status: %s", readResultUT3.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

func TestUpdateWithMultipleRequiredJoins_Success(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// UT4 has RP2 -> UT4 -> RC2. Update should proceed.
	qry := New().Update("UpdateTarget").Match("Value", "==", "UT4").
		From(New().Read("RequiredParent").Match("Value", "==", "RP2")).
		To(New().Read("RequiredChild").Match("Value", "==", "RC2")).
		Set("Properties.Status", "UpdatedMulti")
	result := Execute(testStorage, qry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1, got %d", result.Amount)
	}

	readQry := New().Read("UpdateTarget").Match("Value", "==", "UT4")
	readResult := Execute(testStorage, readQry)
	if readResult.Amount != 1 || readResult.Entities[0].Properties["Status"] != "UpdatedMulti" {
		t.Errorf("UT4 was not updated as expected. Status: %s", readResult.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

func TestUpdateWithMultipleRequiredJoins_Fail(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// Query targets UT5 and UT6.
	// UT5: RP3 -> UT5 -> RC3 (OK)
	// UT6: (No Parent) -> UT6 -> RC4 (FAILS From join)
	// Batch should fail.
	qry := New().Update("UpdateTarget").Match("Value", "in", "UT5,UT6").
		From(New().Read("RequiredParent")). // UT6 fails this
		To(New().Read("RequiredChild")).    // Both UT5 and UT6 have a RequiredChild
		Set("Properties.Status", "UpdatedMultiFail")
	result := Execute(testStorage, qry)

	if result.Amount != 0 {
		t.Errorf("Expected Amount to be 0 because UT6 misses a required From join, got %d", result.Amount)
	}

	// Verify UT5 was NOT updated
	readQryUT5 := New().Read("UpdateTarget").Match("Value", "==", "UT5")
	readResultUT5 := Execute(testStorage, readQryUT5)
	if readResultUT5.Entities[0].Properties["Status"] != "Initial" {
		t.Errorf("UT5 should not have been updated. Status: %s", readResultUT5.Entities[0].Properties["Status"])
	}
	// Verify UT6 was NOT updated
	readQryUT6 := New().Read("UpdateTarget").Match("Value", "==", "UT6")
	readResultUT6 := Execute(testStorage, readQryUT6)
	if readResultUT6.Entities[0].Properties["Status"] != "Initial" {
		t.Errorf("UT6 should not have been updated. Status: %s", readResultUT6.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

func TestUpdateWithOptionalCanToJoin_Proceeds(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// UT1 has RequiredChild1. UT2 does not.
	// Update UT1 and UT2. Optional CanTo join for OptionalChild.
	// UT1 does not have OptionalChild. UT4 has OptionalChild1.
	// The update should proceed for both UT1 and UT2 as CanTo is optional.
	qry := New().Update("UpdateTarget").Match("Value", "in", "UT1,UT2").
		CanTo(New().Read("OptionalChild")). // This join is optional
		Set("Properties.Status", "UpdatedWithOptional")
	result := Execute(testStorage, qry)

	if result.Amount != 2 { // Both UT1 and UT2 should be "updated" (attempted)
		t.Errorf("Expected Amount to be 2, got %d", result.Amount)
	}

	// Verify UT1 was updated
	readQryUT1 := New().Read("UpdateTarget").Match("Value", "==", "UT1")
	readResultUT1 := Execute(testStorage, readQryUT1)
	if readResultUT1.Amount != 1 || readResultUT1.Entities[0].Properties["Status"] != "UpdatedWithOptional" {
		t.Errorf("UT1 was not updated as expected. Status: %s", readResultUT1.Entities[0].Properties["Status"])
	}
	// Verify UT2 was updated
	readQryUT2 := New().Read("UpdateTarget").Match("Value", "==", "UT2")
	readResultUT2 := Execute(testStorage, readQryUT2)
	if readResultUT2.Amount != 1 || readResultUT2.Entities[0].Properties["Status"] != "UpdatedWithOptional" {
		t.Errorf("UT2 was not updated as expected. Status: %s", readResultUT2.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

func TestUpdateWithNoJoins_Success(t *testing.T) {
	initStorage()
	createTestDataForUpdateJoins()

	// Update UT1, no joins specified. Should proceed.
	qry := New().Update("UpdateTarget").Match("Value", "==", "UT1").
		Set("Properties.Status", "UpdatedNoJoins")
	result := Execute(testStorage, qry)

	if result.Amount != 1 {
		t.Errorf("Expected Amount to be 1, got %d", result.Amount)
	}

	readQry := New().Read("UpdateTarget").Match("Value", "==", "UT1")
	readResult := Execute(testStorage, readQry)
	if readResult.Amount != 1 || readResult.Entities[0].Properties["Status"] != "UpdatedNoJoins" {
		t.Errorf("UT1 was not updated as expected. Status: %s", readResult.Entities[0].Properties["Status"])
	}

	t.Cleanup(func() { Cleanup() })
}

// --- End of Test Suite for Update Query with Required Joins ---

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
	fmt.Println("data return", ret)
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
	fmt.Println("data return", ret)
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
	fmt.Println("Get all marketplaces implemented by Max Mustermann from person")
	qry := New().Read("Person").Match("Value", "==", "Max Mustermann").To(
		New().Read("Marketplace").Match("Properties.IsAbstract", "==", "false"),
	).To(
		New().Read("Marketplace").To(
			New().Read("Marketplace"),
		),
	)
	//printData(qry)

	//
	fmt.Println("Get all marketplaces shipping to germany")
	qry = New().Read("Marketplace").To(
		New().Reduce("Country").Match("Value", "==", "Germany"),
	)
	//printData(qry)

	//
	fmt.Println("Get all marketplaces ")
	qry = New().Read("Person").Match("Value", "==", "Max Mustermann").To(
		New().Read("Marketplace"),
	).To(
		New().Read("Marketplace").To(
			New().Read("Marketplace"),
		),
	)
	printData(qry)

	fmt.Println("Get Person that implemented marketplace")
	qry = New().Read("Marketplace").From(
		New().Read("Person"),
	)
	//printData(qry)
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
	fmt.Println(" - - - - - - - - - Test required first level join  - - - - - - - - -")
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
	fmt.Println(" - - - - - - - - - Test required first level join  - - - - - - - - -")
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

func Cleanup() {
	testStorage.EntityStorage = make(map[int]map[int]types.StorageEntity)
	testStorage.EntityIDMax = make(map[int]int)
	testStorage.EntityTypes = make(map[int]string)
	testStorage.EntityRTypes = make(map[string]int)
	testStorage.EntityTypeIDMax = 0
	testStorage.RelationStorage = make(map[int]map[int]map[int]map[int]types.StorageRelation)
	testStorage.RelationRStorage = make(map[int]map[int]map[int]map[int]bool)
}

// Test data for complex filter tests
func createTestDataForComplexFilters() {
	typeIDcomplex, _ := testStorage.CreateEntityType("Complex")
	testStorage.CreateEntity(types.StorageEntity{
		Type:       typeIDcomplex,
		Value:      "ValA",
		Context:    "Ctx1",
		Properties: map[string]string{"Status": "Active", "Count": "10"},
	})
	testStorage.CreateEntity(types.StorageEntity{
		Type:       typeIDcomplex,
		Value:      "ValB",
		Context:    "Ctx1",
		Properties: map[string]string{"Status": "Active", "Count": "20"},
	})
	testStorage.CreateEntity(types.StorageEntity{
		Type:       typeIDcomplex,
		Value:      "ValC",
		Context:    "Ctx2",
		Properties: map[string]string{"Status": "Inactive", "Count": "30"},
	})
	testStorage.CreateEntity(types.StorageEntity{
		Type:       typeIDcomplex,
		Value:      "ValD",
		Context:    "Ctx2",
		Properties: map[string]string{"Status": "Active", "Count": "5"},
	})
	testStorage.CreateEntity(types.StorageEntity{ // For NOT tests
		Type:       typeIDcomplex,
		Value:      "ValE_NotThis",
		Context:    "Ctx_Not",
		Properties: map[string]string{"Status": "Pending", "Count": "100"},
	})
}

func TestComplexFilter_SimpleAnd(t *testing.T) {
	initStorage()
	createTestDataForComplexFilters()
	qry := New().Read("Complex").Filter(
		cond.And(
			cond.Match("Context", "==", "Ctx1"),
			cond.Match("Properties.Status", "==", "Active"),
		),
	)
	result := Execute(testStorage, qry)
	if 2 != result.Amount {
		t.Errorf("Expected 2 entities, got %d. Results: %+v", result.Amount, result.Entities)
	}
	// Further checks for specific values if needed
	foundValA := false
	foundValB := false
	for _, e := range result.Entities {
		if e.Value == "ValA" {
			foundValA = true
		}
		if e.Value == "ValB" {
			foundValB = true
		}
	}
	if !foundValA || !foundValB {
		t.Errorf("Expected ValA and ValB, got: %+v", result.Entities)
	}

	t.Cleanup(func() {
		Cleanup()
	})
}

func TestComplexFilter_SimpleOr(t *testing.T) {
	initStorage()
	createTestDataForComplexFilters()
	qry := New().Read("Complex").Filter(
		cond.Or(
			cond.Match("Value", "==", "ValC"),
			cond.Match("Properties.Count", "==", "5"),
		),
	)
	result := Execute(testStorage, qry)
	if 2 != result.Amount {
		t.Errorf("Expected 2 entities, got %d. Results: %+v", result.Amount, result.Entities)
	}
	foundValC := false
	foundValD := false
	for _, e := range result.Entities {
		if e.Value == "ValC" {
			foundValC = true
		}
		if e.Value == "ValD" {
			foundValD = true
		}
	}
	if !foundValC || !foundValD {
		t.Errorf("Expected ValC and ValD, got: %+v", result.Entities)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestComplexFilter_NestedAndOr(t *testing.T) {
	initStorage()
	createTestDataForComplexFilters()
	// (Context == "Ctx1" AND Properties.Status == "Active") OR Value == "ValD"
	// Should return ValA, ValB, ValD
	qry := New().Read("Complex").Filter(
		cond.Or(
			cond.And(
				cond.Match("Context", "==", "Ctx1"),
				cond.Match("Properties.Status", "==", "Active"),
			),
			cond.Match("Value", "==", "ValD"),
		),
	)
	result := Execute(testStorage, qry)
	if 3 != result.Amount {
		t.Errorf("Expected 3 entities, got %d. Results: %+v", result.Amount, result.Entities)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestComplexFilter_NegatedMatch(t *testing.T) {
	initStorage()
	createTestDataForComplexFilters()
	// Context == "Ctx2" AND Properties.Status != "Inactive"
	// Should return ValD
	qry := New().Read("Complex").Filter(
		cond.And(
			cond.Match("Context", "==", "Ctx2"),
			cond.Match("Properties.Status", "==", "Inactive").SetNegated(true),
		),
	)
	result := Execute(testStorage, qry)
	if 1 != result.Amount {
		t.Errorf("Expected 1 entity, got %d. Results: %+v", result.Amount, result.Entities)
	}
	if result.Entities[0].Value != "ValD" {
		t.Errorf("Expected ValD, got: %+v", result.Entities[0])
	}
	t.Cleanup(func() {
		Cleanup()
	})
}

func TestComplexFilter_NegatedGroup(t *testing.T) {
	initStorage()
	createTestDataForComplexFilters()
	// NOT (Context == "Ctx1" OR Properties.Status == "Inactive")
	// This means Context != "Ctx1" AND Properties.Status != "Inactive"
	// (Context == Ctx2 AND Status == Active) -> ValD
	// (Context == Ctx_Not AND Status == Pending) -> ValE_NotThis
	// Should return ValD and ValE_NotThis
	qry := New().Read("Complex").Filter(
		cond.Or(
			cond.Match("Context", "==", "Ctx1"),
			cond.Match("Properties.Status", "==", "Inactive"),
		).SetNegated(true),
	)
	result := Execute(testStorage, qry)
	if 2 != result.Amount {
		t.Errorf("Expected 2 entities, got %d. Results: %+v", result.Amount, result.Entities)
	}
	foundValD := false
	foundValE := false
	for _, e := range result.Entities {
		if e.Value == "ValD" {
			foundValD = true
		}
		if e.Value == "ValE_NotThis" {
			foundValE = true
		}
	}
	if !foundValD || !foundValE {
		t.Errorf("Expected ValD and ValE_NotThis, got: %+v", result.Entities)
	}

	t.Cleanup(func() {
		Cleanup()
	})
}

func TestComplexFilter_CoexistenceWithLegacyMatch_FilterTakesPrecedence(t *testing.T) {
	initStorage()
	createTestDataForComplexFilters()
	// Legacy Match would find ValA. Filter should find ValC.
	qry := New().Read("Complex").
		Match("Value", "==", "ValA"). // This should be ignored
		Filter(cond.Match("Value", "==", "ValC"))

	result := Execute(testStorage, qry)
	if 1 != result.Amount {
		t.Errorf("Expected 1 entity from Filter, got %d. Results: %+v", result.Amount, result.Entities)
	}
	if result.Entities[0].Value != "ValC" {
		t.Errorf("Expected ValC from Filter, got: %s", result.Entities[0].Value)
	}
	t.Cleanup(func() {
		Cleanup()
	})
}
