package gits_test

import (
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/types"
	"testing"
)

func TestCreateEntityType(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityTypes := gits.EntityTypes
	originalEntityRTypes := gits.EntityRTypes
	originalEntityTypeIDMax := gits.EntityTypeIDMax

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityTypes = originalEntityTypes
		gits.EntityRTypes = originalEntityRTypes
		gits.EntityTypeIDMax = originalEntityTypeIDMax

		// Reset any other necessary state or clean up any previously created entities.
	}()

	// Call the CreateEntityType function with a unique name for testing
	entityTypeName := "TestEntityType"
	id, err := gits.CreateEntityType(entityTypeName)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("CreateEntityType returned an error: %v", err)
	}

	// Check if the ID returned is greater than zero
	if id <= 0 {
		t.Fatalf("CreateEntityType returned an invalid ID: %d", id)
	}

	// Check if the entity type name is correctly mapped to the ID in EntityTypes
	entityType, ok := gits.EntityTypes[id]
	if !ok || entityType != entityTypeName {
		t.Fatalf("EntityTypes mapping is incorrect: expected %s, got %s", entityTypeName, entityType)
	}

	// Check if the reverse mapping in EntityRTypes is correct
	reverseID, ok := gits.EntityRTypes[entityTypeName]
	if !ok || reverseID != id {
		t.Fatalf("EntityRTypes mapping is incorrect: expected %d, got %d", id, reverseID)
	}

	// You can add more checks here based on your specific requirements and validation rules.
}

func TestCreateEntityTypeUnsafe(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityTypes := gits.EntityTypes
	originalEntityRTypes := gits.EntityRTypes
	originalEntityTypeIDMax := gits.EntityTypeIDMax

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityTypes = originalEntityTypes
		gits.EntityRTypes = originalEntityRTypes
		gits.EntityTypeIDMax = originalEntityTypeIDMax

		// Reset any other necessary state or clean up any previously created entities.
	}()

	// Call the CreateEntityTypeUnsafe function with a unique name for testing
	entityTypeName := "TestEntityTypeUnsafe"
	id, err := gits.CreateEntityTypeUnsafe(entityTypeName)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("CreateEntityTypeUnsafe returned an error: %v", err)
	}

	// Check if the ID returned is greater than zero
	if id <= 0 {
		t.Fatalf("CreateEntityTypeUnsafe returned an invalid ID: %d", id)
	}

	// Check if the entity type name is correctly mapped to the ID in EntityTypes
	entityType, ok := gits.EntityTypes[id]
	if !ok || entityType != entityTypeName {
		t.Fatalf("EntityTypes mapping is incorrect: expected %s, got %s", entityTypeName, entityType)
	}

	// Check if the reverse mapping in EntityRTypes is correct
	reverseID, ok := gits.EntityRTypes[entityTypeName]
	if !ok || reverseID != id {
		t.Fatalf("EntityRTypes mapping is incorrect: expected %d, got %d", id, reverseID)
	}

	// You can add more checks here based on your specific requirements and validation rules.
}

func TestCreateEntity(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityStorageMutex := gits.EntityStorageMutex
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypeMutex := gits.EntityTypeMutex

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type: 1, // Replace with a valid type ID for your test
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityStorageMutex = originalEntityStorageMutex
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypeMutex = originalEntityTypeMutex

		// Reset any other necessary state or clean up any created entities.
	}()

	// Call the CreateEntity function with the mock entity
	id, err := gits.CreateEntity(mockEntity)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("CreateEntity returned an error: %v", err)
	}

	// Check if the ID returned is greater than zero
	if id <= 0 {
		t.Fatalf("CreateEntity returned an invalid ID: %d", id)
	}

	// Check if the entity is correctly stored in EntityStorage
	storedEntity, ok := gits.EntityStorage[mockEntity.Type][id]
	if !ok {
		t.Fatalf("EntityStorage does not contain the created entity")
	}

	// Compare the stored entity with the mock entity for equality
	// You may need to implement an equality check for your entity type.
	if !isEqualEntity(mockEntity, storedEntity) {
		t.Fatalf("Stored entity differs from the mock entity")
	}

	// You can add more checks here based on your specific requirements and validation rules.
}

// Implement a custom function to compare two entities for equality.
func isEqualEntity(entity1, entity2 types.StorageEntity) bool {
	// Implement your equality comparison logic here based on your entity structure.
	// Return true if the entities are equal, false otherwise.
	// Example:
	return entity1.Type == entity2.Type && entity1.ID == entity2.ID
}

func TestCreateEntityUnsafe(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type: 1, // Replace with a valid type ID for your test
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Call the CreateEntityUnsafe function with the mock entity
	id, err := gits.CreateEntityUnsafe(mockEntity)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("CreateEntityUnsafe returned an error: %v", err)
	}

	// Check if the ID returned is greater than zero
	if id <= 0 {
		t.Fatalf("CreateEntityUnsafe returned an invalid ID: %d", id)
	}

	// Check if the entity is correctly stored in EntityStorage
	storedEntity, ok := gits.EntityStorage[entityType][id]
	if !ok {
		t.Fatalf("EntityStorage does not contain the created entity")
	}

	// Compare the stored entity with the mock entity for equality
	// You may need to implement an equality check for your entity type.
	if !isEqualEntity(mockEntity, storedEntity) {
		t.Fatalf("Stored entity differs from the mock entity")
	}

	// You can add more checks here based on your specific requirements and validation rules.
}

