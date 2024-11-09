package main

import (
	"encoding/json"
	"fmt"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/storage"
	"github.com/voodooEntity/gits/src/transport"
	"strconv"
)

var g *gits.Gits
var qa *gits.QueryAdapter

func main() {
	g = gits.NewInstance("examples")
	qa = g.Query()

	// - - - - - - - - - - - - - - - - - - - - - - - - - -
	// + + + + + + + + + +  EXAMPLES + + + + + + + + + + +
	// - - - - - - - - - - - - - - - - - - - - - - - - - -
	//
	// This document contains a set of examples grouped by
	// category. These examples mirror the one'g given
	// in the documentation.
	//
	// If you want to check out one of these examples,
	// simply uncomment the call in the main and run it.
	//
	// Note: The examples are written to work when run
	//        alone, not in combination with other examples.
	//        Uncommenting multiple at the same time and
	//        run them might result in unexpected results
	//        due to overlapping seedings.
	// - - - - - - - - - - - - - - - - - - - - - - - - - -

	//queryExample1()
	//queryExample2()
	//queryExample3()
	//queryExample4()
	//queryExample5()
	//queryExample6()
	//queryExample7()
	//queryExample8()
	//queryExample9()
	//queryExample10()
	//queryExample11()
	//queryExample12()
	//queryExample13()
	//queryExample14()
	//queryExample15()
	//queryExample16()
	//queryExample17()
	//queryExample18()
	//queryExample19()
	//queryExample20()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// + + + + + + + + + +   QUERY   + + + + + + + + + + +
// - - - - - - - - - - - - - - - - - - - - - - - - - -

func queryExample1() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Something",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "someValue",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "lorem",
	})
	qry := qa.New().Read("Alpha")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample2() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Something",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "someValue",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Beta",
		Value: "ipsum",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Beta",
		Value: "appropinquare",
	})
	qry := qa.New().Read("Alpha", "Beta")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample3() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Something",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "someValue",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "lorem",
	})
	qry := qa.New().Read("Alpha").Match("Value", "==", "someValue")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample4() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Something",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "someValue",
	})
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "lorem",
		Context: "someContext",
	})
	qry := qa.New().Read("Alpha").Match("Context", "==", "someContext")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample5() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Something",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "someValue",
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "lorem",
		Context:    "someContext",
		Properties: map[string]string{"MyPropertyName": "propertyValue"},
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "another",
		Properties: map[string]string{"MyPropertyName": "propertyValue"},
	})
	qry := qa.New().Read("Alpha").Match("Properties.MyPropertyName", "==", "propertyValue")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample6() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Something",
	})
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "someValue",
		Context: "someContext",
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "lorem",
		Context:    "someContext",
		Properties: map[string]string{"MyPropertyName": "propertyValue"},
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "another",
		Properties: map[string]string{"MyPropertyName": "propertyValue"},
	})
	qry := qa.New().Read("Alpha").Match("Value", "==", "someValue").Match("Context", "==", "someContext")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample7() {
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "Something",
		Context: "Lorem",
	})
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "someValue",
		Context: "dolor",
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "someValue",
		Context:    "Lorem",
		Properties: map[string]string{"MyPropertyName": "propertyValue"},
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "finally",
		Context:    "ipsum",
		Properties: map[string]string{"MyPropertyName": "propertyValue"},
	})

	qry := qa.New().Read("Alpha").Match("Context", "==", "Lorem").Match("Value", "==", "someValue").
		OrMatch("Context", "==", "ipsum").Match("Value", "==", "finally")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample8() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "some",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "thing",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "else",
				},
			},
		},
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "lorem",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "ipsum",
				},
			},
		},
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Well",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Beta",
		Value: "Done",
	})
	qry := qa.New().Read("Alpha").To(qa.New().Read("Beta"))
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample9() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "some",
		ParentRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "thing",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "else",
				},
			},
		},
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "lorem",
		ParentRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "ipsum",
				},
			},
		},
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Well",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Beta",
		Value: "Done",
	})
	qry := qa.New().Read("Alpha").From(qa.New().Read("Beta"))
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample10() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "some",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "thing",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "else",
				},
			},
		},
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Well",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Done",
	})
	qry := qa.New().Read("Alpha").CanTo(qa.New().Read("Beta"))
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample11() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "some",
		ParentRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "thing",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "else",
				},
			},
		},
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Well",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "Done",
	})
	qry := qa.New().Read("Alpha").CanFrom(qa.New().Read("Beta"))
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample12() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "some",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "thing",
					ParentRelations: []transport.TransportRelation{
						{
							Target: transport.TransportEntity{
								Type:  "Gamma",
								Value: "thatsit",
							},
						},
					},
				},
			},
		},
	})
	qry := qa.New().Read("Alpha").To(qa.New().Read("Beta").From(qa.New().Read("Gamma")))
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample13() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "some",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "someValue",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "else",
				},
			},
		},
	})
	qry := qa.New().Read("Alpha").To(qa.New().Read("Beta").Match("Value", "==", "someValue"))
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample14() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "never",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:  "Gamma",
					Value: "gonne",
					ChildRelations: []transport.TransportRelation{
						{
							Target: transport.TransportEntity{
								Type:  "Epsilon",
								Value: "give",
								ChildRelations: []transport.TransportRelation{
									{
										Target: transport.TransportEntity{
											Type:  "Psi",
											Value: "you",
										},
									},
									{
										Target: transport.TransportEntity{
											Type:  "Poi",
											Value: "up",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Beta",
					Value: "never",
					ChildRelations: []transport.TransportRelation{
						{
							Target: transport.TransportEntity{
								Type:  "Foo",
								Value: "gonne",
							},
						},
						{
							Target: transport.TransportEntity{
								Type:  "Bar",
								Value: "let",
							},
						},
					},
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Kato",
					Value: "you",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:  "Osa",
					Value: "down",
				},
			},
		},
	})

	qry := qa.New().Read("Alpha").TraverseOut(3)
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample15() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "old",
	})

	qry := qa.New().Update("Alpha").Match("Value", "==", "old").Set("Value", "Lorem").Set("Context", "Ipsum").Set("Properties.dolor", "appropinquare")
	result := qa.Execute(qry)
	printResult(result)
	qry = qa.New().Read("Alpha").Match("Context", "==", "Ipsum")
	result = qa.Execute(qry)
	printResult(result)
}

