package gits

import (
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
	"reflect"
	"testing"
)

func TestCreateEntityType(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateEntityType(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEntityType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateEntityType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateEntityTypeUnsafe(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateEntityTypeUnsafe(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEntityTypeUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateEntityTypeUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchDeleteAddressList(t *testing.T) {
	type args struct {
		addressList [][2]int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BatchDeleteAddressList(tt.args.addressList)
		})
	}
}

func TestBatchUpdateAddressList(t *testing.T) {
	type args struct {
		addressList [][2]int
		values      map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BatchUpdateAddressList(tt.args.addressList, tt.args.values)
		})
	}
}

func TestCreateEntity(t *testing.T) {
	type args struct {
		entity types.StorageEntity
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateEntity(tt.args.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEntity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateEntity() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateEntityUniqueValue(t *testing.T) {
	type args struct {
		entity types.StorageEntity
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := CreateEntityUniqueValue(tt.args.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEntityUniqueValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateEntityUniqueValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CreateEntityUniqueValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCreateEntityUniqueValueUnsafe(t *testing.T) {
	type args struct {
		entity types.StorageEntity
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := CreateEntityUniqueValueUnsafe(tt.args.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEntityUniqueValueUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateEntityUniqueValueUnsafe() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CreateEntityUniqueValueUnsafe() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCreateEntityUnsafe(t *testing.T) {
	type args struct {
		entity types.StorageEntity
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateEntityUnsafe(tt.args.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEntityUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateEntityUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateRelation(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
		relation   types.StorageRelation
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateRelation(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID, tt.args.relation)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateRelation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateRelationUnsafe(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
		relation   types.StorageRelation
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateRelationUnsafe(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID, tt.args.relation)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRelationUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateRelationUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteChildRelations(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteChildRelations(tt.args.Type, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteChildRelations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteChildRelationsUnsafe(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteChildRelationsUnsafe(tt.args.Type, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteChildRelationsUnsafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteEntity(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteEntity(tt.args.Type, tt.args.id)
		})
	}
}

func TestDeleteEntityUnsafe(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteEntityUnsafe(tt.args.Type, tt.args.id)
		})
	}
}

func TestDeleteParentRelations(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteParentRelations(tt.args.Type, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteParentRelations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteParentRelationsUnsafe(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteParentRelationsUnsafe(tt.args.Type, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteParentRelationsUnsafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteRelation(t *testing.T) {
	type args struct {
		sourceType int
		sourceID   int
		targetType int
		targetID   int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteRelation(tt.args.sourceType, tt.args.sourceID, tt.args.targetType, tt.args.targetID)
		})
	}
}

func TestDeleteRelationList(t *testing.T) {
	type args struct {
		relationList map[int]types.StorageRelation
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteRelationList(tt.args.relationList)
		})
	}
}

func TestDeleteRelationListUnsafe(t *testing.T) {
	type args struct {
		relationList map[int]types.StorageRelation
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteRelationListUnsafe(tt.args.relationList)
		})
	}
}

func TestDeleteRelationUnsafe(t *testing.T) {
	type args struct {
		sourceType int
		sourceID   int
		targetType int
		targetID   int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteRelationUnsafe(tt.args.sourceType, tt.args.sourceID, tt.args.targetType, tt.args.targetID)
		})
	}
}

func TestEntityExists(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EntityExists(tt.args.Type, tt.args.id); got != tt.want {
				t.Errorf("EntityExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntityExistsUnsafe(t *testing.T) {
	type args struct {
		Type int
		id   int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EntityExistsUnsafe(tt.args.Type, tt.args.id); got != tt.want {
				t.Errorf("EntityExistsUnsafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAmountPersistencePayloadsPending(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAmountPersistencePayloadsPending(); got != tt.want {
				t.Errorf("GetAmountPersistencePayloadsPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetChildRelationsBySourceTypeAndSourceId(t *testing.T) {
	type args struct {
		Type    int
		id      int
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChildRelationsBySourceTypeAndSourceId(tt.args.Type, tt.args.id, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChildRelationsBySourceTypeAndSourceId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChildRelationsBySourceTypeAndSourceId() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetChildRelationsBySourceTypeAndSourceIdUnsafe(t *testing.T) {
	type args struct {
		Type    int
		id      int
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChildRelationsBySourceTypeAndSourceIdUnsafe(tt.args.Type, tt.args.id, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChildRelationsBySourceTypeAndSourceIdUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChildRelationsBySourceTypeAndSourceIdUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntitiesByQueryFilter(t *testing.T) {
	type args struct {
		typePool       []string
		conditions     [][][3]string
		idFilter       [][]int
		valueFilter    [][]int
		contextFilter  [][]int
		propertyList   []map[string][]int
		returnDataFlag bool
	}
	tests := []struct {
		name  string
		args  args
		want  []transport.TransportEntity
		want1 [][2]int
		want2 int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := GetEntitiesByQueryFilter(tt.args.typePool, tt.args.conditions, tt.args.idFilter, tt.args.valueFilter, tt.args.contextFilter, tt.args.propertyList, tt.args.returnDataFlag)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByQueryFilter() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetEntitiesByQueryFilter() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GetEntitiesByQueryFilter() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetEntitiesByQueryFilterAndSourceAddress(t *testing.T) {
	type args struct {
		typePool       []string
		conditions     [][][3]string
		idFilter       [][]int
		valueFilter    [][]int
		contextFilter  [][]int
		propertyList   []map[string][]int
		sourceAddress  [2]int
		direction      int
		returnDataFlag bool
	}
	tests := []struct {
		name  string
		args  args
		want  []transport.TransportRelation
		want1 [][2]int
		want2 int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := GetEntitiesByQueryFilterAndSourceAddress(tt.args.typePool, tt.args.conditions, tt.args.idFilter, tt.args.valueFilter, tt.args.contextFilter, tt.args.propertyList, tt.args.sourceAddress, tt.args.direction, tt.args.returnDataFlag)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByQueryFilterAndSourceAddress() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetEntitiesByQueryFilterAndSourceAddress() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GetEntitiesByQueryFilterAndSourceAddress() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetEntitiesByType(t *testing.T) {
	type args struct {
		Type    string
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntitiesByType(tt.args.Type, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntitiesByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntitiesByTypeAndValue(t *testing.T) {
	type args struct {
		Type    string
		value   string
		mode    string
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntitiesByTypeAndValue(tt.args.Type, tt.args.value, tt.args.mode, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntitiesByTypeAndValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByTypeAndValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntitiesByTypeAndValueUnsafe(t *testing.T) {
	type args struct {
		Type    string
		value   string
		mode    string
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntitiesByTypeAndValueUnsafe(tt.args.Type, tt.args.value, tt.args.mode, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntitiesByTypeAndValueUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByTypeAndValueUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntitiesByTypeUnsafe(t *testing.T) {
	type args struct {
		Type    string
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntitiesByTypeUnsafe(tt.args.Type, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntitiesByTypeUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByTypeUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntitiesByValue(t *testing.T) {
	type args struct {
		value   string
		mode    string
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntitiesByValue(tt.args.value, tt.args.mode, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntitiesByValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntitiesByValueUnsafe(t *testing.T) {
	type args struct {
		value   string
		mode    string
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntitiesByValueUnsafe(tt.args.value, tt.args.mode, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntitiesByValueUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntitiesByValueUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityAmount(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEntityAmount(); got != tt.want {
				t.Errorf("GetEntityAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityAmountByType(t *testing.T) {
	type args struct {
		intType int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntityAmountByType(tt.args.intType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntityAmountByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEntityAmountByType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityByPath(t *testing.T) {
	type args struct {
		Type    int
		id      int
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntityByPath(tt.args.Type, tt.args.id, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntityByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntityByPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityByPathUnsafe(t *testing.T) {
	type args struct {
		Type    int
		id      int
		context string
	}
	tests := []struct {
		name    string
		args    args
		want    types.StorageEntity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEntityByPathUnsafe(tt.args.Type, tt.args.id, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntityByPathUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntityByPathUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityRTypes(t *testing.T) {
	tests := []struct {
		name string
		want map[string]int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEntityRTypes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntityRTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityRTypesUnsafe(t *testing.T) {
	tests := []struct {
		name string
		want map[string]int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEntityRTypesUnsafe(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntityRTypesUnsafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityTypes(t *testing.T) {
	tests := []struct {
		name string
		want map[int]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEntityTypes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntityTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEntityTypesUnsafe(t *testing.T) {
	tests := []struct {
		name string
		want map[int]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEntityTypesUnsafe(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntityTypesUnsafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetParentEntitiesByTargetTypeAndTargetIdAndSourceType(t *testing.T) {
	type args struct {
		targetType int
		targetID   int
		sourceType int
		context    string
	}
	tests := []struct {
		name string
		args args
		want map[int]types.StorageEntity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetParentEntitiesByTargetTypeAndTargetIdAndSourceType(tt.args.targetType, tt.args.targetID, tt.args.sourceType, tt.args.context); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetParentEntitiesByTargetTypeAndTargetIdAndSourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetParentEntitiesByTargetTypeAndTargetIdAndSourceTypeUnsafe(t *testing.T) {
	type args struct {
		targetType int
		targetID   int
		sourceType int
		context    string
	}
	tests := []struct {
		name string
		args args
		want map[int]types.StorageEntity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetParentEntitiesByTargetTypeAndTargetIdAndSourceTypeUnsafe(tt.args.targetType, tt.args.targetID, tt.args.sourceType, tt.args.context); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetParentEntitiesByTargetTypeAndTargetIdAndSourceTypeUnsafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetParentRelationsByTargetTypeAndTargetId(t *testing.T) {
	type args struct {
		targetType int
		targetID   int
		context    string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetParentRelationsByTargetTypeAndTargetId(tt.args.targetType, tt.args.targetID, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetParentRelationsByTargetTypeAndTargetId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetParentRelationsByTargetTypeAndTargetId() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetParentRelationsByTargetTypeAndTargetIdUnsafe(t *testing.T) {
	type args struct {
		targetType int
		targetID   int
		context    string
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetParentRelationsByTargetTypeAndTargetIdUnsafe(tt.args.targetType, tt.args.targetID, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetParentRelationsByTargetTypeAndTargetIdUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetParentRelationsByTargetTypeAndTargetIdUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRelation(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
	}
	tests := []struct {
		name    string
		args    args
		want    types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRelation(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRelation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRelationUnsafe(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
	}
	tests := []struct {
		name    string
		args    args
		want    types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRelationUnsafe(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRelationUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRelationUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTypeIdByString(t *testing.T) {
	type args struct {
		strType string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTypeIdByString(tt.args.strType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeIdByString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTypeIdByString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTypeIdByStringUnsafe(t *testing.T) {
	type args struct {
		strType string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTypeIdByStringUnsafe(tt.args.strType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeIdByStringUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTypeIdByStringUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTypeStringById(t *testing.T) {
	type args struct {
		intType int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTypeStringById(tt.args.intType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeStringById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTypeStringById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTypeStringByIdUnsafe(t *testing.T) {
	type args struct {
		intType int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTypeStringByIdUnsafe(tt.args.intType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeStringByIdUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTypeStringByIdUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	type args struct {
		persistenceCfg types.PersistenceConfig
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.args.persistenceCfg)
		})
	}
}

func TestLinkAddressLists(t *testing.T) {
	type args struct {
		from [][2]int
		to   [][2]int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinkAddressLists(tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("LinkAddressLists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapTransportData(t *testing.T) {
	type args struct {
		data transport.TransportEntity
	}
	tests := []struct {
		name string
		args args
		want transport.TransportEntity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapTransportData(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapTransportData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationExists(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RelationExists(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID); got != tt.want {
				t.Errorf("RelationExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationExistsUnsafe(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RelationExistsUnsafe(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID); got != tt.want {
				t.Errorf("RelationExistsUnsafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTraverseEnrich(t *testing.T) {
	type args struct {
		entity    *transport.TransportEntity
		direction int
		depth     int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TraverseEnrich(tt.args.entity, tt.args.direction, tt.args.depth)
		})
	}
}

func TestTypeExists(t *testing.T) {
	type args struct {
		strType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TypeExists(tt.args.strType); got != tt.want {
				t.Errorf("TypeExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeExistsUnsafe(t *testing.T) {
	type args struct {
		strType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TypeExistsUnsafe(tt.args.strType); got != tt.want {
				t.Errorf("TypeExistsUnsafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeIdExists(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TypeIdExists(tt.args.id); got != tt.want {
				t.Errorf("TypeIdExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeIdExistsUnsafe(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TypeIdExistsUnsafe(tt.args.id); got != tt.want {
				t.Errorf("TypeIdExistsUnsafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateEntity(t *testing.T) {
	type args struct {
		entity types.StorageEntity
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateEntity(tt.args.entity); (err != nil) != tt.wantErr {
				t.Errorf("UpdateEntity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateEntityUnsafe(t *testing.T) {
	type args struct {
		entity types.StorageEntity
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateEntityUnsafe(tt.args.entity); (err != nil) != tt.wantErr {
				t.Errorf("UpdateEntityUnsafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateRelation(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
		relation   types.StorageRelation
	}
	tests := []struct {
		name    string
		args    args
		want    types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateRelation(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID, tt.args.relation)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateRelation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateRelationUnsafe(t *testing.T) {
	type args struct {
		srcType    int
		srcID      int
		targetType int
		targetID   int
		relation   types.StorageRelation
	}
	tests := []struct {
		name    string
		args    args
		want    types.StorageRelation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateRelationUnsafe(tt.args.srcType, tt.args.srcID, tt.args.targetType, tt.args.targetID, tt.args.relation)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRelationUnsafe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateRelationUnsafe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deepCopyEntity(t *testing.T) {
	type args struct {
		entity types.StorageEntity
	}
	tests := []struct {
		name string
		args args
		want types.StorageEntity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deepCopyEntity(tt.args.entity); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deepCopyEntity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deepCopyRelation(t *testing.T) {
	type args struct {
		relation types.StorageRelation
	}
	tests := []struct {
		name string
		args args
		want types.StorageRelation
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deepCopyRelation(tt.args.relation); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deepCopyRelation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRRelationTargetIDsBySourceAddressAndTargetType(t *testing.T) {
	type args struct {
		sourceType int
		sourceID   int
		targetType int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRRelationTargetIDsBySourceAddressAndTargetType(tt.args.sourceType, tt.args.sourceID, tt.args.targetType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRRelationTargetIDsBySourceAddressAndTargetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRelationContextByAddressAndDirection(t *testing.T) {
	type args struct {
		sourceType int
		sourceID   int
		targetType int
		targetID   int
		direction  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRelationContextByAddressAndDirection(tt.args.sourceType, tt.args.sourceID, tt.args.targetType, tt.args.targetID, tt.args.direction); got != tt.want {
				t.Errorf("getRelationContextByAddressAndDirection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRelationPropertiesByAddressAndDirection(t *testing.T) {
	type args struct {
		sourceType int
		sourceID   int
		targetType int
		targetID   int
		direction  int
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRelationPropertiesByAddressAndDirection(tt.args.sourceType, tt.args.sourceID, tt.args.targetType, tt.args.targetID, tt.args.direction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRelationPropertiesByAddressAndDirection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRelationTargetIDsBySourceAddressAndTargetType(t *testing.T) {
	type args struct {
		sourceType int
		sourceID   int
		targetType int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRelationTargetIDsBySourceAddressAndTargetType(tt.args.sourceType, tt.args.sourceID, tt.args.targetType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRelationTargetIDsBySourceAddressAndTargetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleImport(t *testing.T) {
	type args struct {
		importChan chan types.PersistencePayload
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleImport(tt.args.importChan)
		})
	}
}

func Test_importEntity(t *testing.T) {
	type args struct {
		payload types.PersistencePayload
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importEntity(tt.args.payload)
		})
	}
}

func Test_importEntityTypes(t *testing.T) {
	type args struct {
		payload types.PersistencePayload
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importEntityTypes(tt.args.payload)
		})
	}
}

func Test_importRelation(t *testing.T) {
	type args struct {
		payload types.PersistencePayload
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importRelation(tt.args.payload)
		})
	}
}

func Test_mapRecursive(t *testing.T) {
	type args struct {
		entity      transport.TransportEntity
		relatedType int
		relatedID   int
		direction   int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapRecursive(tt.args.entity, tt.args.relatedType, tt.args.relatedID, tt.args.direction); got != tt.want {
				t.Errorf("mapRecursive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_match(t *testing.T) {
	type args struct {
		alpha    string
		operator string
		beta     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := match(tt.args.alpha, tt.args.operator, tt.args.beta); got != tt.want {
				t.Errorf("match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchGroup(t *testing.T) {
	type args struct {
		filterGroup []int
		conditions  [][3]string
		test        string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchGroup(tt.args.filterGroup, tt.args.conditions, tt.args.test); got != tt.want {
				t.Errorf("matchGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Cleanup() {
	EntityStorage = make(map[int]map[int]types.StorageEntity)
	EntityIDMax = make(map[int]int)
	EntityTypes = make(map[int]string)
	EntityRTypes = make(map[string]int)
	EntityTypeIDMax = 0
	RelationStorage = make(map[int]map[int]map[int]map[int]types.StorageRelation)
	RelationRStorage = make(map[int]map[int]map[int]map[int]bool)
}
