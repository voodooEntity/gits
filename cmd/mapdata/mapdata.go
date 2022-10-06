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
	//testNewMappedStructureWithExistingMappedStructureEntity()
	testBigStrucutreMap()
	//testQbStructureMap()
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

func testBigStrucutreMap() {
	archivist.Info("Print testdata")
	// create testdata
	testdata := transport.TransportEntity{
		ID:    -1,
		Type:  "IP",
		Value: "127.0.0.1",
		ChildRelations: []transport.TransportRelation{
			{
				Context: "",
				Target: transport.TransportEntity{
					ID:    -1,
					Type:  "Port",
					Value: "80",
					ChildRelations: []transport.TransportRelation{
						{
							Context: "Else",
							Target: transport.TransportEntity{
								ID:         -1,
								Type:       "Software",
								Value:      "Apache",
								Properties: map[string]string{"Version": "2.4.6"},
							},
						},
					},
				},
			}, {
				Context: "",
				Target: transport.TransportEntity{
					ID:    -1,
					Type:  "Port",
					Value: "443",
					ChildRelations: []transport.TransportRelation{
						{
							Context: "Else",
							Target: transport.TransportEntity{
								ID:         -1,
								Type:       "Software",
								Value:      "Apache",
								Properties: map[string]string{"Version": "2.4.6"},
								ChildRelations: []transport.TransportRelation{
									{
										Context: "Else",
										Target: transport.TransportEntity{
											ID:    -1,
											Type:  "Vhost",
											Value: "laughingman.dev",
										},
									},
								},
							},
						},
					},
				},
			}, {
				Context: "",
				Target: transport.TransportEntity{
					ID:    -1,
					Type:  "Port",
					Value: "8090",
					ChildRelations: []transport.TransportRelation{
						{
							Context: "Else",
							Target: transport.TransportEntity{
								ID:         -1,
								Type:       "Software",
								Value:      "gitsapi",
								Properties: map[string]string{"Version": "0.0.9"},
							},
						},
					},
				},
			},
		},
	}
	printData(testdata)
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

func printData(data any) {
	t, _ := json.MarshalIndent(data, "", "\t")b
	archivist.Info("Query Data Struct", string(t))
}