func queryExample16() {
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "one",
		Context: "deleteme",
	})
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "two",
		Context: "keepme",
	})
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "three",
		Context: "deleteme",
	})
	g.MapData(transport.TransportEntity{
		Type:    "Alpha",
		Value:   "four",
		Context: "keepme",
	})

	qry := qa.New().Delete("Alpha").Match("Context", "==", "deleteme")
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample17() {
	g.MapData(transport.TransportEntity{
		ID:    storage.MAP_FORCE_CREATE,
		Type:  "Alpha",
		Value: "psi",
	})
	target := g.MapData(transport.TransportEntity{
		Type:  "Beta",
		Value: "omega",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "phi",
	})
	g.MapData(transport.TransportEntity{
		Type:  "Beta",
		Value: "muh",
	})
	qry := qa.New().Link("Alpha").Match("Value", "==", "psi").To(
		qa.New().Find("Beta").Match("Value", "==", "omega"),
	)
	result := qa.Execute(qry)
	printResult(result)
	qry = qa.New().Read("Beta").Match("ID", "==", strconv.Itoa(target.ID)).TraverseIn(1)
	printResult(qa.Execute(qry))
}

func queryExample18() {
	g.MapData(transport.TransportEntity{
		Type:  "Alpha",
		Value: "psi",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:    "Beta",
					Value:   "thing",
					Context: "omega",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:    "Beta",
					Value:   "cool",
					Context: "omega",
				},
			},
			{
				Target: transport.TransportEntity{
					Type:    "Beta",
					Value:   "lorem",
					Context: "psi",
				},
			},
		},
	})
	qry := qa.New().Unlink("Alpha").Match("Value", "==", "psi").To(
		qa.New().Find("Beta").Match("Context", "==", "omega"),
	)
	result := qa.Execute(qry)
	printResult(result)
	qry = qa.New().Read("Alpha").TraverseOut(1)
	printResult(qa.Execute(qry))
}

