package main

import (
	"encoding/json"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
)

func main() {

	//testNewMappedStructure()
	//testNewMappedStructureWithExistingEntity()
	testNewMappedStructureWithExistingMappedStructureEntity()

}

func testNewMappedStructure() {
	archivist.Info("Creating completly new mapped data")
	// create testdata
	testdata := transport.TransportEntity{
		ID:    -1,
		Type:  "IP",
		Value: "127.0.0.1",
		ChildRelations: []transport.TransportRelation{
			{
				Context: "Something",
				Target: transport.TransportEntity{
					ID:    -1,
					Type:  "Port",
					Value: "80",
					ChildRelations: []transport.TransportRelation{
						{
							Context: "Else",
							Target: transport.TransportEntity{
								ID:    -1,
								Type:  "Software",
								Value: "Apache",
							},
						},
					},
				},
			},
		},
	}
	id := gits.MapTransportData(testdata)
	archivist.Info("Retrieved new ID", id)
	archivist.Info("Reading out the data using query")
	search := query.New().Read("IP").Match("Value", "==", "127.0.0.1").To(
		query.New().Read("Port").To(
			query.New().Read("Software"),
		),
	)
	ret := query.Execute(search)
	printData(ret)
}

func testNewMappedStructureWithExistingEntity() {
	archivist.Info("Creating  new mapped data with 1 entity existing inbetween")
	archivist.Info("Precreating Port entity")
	portTypeID, _ := gits.CreateEntityType("Port")
	entityID, _ := gits.CreateEntity(types.StorageEntity{
		ID:    -1,
		Type:  portTypeID,
		Value: "80",
	})
	archivist.Info("Approving that port entity got created inb4")
	portEntity, _ := gits.GetEntityByPath(portTypeID, entityID, "")
	printData(portEntity)
	archivist.Info("Now creating actual mapped data with existing port entity mapped inbetweet")
	// create testdata
	testdata := transport.TransportEntity{
		ID:    -1,
		Type:  "IP",
		Value: "127.0.0.1",
		ChildRelations: []transport.TransportRelation{
			{
				Context: "Something",
				Target: transport.TransportEntity{
					ID:   portEntity.ID,
					Type: "Port",
					ChildRelations: []transport.TransportRelation{
						{
							Context: "Else",
							Target: transport.TransportEntity{
								ID:    -1,
								Type:  "Software",
								Value: "Apache",
							},
						},
					},
				},
			},
		},
	}
	id := gits.MapTransportData(testdata)
	archivist.Info("Retrieved new ID", id)
	archivist.Info("Reading out the data using query")
	search := query.New().Read("IP").Match("Value", "==", "127.0.0.1").To(
		query.New().Read("Port").To(
			query.New().Read("Software"),
		),
	)
	ret := query.Execute(search)
	printData(ret)
}

func testNewMappedStructureWithExistingMappedStructureEntity() {
	archivist.Info("Creating  new mapped data with 1 entity existing inbetween")
	archivist.Info("Precreating Port entity")
	portTypeID, _ := gits.CreateEntityType("Port")
	portID, _ := gits.CreateEntity(types.StorageEntity{
		ID:    -1,
		Type:  portTypeID,
		Value: "80",
	})
	archivist.Info("Approving that port entity got created inb4")
	portEntity, _ := gits.GetEntityByPath(portTypeID, portID, "")
	printData(portEntity)
	archivist.Info("Precreating Software entity")
	softwareTypeID, _ := gits.CreateEntityType("Software")
	softwareID, _ := gits.CreateEntity(types.StorageEntity{
		ID:    -1,
		Type:  softwareTypeID,
		Value: "Apache",
	})
	archivist.Info("Approving that software entity got created inb4")
	softwareEntity, _ := gits.GetEntityByPath(softwareTypeID, softwareID, "")
	printData(softwareEntity)
	archivist.Info("Now creating actual mapped data with existing port entity mapped inbetween")
	// create testdata
	testdata := transport.TransportEntity{
		ID:    -1,
		Type:  "IP",
		Value: "127.0.0.1",
		ChildRelations: []transport.TransportRelation{
			{
				Context: "Something",
				Target: transport.TransportEntity{
					ID:   portEntity.ID,
					Type: "Port",
					ChildRelations: []transport.TransportRelation{
						{
							Context: "Else",
							Target: transport.TransportEntity{
								ID:   softwareEntity.ID,
								Type: "Software",
								ChildRelations: []transport.TransportRelation{
									{
										Target: transport.TransportEntity{
											ID:    -1,
											Type:  "Status",
											Value: "Active",
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
	id := gits.MapTransportData(testdata)
	archivist.Info("Retrieved new ID", id)
	archivist.Info("Reading out the data using query")
	search := query.New().Read("IP").Match("Value", "==", "127.0.0.1").To(
		query.New().Read("Port").To(
			query.New().Read("Software").To(
				query.New().Read("Status"),
			),
		),
	)
	ret := query.Execute(search)
	printData(ret)
}

func printData(data any) {
	t, _ := json.MarshalIndent(data, "", "\t")
	archivist.Info("Query Data Struct", string(t))
}
