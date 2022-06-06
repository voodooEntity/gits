package persistence

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/voodooEntity/gits/src/types"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

var PersistenceChan chan types.PersistencePayload
var PersistenceFlag bool

var config types.PersistenceConfig

var importedEntitiesCache = make(map[string]bool)
var importedRelationsCache = make(map[string]bool)

var currentPersistenceFilename int
var currentPersistenceLineCount = 0

func Init(conf types.PersistenceConfig) chan types.PersistencePayload {
	// store the given config ### maybe rename/restructure dunno
	config = conf

	// store current timestamp as current index filename
	currentPersistenceFilename = int(time.Now().UnixMicro())

	// first we make sure we have a storage directories and they are writable
	err := handleDirectory("storage/")
	if nil != err {
		fmt.Printf("Error handling storage base directory", err.Error())
		os.Exit(1)
	}

	// lets created the persistence channel & flag ### rework if we actually need this
	PersistenceFlag = true
	PersistenceChan = make(chan types.PersistencePayload, config.PersistenceChannelBufferSize) // ### 1000000
	// and add the temporary channel for import
	importChan := make(chan types.PersistencePayload, config.PersistenceChannelBufferSize) // ### 1000000
	go startWorker(importChan)

	// return ImportChan
	return importChan
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
//  worker
func startWorker(importChan chan types.PersistencePayload) {
	// ### config.Logger.Print("> Persistance worker started")
	// first we import existing data
	importData(importChan)

	// now we define the current timestamp as current filename
	currentPersistenceFilename = int(time.Now().UnixMicro())

	// now we handle further persistance
	var err error
	for elem := range PersistenceChan {
		// now we have to differ between entity type since
		// they get stored in a seperat file
		if "EntityType" == elem.Type {
			err = handleEntityType(elem)
		} else {
			// each other type will be handled with a new line
			// in our persistant storage log
			err = storeLine(elem)
		}
		if nil != err {
			fmt.Printf("Could not handle given payload error %+v", elem, err)
			os.Exit(1)
		}
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// Handle storage directory
func storeLine(payload types.PersistencePayload) error {
	// letss see if we reaced the max count of lines
	// per file
	currentPersistenceLineCount++

	// as soon we got our first entry we gonne add the filename
	// to our index file
	if 1 == currentPersistenceLineCount {
		f, err := os.OpenFile("storage/index", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println("Could not open storage index to add new logfile")
			panic(err)
		}
		defer f.Close()
		if _, err = f.WriteString(strconv.Itoa(currentPersistenceFilename) + "\n"); err != nil {
			fmt.Println("Could not write to storage index")
			panic(err)
		}
	}

	// now we check if we have to rotate logfile
	if config.RotationEntriesMax == currentPersistenceLineCount {
		currentPersistenceFilename = int(time.Now().UnixMicro())
		currentPersistenceLineCount = 0
	}

	// so we create the logline by json encoding the payload object
	// and base64 encoding it afterwards for some stability safety
	// could be more efficient but ok for the start ### refactor
	bytesPayloadJson, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return err
	}
	base64StringPayload := base64.StdEncoding.EncodeToString(bytesPayloadJson)

	// so we open the file , go is nice to us since we can open with o_append and o_create
	// so we dont need to make sure the file exists already. sometimes magic can be handy.....
	f, err := os.OpenFile("storage/"+strconv.Itoa(currentPersistenceFilename)+".log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Unable to open current storage file %+v", strconv.Itoa(currentPersistenceFilename)+".log")
		return err
	}

	// now we write the base64(json(payload)) into a line
	if _, err = f.WriteString(base64StringPayload + "\n"); err != nil {
		fmt.Printf("Unable to write data to current storage file %+v", strconv.Itoa(currentPersistenceFilename)+".log")
		return err
	}

	f.Close()
	return nil
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// Handle storage directory
func handleDirectory(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// directory doesnt exist, lets create it
		dirErr := os.MkdirAll(directory, os.ModePerm)
		if nil != dirErr {
			return dirErr
		}
	}
	return nil
}

func writeFile(content []byte, path string) error {
	// write content to file
	err := ioutil.WriteFile(path, content, 0644)
	if nil != err {
		return err
	}
	return nil
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// Handle entity type
func handleEntityType(payload types.PersistencePayload) error {
	//ltes marshall the entity type map
	data, err := json.Marshal(payload.EntityTypes)
	if nil != err {
		return err
	}

	// we got the json, lets build the path and write the file
	path := "storage/entityTypes"
	err = writeFile(data, path)
	if nil != err {
		return err
	}

	return nil
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// import the persistance data
func importData(importChan chan types.PersistencePayload) {
	// first of all we gonne import the entity types
	importEntityTypes(importChan)

	// now we gonne parse the storage index
	// so we know all the storage files to parse
	arrFileIndex := parseStorageIndex()
	fileIndexLen := len(arrFileIndex)
	if 0 < fileIndexLen {
		for c := fileIndexLen; c > 0; c-- {
			persistenceFile := arrFileIndex[c-1]
			// make sure we dont try to read emptystring due to
			// trailing \n in storage
			if "" != persistenceFile {
				handlePersistenceFile(persistenceFile, importChan)
			}
		}
	}

	// now we send the Done payload so the storage knows
	// we imported all data
	importChan <- types.PersistencePayload{
		Type: "Done",
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// handle import of a single persistance logfile
func handlePersistenceFile(persistenceFile string, importChan chan types.PersistencePayload) {
	// first we read the file
	fileBytes, err := readFile("storage/" + persistenceFile + ".log")
	if err != nil {
		fmt.Printf("Could not reaed persistence file %+v", persistenceFile)
	}

	// ok seems fine lets put it back to a string
	// and split it linewise
	fileString := string(fileBytes)
	arrFile := strings.Split(fileString, "\n")

	// now we iterate the file backwards since
	lineCount := len(arrFile)
	for c := lineCount; c > 0; c-- {
		handleImportLine(arrFile[c-1], importChan)
	}

}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// parse the storage index file for existing persistence files
// Format: type|action|jsondata
func handleImportLine(data string, importChan chan types.PersistencePayload) {
	// make sure we dont get any invalid emptystring line maybe not neccesary but better
	// be safe than sorry ### refactor
	if "" == data {
		return
	}

	// now we decode a persistance payload per line.
	// first we base64 decode (so \n cant fuck us up - maybe
	// think about a better way ### refactor
	rawDecodedText, err := base64.StdEncoding.DecodeString(data)
	if nil != err {
		fmt.Printf("Cannot decode base64 line from persistence file", data)
		return
	}

	// now we decode the json. i thought about using gob format but
	// have to run several tests if we can safely store multiple gob inside
	// one file. in theory it should work but ye.... ### refactor
	var payload types.PersistencePayload
	err = json.Unmarshal(rawDecodedText, &payload)
	if nil != err {
		fmt.Printf("Cannot decode json line from persistence file", rawDecodedText)
		return
	}

	// now we check if we handled the dataset already. due to our logic
	// parsing backwards starting from newest can skip already handeled datasets
	send := false
	switch payload.Type {
	case "Entity":
		// first we check if that dataset already has been handeled
		typeStr := string(payload.Entity.Type)
		idStr := string(payload.Entity.ID)
		key := typeStr + "-" + idStr
		if _, ok := importedEntitiesCache[key]; !ok {
			importedEntitiesCache[key] = true
			send = true
		}
	case "Relation":
		// first we check if that dataset already has been handeled
		srcTypeStr := string(payload.Relation.SourceType)
		srcIdStr := string(payload.Relation.SourceID)
		trgtTypeStr := string(payload.Relation.TargetType)
		trgtIdStr := string(payload.Relation.TargetID)
		key := srcTypeStr + "-" + srcIdStr + "-" + trgtTypeStr + "-" + trgtIdStr
		if _, ok := importedEntitiesCache[key]; !ok {
			importedEntitiesCache[key] = true
			send = true
		}
	}

	// now we make sure if its a to-handle dataset if the method needs us to
	// to take further action
	if true == send {
		if "Create" == payload.Method || "Update" == payload.Method {
			importChan <- payload
		}
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// parse the storage index file for existing persistence files
func parseStorageIndex() []string {
	// in case this is the first run ever we need to create the storage/index
	if _, err := os.Stat("storage/index"); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create("storage/index")
		if nil != err {
			fmt.Printf("Could not create initial storage index file. Unrecoverable - exiting %+v", err)
			os.Exit(1)
		}
	}
	// first we read the entityTypes file
	storageIndexBytes, err := readFile("storage/index")

	// if it is an error
	if nil != err {
		fmt.Printf("Could not read  storage index file. Unrecoverable - exiting %+v", err)
		os.Exit(1)
	}

	// seems fine lets split it to array and return
	arrPersistenceFileIndex := strings.Split(string(storageIndexBytes), "\n")
	return arrPersistenceFileIndex
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// import the persistance types
func importEntityTypes(importChan chan types.PersistencePayload) {
	// in case this is the first run ever we need to create the storage/index
	if _, err := os.Stat("storage/entityTypes"); errors.Is(err, os.ErrNotExist) {
		err := writeFile([]byte("{}"), "storage/entityTypes")
		if nil != err {
			fmt.Printf("Could not create initial entity types file. Unrecoverable - exiting %+v", err)
			os.Exit(1)
		}
	}
	// first we read the entityTypes file
	entityTypesJsonBytes, err := readFile("storage/entityTypes")

	// if it is an error
	if nil != err {
		fmt.Printf("Could not read entityTypes file - unrecoverable")
		os.Exit(1)
	}

	// seems fine lets unmarshall it
	var entityTypes map[int]string
	err = json.Unmarshal(entityTypesJsonBytes, &entityTypes)
	if nil != err {
		fmt.Printf("Could json parse entityTypes file data - unrecoverable")
		os.Exit(1)
	}

	// ok we got the entity types, lets pack a payload and send it to storage
	payload := types.PersistencePayload{
		EntityTypes: entityTypes,
		Type:        "EntityTypes",
	}
	importChan <- payload

}

func readFile(filePath string) ([]byte, error) {
	// first we read the json data
	data, err := ioutil.ReadFile(filePath)
	if nil != err {
		// ### config.Logger.Print("> Error reading persistant storage file. Check your permissions")
		fmt.Printf("Could not read given file %+v", filePath)
		os.Exit(1)
	}
	return data, nil
}