func queryExample19() {
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "gonne",
		Properties: map[string]string{"Psi": "2"},
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "up",
		Properties: map[string]string{"Psi": "5"},
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "give",
		Properties: map[string]string{"Psi": "3"},
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "never",
		Properties: map[string]string{"Psi": "1"},
	})
	g.MapData(transport.TransportEntity{
		Type:       "Alpha",
		Value:      "you",
		Properties: map[string]string{"Psi": "4"},
	})

	qry := qa.New().Read("Alpha").Order("Properties.Psi", query.ORDER_DIRECTION_ASC, query.ORDER_MODE_NUM)
	result := qa.Execute(qry)
	printResult(result)
}

func queryExample20() {
	g.MapData(transport.TransportEntity{
		ID:    storage.MAP_FORCE_CREATE,
		Type:  "Alpha",
		Value: "could",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					ID:    storage.MAP_FORCE_CREATE,
					Type:  "Beta",
					Value: "should",
					ChildRelations: []transport.TransportRelation{
						{
							Target: transport.TransportEntity{
								ID:         storage.MAP_FORCE_CREATE,
								Type:       "Gamma",
								Value:      "would",
								Properties: map[string]string{"queryExample": "that"},
							},
						},
					},
				},
			},
			{
				Target: transport.TransportEntity{
					ID:      storage.MAP_FORCE_CREATE,
					Type:    "Epsilon",
					Value:   "never",
					Context: "notbeta",
					ChildRelations: []transport.TransportRelation{
						{
							Target: transport.TransportEntity{
								ID:         storage.MAP_FORCE_CREATE,
								Type:       "Gamma",
								Value:      "gonne",
								Properties: map[string]string{"queryExample": "that"},
								ChildRelations: []transport.TransportRelation{
									{
										Target: transport.TransportEntity{
											ID:         storage.MAP_FORCE_CREATE,
											Type:       "Gamma",
											Value:      "give",
											Properties: map[string]string{"queryExample": "that"},
											ChildRelations: []transport.TransportRelation{
												{
													Target: transport.TransportEntity{
														ID:         storage.MAP_FORCE_CREATE,
														Type:       "Gamma",
														Value:      "you",
														Properties: map[string]string{"queryExample": "that"},
													},
												},
												{
													Target: transport.TransportEntity{
														ID:         storage.MAP_FORCE_CREATE,
														Type:       "Gamma",
														Value:      "up",
														Properties: map[string]string{"queryExample": "that"},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Target: transport.TransportEntity{
								ID:         storage.MAP_FORCE_CREATE,
								Type:       "Gamma",
								Value:      "never gonne",
								Properties: map[string]string{"queryExample": "that"},
							},
						},
						{
							Target: transport.TransportEntity{
								ID:         storage.MAP_FORCE_CREATE,
								Type:       "Gamma",
								Value:      "let you down",
								Properties: map[string]string{"queryExample": "that"},
							},
						},
					},
				},
			},
		},
		ParentRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					ID:    storage.MAP_FORCE_CREATE,
					Type:  "Psi",
					Value: "ipsum",
				},
			},
		},
	})

	g.MapData(transport.TransportEntity{
		ID:    storage.MAP_FORCE_CREATE,
		Type:  "Alpha",
		Value: "could",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					ID:    storage.MAP_FORCE_CREATE,
					Type:  "Beta",
					Value: "should",
					ChildRelations: []transport.TransportRelation{
						{
							Target: transport.TransportEntity{
								ID:         storage.MAP_FORCE_CREATE,
								Type:       "Gamma",
								Value:      "would",
								Properties: map[string]string{"queryExample": "that"},
							},
						},
					},
				},
			},
			{
				Target: transport.TransportEntity{
					ID:      storage.MAP_FORCE_CREATE,
					Type:    "Epsilon",
					Value:   "never",
					Context: "notbeta",
				},
			},
		},
	})

	qry := qa.New().Read("Alpha").To(
		qa.New().Reduce("Beta").To(
			qa.New().Reduce("Gamma").Match("Properties.queryExample", "==", "this").OrMatch("Properties.queryExample", "==", "that"),
		),
	).To(
		qa.New().Read("Epsilon").Match("Context", "!=", "beta").TraverseOut(3),
	).CanFrom(
		qa.New().Read("Psi").Match("Value", "==", "ipsum"),
	)
	result := qa.Execute(qry)
	printResult(result)
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// + + + + + + + + + +   UTILS   + + + + + + + + + + +
// - - - - - - - - - - - - - - - - - - - - - - - - - -

func printResult(result transport.Transport) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}

func printQuery(qry query.Query) {
	jsonData, err := json.MarshalIndent(qry, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}
