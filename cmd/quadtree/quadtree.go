package main

import (
	"encoding/json"
	"fmt"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/storage"
	"github.com/voodooEntity/gits/src/transport"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	fmt.Println("Starting example run")
	tree := NewQuadtree("test", 6, [2]int{20000, 20000})
	fmt.Println("Populated quadtree")
	elements := generateRandomElements(20000, 20000, 30000)
	fmt.Println("Generated random elements", len(elements))
	for _, value := range elements {
		//fmt.Print("* ")
		tree.AddElement(&value)
	}
	fmt.Println("Mapped random elements")
	start := time.Now()
	result := tree.GetElements(1, 1600, 900, 1800)
	elapsed := time.Since(start)
	//pe(result)
	log.Printf("Locate took %s", elapsed)
	log.Println("Locate found ", len(result), " elements")
}

func generateRandomElements(width, height, amountOfElements int) []Element {
	rand.Seed(time.Now().UnixNano())

	elements := make([]Element, amountOfElements)
	for i := 0; i < amountOfElements; i++ {
		elements[i] = Element{
			X:     rand.Intn(width),
			Y:     rand.Intn(height),
			Ident: fmt.Sprintf("Element%d", i+1),
			Properties: map[string]string{
				"property1": fmt.Sprintf("value%d", i+1),
				"property2": fmt.Sprintf("another_value%d", i+1),
			},
		}
	}

	return elements
}

type Quadtree struct {
	Name       string
	Gits       *gits.Gits
	Depth      int
	Resolution [2]int
}

type Element struct {
	X          int
	Y          int
	Ident      string
	Properties map[string]string
}

type Range struct {
	XiStart, XiEnd, YiStart, YiEnd int
	XsStart, XsEnd, YsStart, YsEnd string
}

func NewQuadtree(name string, depth int, resolution [2]int) *Quadtree {
	qt := &Quadtree{
		Name:       name,
		Gits:       gits.NewInstance("tree"),
		Depth:      depth,
		Resolution: resolution,
	}
	qt.populate()
	return qt
}

func (qt *Quadtree) populate() {
	e := transport.TransportEntity{
		Type:  "QuadTree",
		Value: qt.Name,
		Properties: map[string]string{
			"Width":  strconv.Itoa(qt.Resolution[0]),
			"Height": strconv.Itoa(qt.Resolution[1]),
			"Depth":  strconv.Itoa(qt.Depth),
		},
		ChildRelations: make([]transport.TransportRelation, 0),
	}

	qt.rPopulate(&e, 1, 0, qt.Resolution[0], 0, qt.Resolution[1])
	qt.Gits.MapData(e)
}

func (qt *Quadtree) rPopulate(parent *transport.TransportEntity, currDepth int, currXStart int, currXEnd int, currYStart int, currYEnd int) {
	ranges := qt.BuildRanges(currXStart, currXEnd, currYStart, currYEnd)
	nextDepth := currDepth + 1

	for rangeName, rangeValues := range ranges {
		r := transport.TransportRelation{
			Target: transport.TransportEntity{
				ID:    storage.MAP_FORCE_CREATE,
				Type:  "Node",
				Value: rangeName + ":" + rangeValues.XsStart + ":" + rangeValues.XsEnd + ":" + rangeValues.YsStart + ":" + rangeValues.YsEnd,
				Properties: map[string]string{
					"xStart": rangeValues.XsStart,
					"xEnd":   rangeValues.XsEnd,
					"yStart": rangeValues.YsStart,
					"yEnd":   rangeValues.YsEnd,
				},
			},
		}
		if nextDepth > qt.Depth {
			// we still need to add the current children even if we dont further recursive populate
			parent.ChildRelations = append(parent.ChildRelations, r)
			continue
		}
		qt.rPopulate(&r.Target, nextDepth, rangeValues.XiStart, rangeValues.XiEnd, rangeValues.YiStart, rangeValues.YiEnd)
		parent.ChildRelations = append(parent.ChildRelations, r)
	}
}

func (qt *Quadtree) BuildRanges(currXStart, currXEnd, currYStart, currYEnd int) map[string]*Range {
	width := currXEnd - currXStart
	height := currYEnd - currYStart
	xi1 := currXStart
	xi2 := currXStart + width/2
	xi3 := currXEnd
	yi1 := currYStart
	yi2 := currYStart + height/2
	yi3 := currYEnd
	xs1 := strconv.Itoa(xi1)
	xs2 := strconv.Itoa(xi2)
	xs3 := strconv.Itoa(xi3)
	ys1 := strconv.Itoa(yi1)
	ys2 := strconv.Itoa(yi2)
	ys3 := strconv.Itoa(yi3)
	ranges := map[string]*Range{
		"alpha": &Range{
			XiStart: xi1,
			XsStart: xs1,
			XiEnd:   xi2,
			XsEnd:   xs2,
			YiStart: yi1,
			YsStart: ys1,
			YiEnd:   yi2,
			YsEnd:   ys2,
		},
		"beta": &Range{
			XiStart: xi1,
			XsStart: xs1,
			XiEnd:   xi2,
			XsEnd:   xs2,
			YiStart: yi2,
			YsStart: ys2,
			YiEnd:   yi3,
			YsEnd:   ys3,
		},
		"gamma": &Range{
			XiStart: xi2,
			XsStart: xs2,
			XiEnd:   xi3,
			XsEnd:   xs3,
			YiStart: yi1,
			YsStart: ys1,
			YiEnd:   yi2,
			YsEnd:   ys2,
		},
		"delta": &Range{
			XiStart: xi2,
			XsStart: xs2,
			XiEnd:   xi3,
			XsEnd:   xs3,
			YiStart: yi2,
			YsStart: ys2,
			YiEnd:   yi3,
			YsEnd:   ys3,
		},
	}
	return ranges
}