func TestCreateEntityUniqueValue(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type:    1,             // Replace with a valid type ID for your test
		Value:   "UniqueValue", // Replace with a unique value for testing
		Context: "TestContext",
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Call the CreateEntityUniqueValue function with the mock entity
	id, created, err := gits.CreateEntityUniqueValue(mockEntity)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("CreateEntityUniqueValue returned an error: %v", err)
	}

	// Check if the entity was created
	if created {
		// Check if the ID returned is greater than zero
		if id <= 0 {
			t.Fatalf("CreateEntityUniqueValue returned an invalid ID: %d", id)
		}

		// Check if the entity is correctly stored in EntityStorage
		storedEntity, ok := gits.EntityStorage[entityType][id]
		if !ok {
			t.Fatalf("EntityStorage does not contain the created entity")
		}

		// Compare the stored entity with the mock entity for equality
		// You may need to implement an equality check for your entity type.
		if !isEqualEntity(mockEntity, storedEntity) {
			t.Fatalf("Stored entity differs from the mock entity")
		}
	} else {
		// Entity with the same unique value already exists
		// You can add additional checks or assertions here if needed.
	}
}

func TestCreateEntityUniqueValueUnsafe(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type:    1,             // Replace with a valid type ID for your test
		Value:   "UniqueValue", // Replace with a unique value for testing
		Context: "TestContext",
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Call the CreateEntityUniqueValueUnsafe function with the mock entity
	id, created, err := gits.CreateEntityUniqueValueUnsafe(mockEntity)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("CreateEntityUniqueValueUnsafe returned an error: %v", err)
	}

	if created {
		// Check if the ID returned is greater than zero
		if id <= 0 {
			t.Fatalf("CreateEntityUniqueValueUnsafe returned an invalid ID: %d", id)
		}

		// Check if the entity is correctly stored in EntityStorage
		storedEntity, ok := gits.EntityStorage[entityType][id]
		if !ok {
			t.Fatalf("EntityStorage does not contain the created entity")
		}

		// Compare the stored entity with the mock entity for equality
		// You may need to implement an equality check for your entity type.
		if !isEqualEntity(mockEntity, storedEntity) {
			t.Fatalf("Stored entity differs from the mock entity")
		}
	} else {
		// Entity with the same unique value already exists
		// You can add additional checks or assertions here if needed.
	}
}

func TestGetEntityByPath(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type:    1, // Replace with a valid type ID for your test
		ID:      1, // Replace with a valid entity ID for your test
		Context: "TestContext",
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Add the mock entity to EntityStorage
	gits.EntityStorage[entityType] = make(map[int]types.StorageEntity)
	gits.EntityStorage[entityType][mockEntity.ID] = mockEntity

	// Call the GetEntityByPath function with the mock entity's path
	entity, err := gits.GetEntityByPath(entityType, mockEntity.ID, mockEntity.Context)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("GetEntityByPath returned an error: %v", err)
	}

	// Check if the returned entity matches the mock entity
	if !isEqualEntity(mockEntity, entity) {
		t.Fatalf("Returned entity differs from the mock entity")
	}

	// Test with an invalid context (should return an error)
	_, err = gits.GetEntityByPath(entityType, mockEntity.ID, "InvalidContext")
	if err == nil {
		t.Fatalf("GetEntityByPath did not return an error for an invalid context")
	}
}

func TestGetEntityByPathUnsafe(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type:    1, // Replace with a valid type ID for your test
		ID:      1, // Replace with a valid entity ID for your test
		Context: "TestContext",
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Add the mock entity to EntityStorage
	gits.EntityStorage[entityType] = make(map[int]types.StorageEntity)
	gits.EntityStorage[entityType][mockEntity.ID] = mockEntity

	// Call the GetEntityByPathUnsafe function with the mock entity's path
	entity, err := gits.GetEntityByPathUnsafe(entityType, mockEntity.ID, mockEntity.Context)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("GetEntityByPathUnsafe returned an error: %v", err)
	}

	// Check if the returned entity matches the mock entity
	if !isEqualEntity(mockEntity, entity) {
		t.Fatalf("Returned entity differs from the mock entity")
	}

	// Test with an invalid context (should return an error)
	_, err = gits.GetEntityByPathUnsafe(entityType, mockEntity.ID, "InvalidContext")
	if err == nil {
		t.Fatalf("GetEntityByPathUnsafe did not return an error for an invalid context")
	}
}