func (qt *Quadtree) AddElement(el *Element) int {
	x := el.X
	y := el.Y
	identifier := el.Ident
	properties := el.Properties
	properties["X"] = strconv.Itoa(el.X)
	properties["Y"] = strconv.Itoa(el.Y)

	nodeID := qt.locateNode(strconv.Itoa(x), strconv.Itoa(y))
	toMap := transport.TransportEntity{
		ID:         storage.MAP_FORCE_CREATE,
		Type:       "Element",
		Value:      identifier,
		Properties: properties,
		ParentRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type: "Node",
					ID:   nodeID,
				},
			},
		},
	}
	ret := qt.Gits.MapData(toMap)
	return ret.ID
}

func (qt *Quadtree) locateNode(x, y string) int {
	qry := qt.Gits.Query().New().Read("QuadTree")
	qt.rBuildTargetNodeQuery(1, qry, x, y)
	return qt.rGetDeepestNodeID(qt.Gits.Query().Execute(qry).Entities[0])
}

func (qt *Quadtree) rGetDeepestNodeID(entity transport.TransportEntity) int {
	if 0 < len(entity.ChildRelations) {
		return qt.rGetDeepestNodeID(entity.ChildRelations[0].Target)
	}
	return entity.ID
}

func (qt *Quadtree) rBuildTargetNodeQuery(currDepth int, qry *query.Query, x string, y string) {
	nextDepth := currDepth + 1
	subQry := qt.Gits.Query().New().Read("Node").
		Match("Properties.xStart", "<=", x).
		Match("Properties.xEnd", ">=", x).
		Match("Properties.yStart", "<=", y).
		Match("Properties.yEnd", ">=", y)
	qry.To(subQry)
	if nextDepth <= qt.Depth {
		qt.rBuildTargetNodeQuery(nextDepth, subQry, x, y)
	}
}

func (qt *Quadtree) GetElements(xStart, xEnd, yStart, yEnd int) []*Element {
	qry := qt.Gits.Query().New().Read("QuadTree")
	qt.rBuildRangeNodeQuery(1, qry, strconv.Itoa(xStart), strconv.Itoa(xEnd), strconv.Itoa(yStart), strconv.Itoa(yEnd))
	res := qt.Gits.Query().Execute(qry)
	elements := make([]*Element, 0)
	for _, e := range res.Entities {
		elements = qt.rCollectElements(e, elements)
	}
	return elements
}

func (qt *Quadtree) rCollectElements(tree transport.TransportEntity, elements []*Element) []*Element {
	if "Element" == tree.ChildRelations[0].Target.Type {
		for _, rel := range tree.ChildRelations {
			x, _ := strconv.Atoi(rel.Target.Properties["X"])
			y, _ := strconv.Atoi(rel.Target.Properties["Y"])
			elements = append(elements, &Element{
				X:          x,
				Y:          y,
				Ident:      rel.Target.Value,
				Properties: rel.Target.Properties,
			})
		}
	} else {
		for _, rel := range tree.ChildRelations {
			elements = qt.rCollectElements(rel.Target, elements)
		}
	}
	return elements
}

func (qt *Quadtree) rBuildRangeNodeQuery(currDepth int, qry *query.Query, xStart string, xEnd string, yStart string, yEnd string) {
	nextDepth := currDepth + 1
	subQry := qt.Gits.Query().New().Read("Node").
		Match("Properties.xStart", "<=", xEnd).
		Match("Properties.xEnd", ">=", xStart).
		Match("Properties.yStart", "<=", yEnd).
		Match("Properties.yEnd", ">=", yStart)

	qry.To(subQry)
	if nextDepth <= qt.Depth {
		qt.rBuildRangeNodeQuery(nextDepth, subQry, xStart, xEnd, yStart, yEnd)
	} else {
		subQry.To(qt.Gits.Query().New().Read("Element").
			Match("Properties.X", "<=", xEnd).
			Match("Properties.X", ">=", xStart).
			Match("Properties.Y", "<=", yEnd).
			Match("Properties.Y", ">=", yStart))
	}
}

func (qt *Quadtree) DeleteElement(ident string) {
	qt.Gits.Query().Execute(qt.Gits.Query().New().Delete("Element").Match("Value", "==", ident))
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// + + + + + + + + + +   UTILS   + + + + + + + + + + +
// - - - - - - - - - - - - - - - - - - - - - - - - - -
func pr(result transport.Transport) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}

func pq(qry query.Query) {
	jsonData, err := json.MarshalIndent(qry, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}

func pe(d []*Element) {
	jsonData, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}