func TestGetEntitiesByType(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type:    1, // Replace with a valid type ID for your test
		ID:      1, // Replace with a valid entity ID for your test
		Context: "TestContext",
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Add the mock entity to EntityStorage
	gits.EntityStorage[entityType] = make(map[int]types.StorageEntity)
	gits.EntityStorage[entityType][mockEntity.ID] = mockEntity

	// Call the GetEntitiesByType function with the mock entity's type
	entities, err := gits.GetEntitiesByType(gits.EntityTypes[entityType], mockEntity.Context)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("GetEntitiesByType returned an error: %v", err)
	}

	// Check if the returned entity map contains the mock entity
	if !containsEntity(entities, mockEntity) {
		t.Fatalf("Returned entity map does not contain the mock entity")
	}

	// Test with an invalid context (should still return the entity)
	entities, err = gits.GetEntitiesByType(gits.EntityTypes[entityType], "InvalidContext")
	if err != nil {
		t.Fatalf("GetEntitiesByType returned an error with an invalid context: %v", err)
	}

	// Check if the returned entity map still contains the mock entity
	if !containsEntity(entities, mockEntity) {
		t.Fatalf("Returned entity map does not contain the mock entity with an invalid context")
	}
}

// Implement a custom function to check if an entity map contains a specific entity.
func containsEntity(entityMap map[int]types.StorageEntity, entity types.StorageEntity) bool {
	for _, e := range entityMap {
		if e.ID == entity.ID && e.Context == entity.Context {
			return true
		}
	}
	return false
}

func TestGetEntitiesByTypeUnsafe(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type:    1, // Replace with a valid type ID for your test
		ID:      1, // Replace with a valid entity ID for your test
		Context: "TestContext",
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Add the mock entity to EntityStorage
	gits.EntityStorage[entityType] = make(map[int]types.StorageEntity)
	gits.EntityStorage[entityType][mockEntity.ID] = mockEntity

	// Call the GetEntitiesByTypeUnsafe function with the mock entity's type
	entities, err := gits.GetEntitiesByTypeUnsafe(gits.EntityTypes[entityType], mockEntity.Context)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("GetEntitiesByTypeUnsafe returned an error: %v", err)
	}

	// Check if the returned entity map contains the mock entity
	if !containsEntity(entities, mockEntity) {
		t.Fatalf("Returned entity map does not contain the mock entity")
	}

	// Test with an invalid context (should still return the entity)
	entities, err = gits.GetEntitiesByTypeUnsafe(gits.EntityTypes[entityType], "InvalidContext")
	if err != nil {
		t.Fatalf("GetEntitiesByTypeUnsafe returned an error with an invalid context: %v", err)
	}

	// Check if the returned entity map still contains the mock entity
	if !containsEntity(entities, mockEntity) {
		t.Fatalf("Returned entity map does not contain the mock entity with an invalid context")
	}
}

func TestGetEntitiesByValue(t *testing.T) {
	// Save the current state of relevant variables for cleanup
	originalEntityStorage := gits.EntityStorage
	originalEntityIDMax := gits.EntityIDMax
	originalEntityTypes := gits.EntityTypes

	// Define a mock entity for testing
	mockEntity := types.StorageEntity{
		Type:    1, // Replace with a valid type ID for your test
		ID:      1, // Replace with a valid entity ID for your test
		Value:   "TestValue",
		Context: "TestContext",
		// You can set other necessary fields here for testing.
	}

	// Defer a cleanup function to reset the state after the test
	defer func() {
		gits.EntityStorage = originalEntityStorage
		gits.EntityIDMax = originalEntityIDMax
		gits.EntityTypes = originalEntityTypes

		// Reset any other necessary state or clean up any created entities.
	}()

	// Ensure that the entity type exists in EntityTypes before testing
	entityType := mockEntity.Type
	if _, ok := gits.EntityTypes[entityType]; !ok {
		t.Fatalf("Entity type does not exist: %d", entityType)
	}

	// Add the mock entity to EntityStorage
	gits.EntityStorage[entityType] = make(map[int]types.StorageEntity)
	gits.EntityStorage[entityType][mockEntity.ID] = mockEntity

	// Call the GetEntitiesByValue function with the mock entity's value
	entities, err := gits.GetEntitiesByValue(mockEntity.Value, "match", mockEntity.Context)

	// Check if the function returned an error
	if err != nil {
		t.Fatalf("GetEntitiesByValue returned an error: %v", err)
	}

	// Check if the returned entity map contains the mock entity
	if !containsEntity(entities, mockEntity) {
		t.Fatalf("Returned entity map does not contain the mock entity")
	}

	// Test with an invalid value (should return an empty map)
	entities, err = gits.GetEntitiesByValue("InvalidValue", "match", mockEntity.Context)
	if err != nil {
		t.Fatalf("GetEntitiesByValue returned an error with an invalid value: %v", err)
	}

	// Check if the returned entity map is empty
	if len(entities) > 0 {
		t.Fatalf("Returned entity map is not empty with an invalid value")
	}

	// Test with an invalid context (should still return the entity)
	entities, err = gits.GetEntitiesByValue(mockEntity.Value, "match", "InvalidContext")
	if err != nil {
		t.Fatalf("GetEntitiesByValue returned an error with an invalid context: %v", err)
	}

	// Check if the returned entity map still contains the mock entity
	if !containsEntity(entities, mockEntity) {
		t.Fatalf("Returned entity map does not contain the mock entity with an invalid context")
	}
}
