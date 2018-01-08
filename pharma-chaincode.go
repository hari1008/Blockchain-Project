package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
  
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const STATUS_SHIPPED = "shipped"
const STATUS_ACCEPTED = "accepted"
const STATUS_REJECTED = "rejected"
const STATUS_DISPATCHED = "dispatched"
const UNIQUE_ID_COUNTER string = "UniqueIDCounter"
const CONTAINER_OWNER = "ContainerOwner"

type PharmaChaincode struct {
}

type UniqueIDCounter struct {
	ContainerMaxID int `json:"ContainerMaxID"`
	PalletMaxID    int `json:"PalletMaxID"`
}

type Shipment struct{
	ContainerList []Container `json:"container_list"`

}

type Container struct {
	ContainerId       string              `json:"container_id"`
	ParentContainerId string              `json:"parent_container_id"`
	ChildContainerId  []string            `json:"child_container_id"`
	Recipient         string              `json:"recipient_id"`
	Elements          ContainerElements   `json:"elements"`
	Provenance        ContainerProvenance `json:"provenance"`
	CertifiedBy       string              `json:"certified_by"`   
	Address           string              `json:"address"`        
	USN               string              `json:"usn"`            
	ShipmentDate      string              `json:"shipment_date"`  
	InvoiceNumber     string              `json:"invoice_number"` 
	Remarks           string              `json:"remarks"`        
  
}

type ContainerElements struct {
	Pallets []Pallet `json:"pallets"`
}

type Pallet struct {
	PalletId string `json:"pallet_id"`
	Cases    []Case `json:"cases"`
}

type Case struct {
	CaseId string `json:"case_id"`
	Units  []Unit `json:"units"`
}

type Unit struct {
	DrugId       string `json:"drug_id"`
	DrugName     string `json:"drug_name"` 
	UnitId       string `json:"unit_id"`
	ExpiryDate   string `json:"expiry_date"`
	HealthStatus string `json:"health_status"`
	BatchNumber  string `json:"batch_number"`
	LotNumber    string `json:"lot_number"`
	SaleStatus   string `json:"sale_status"`
	ConsumerName string `json:"consumer_name"`
}

type ContainerProvenance struct {
	TransitStatus string          `json:transit_status`
	Sender        string          `json:sender`
	Receiver      string          `json:receiver`
	Supplychain   []ChainActivity `json:supplychain`
}

type ChainActivity struct {
	Sender   string `json:sender`
	Receiver string `json:receiver`
	Status   string `json:transit_status`
	ActivityTimeStamp time.Time `json:activity_timeStamp`
	//ActivityTimeStamp1 time.Time `json:activity_timeStamp`
}

type ContainerOwners struct {
	Owners []Owner `json:owners`
}

type Owner struct {
	OwnerId       string   `json:owner_id`
	ContainerList []string `json:container_id`
}

func main() {
	fmt.Println("Inside PharmaChaincode main function")
	err := shim.Start(new(PharmaChaincode))
	if err != nil {
		fmt.Printf("Error starting MedLabPharma chaincode: %s", err)
	}
}

// Init resets all the things
func (t *PharmaChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "init" {
		return t.init(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Invoke isur entry point to invoke a chaincode function
func (t *PharmaChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "ShipContainerUsingLogistics" {
		return t.ShipContainerUsingLogistics(stub, args[0], args[1], args[2], args[3], args[4])
	} else if function == "SetCurrentOwner"{
		return t.SetCurrentOwnerTest(stub, args[0], args[1])
	} else if function == "AcceptContainerbyLogistics"{
		return t.AcceptContainerbyLogistics(stub, args[0], args[1],args[2], args[3])
	}else if function == "DispatchContainer"{
		return t.DispatchContainer(stub, args[0], args[1],args[2])
	}else if function == "AcceptContainerbyDistributor"{
		return t.AcceptContainerbyDistributor(stub, args[0], args[1],args[2]) 
	}else if function == "RejectContainerbyLogistics"{
		return t.RejectContainerbyLogistics(stub, args[0], args[1],args[2],args[3]) 
	}else if function == "RejectContainerbyDistributor"{
		return t.RejectContainerbyDistributor(stub, args[0], args[1],args[2]) 
	}	 
	fmt.Println("invoke did not find func: " + function)
	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *PharmaChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "GetContainerDetails" { //read a variable
		return t.GetContainerDetails(stub, args[0])
	} else if function == "GetMaxIDValue" {
		return t.GetMaxIDValue(stub)
	} else if function == "GetEmptyContainer" {
		return t.GetEmptyContainer(stub)
	}  else if function == "GetContainerDetailsForOwner" {
		return t.GetContainerDetailsForOwner(stub, args[0])
	}else if function == "GetOwner" {
		return t.GetOwner(stub)
	}else if function == "GetUserAttribute" {
		return t.GetUserAttribute(stub, args[0])
	}
	
	fmt.Println("query did not find func: " + function)
	return nil, errors.New("Received unknown function query: " + function)
}

func (t *PharmaChaincode) init(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	maxIDCounter := UniqueIDCounter{
		ContainerMaxID: 0,
		PalletMaxID:    0}
	jsonVal, _ := json.Marshal(maxIDCounter)
	err := stub.PutState(UNIQUE_ID_COUNTER, []byte(jsonVal))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// write  invoke function to write key/value pair
func (t *PharmaChaincode) ShipContainerUsingLogistics(stub shim.ChaincodeStubInterface,
	senderID string, logisticsID string, receiverID string, remarks string, elementsJSON string) ([]byte, error) {
	var err error

	containerID, jsonValue := ShipContainerUsingLogistics_Internal(senderID, logisticsID, receiverID, remarks, elementsJSON)
	fmt.Println("running ShipContainerUsingLogistics.key:" + containerID)
	fmt.Println(jsonValue)
	err = stub.PutState(containerID, jsonValue) //write the variable into the chaincode state

	incrementCounter(stub) //increment the unique ids for container and Pallet

	setCurrentOwner(stub, senderID, containerID)
	setCurrentOwner(stub, logisticsID, containerID)

	if err != nil {
		return nil, err
	}
	return nil, nil

}
func (t *PharmaChaincode)DispatchContainer(stub shim.ChaincodeStubInterface,containerID string, receiverID string, remarks string) ([]byte, error) {
	var err error
	fmt.Println("running DispatchContainer:" + containerID)
     valAsbytes, err := stub.GetState(containerID)
	 if len(valAsbytes) == 0 {
		 	jsonResp := "{\"Error\":\"Failed to get state for Container id since there is no such container \"}"
		return nil, errors.New(jsonResp)
	 }
	 fmt.Println("json value from the container")
	 fmt.Println(valAsbytes)
	 if err != nil{
		jsonResp := "{\"Error\":\"Failed to get state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	 shipment := Container{}	  
	json.Unmarshal([]byte(valAsbytes), &shipment)
	shipment.Recipient = receiverID
	conprov := shipment.Provenance  
    supplychain := conprov.Supplychain     
	chainActivity := ChainActivity{
		Sender:   shipment.Provenance.Receiver,//
		Receiver: receiverID,
		Status:   STATUS_DISPATCHED,
		ActivityTimeStamp:time.Now().UTC()}  
	supplychain = append(supplychain, chainActivity) 
	conprov.Supplychain = supplychain
   conprov.TransitStatus = STATUS_DISPATCHED
   conprov.Sender = shipment.Provenance.Receiver
   conprov.Receiver = receiverID
   shipment.Provenance = conprov
    jsonVal, _ := json.Marshal(shipment)
   	err = stub.PutState(containerID, jsonVal)//write the variable into the chaincode state
    if err != nil{
		jsonResp := "{\"Error\":\"Failed to put state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("********DISPATCHED JSON***********")	
	fmt.Println("SENDER",shipment.Provenance.Receiver)	
	fmt.Println(string(jsonVal))	
	incrementCounter(stub) //increment the unique ids for container and Pallet
	setCurrentOwner(stub, receiverID, containerID)

	if err != nil {
		return nil, err
	}
	return nil, nil

}



// read  query function to read key/value pair
func (t *PharmaChaincode) GetContainerDetails(stub shim.ChaincodeStubInterface, container_id string) ([]byte, error) {
	fmt.Println("runnin GetContainerDetails ")
	var key, jsonResp string
	var err error

	if container_id == "" {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	fmt.Println("key:" + container_id)
	valAsbytes, err := stub.GetState(container_id)
	fmt.Println(valAsbytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

//Returns the maximum number used for ContainerID and PalletID in the format "ContainerMaxNumber, PalletMaxNumber"
func (t *PharmaChaincode) GetMaxIDValue(stub shim.ChaincodeStubInterface) ([]byte, error) {
	var jsonResp string
	var err error
	ConMaxAsbytes, err := stub.GetState(UNIQUE_ID_COUNTER)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for ContainerMaxNumber \"}"
		return nil, errors.New(jsonResp)
	}
	return ConMaxAsbytes, nil
}

func (t *PharmaChaincode) GetEmptyContainer(stub shim.ChaincodeStubInterface) ([]byte, error) {
	ConMaxAsbytes, err := stub.GetState(UNIQUE_ID_COUNTER)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for ContainerMaxNumber \"}"
		return nil, errors.New(jsonResp)
	}

	counter := UniqueIDCounter{}
	json.Unmarshal([]byte(ConMaxAsbytes), &counter)
	containerID := "CON" + strconv.Itoa(counter.ContainerMaxID+1)
	pallets := createPallet(containerID, counter.PalletMaxID+1)
	conelement := ContainerElements{Pallets: pallets}
	container := Container{
		ContainerId: containerID,
		Elements:    conelement}
	jsonVal, _ := json.Marshal(container)
	return jsonVal, nil
}

func ShipContainerUsingLogistics_Internal(senderID string,
	logisticsID string, receiverID string, remarks string, elementsJSON string) (string, []byte) {
		//ActivityTimeStamp1=time.Now().UTC()
	chainActivity := ChainActivity{
		Sender:   senderID,
		Receiver: logisticsID,
		Status:   STATUS_SHIPPED,
		ActivityTimeStamp:time.Now().UTC()}
		//ActivityTimeStamp: ActivityTimeStamp1.Format("20060102 15:04:05")} 
	var supplyChain []ChainActivity
	supplyChain = append(supplyChain, chainActivity)
	conprov := ContainerProvenance{
		TransitStatus: STATUS_SHIPPED,
		Sender:        senderID,
		Receiver:      logisticsID,
		Supplychain:   supplyChain}
	shipment := Container{}
	json.Unmarshal([]byte(elementsJSON), &shipment)
	shipment.Recipient = receiverID
	shipment.Provenance = conprov
	jsonVal, _ := json.Marshal(shipment)
	return shipment.ContainerId, jsonVal
}

func createUnit(caseID string) []Unit {
	units := make([]Unit, 3)

	for index := 0; index < 3; index++ {
		strIndex := strconv.Itoa(index + 1)
		unitid := caseID + "UNIT" + strIndex
		units[index].UnitId = unitid
	}
	return units
}

func createCase(palletID string) []Case {
	cases := make([]Case, 3)

	for index := 0; index < 3; index++ {
		strIndex := strconv.Itoa(index + 1)
		caseid := palletID + "CASE" + strIndex
		cases[index].CaseId = caseid
		cases[index].Units = createUnit(caseid)
	}
	return cases
}

func createPallet(containerID string, palletMaxID int) []Pallet {
	pallets := make([]Pallet, 3)
	for index := 0; index < 3; index++ {
		strMaxID := strconv.Itoa(palletMaxID)
		palletid := containerID + "PAL" + strMaxID
		pallets[index].PalletId = palletid
		pallets[index].Cases = createCase(palletid)
		palletMaxID++
	}
	return pallets
}

func incrementCounter(stub shim.ChaincodeStubInterface) error {
	ConMaxAsbytes, err := stub.GetState(UNIQUE_ID_COUNTER)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for ContainerMaxNumber \"}"
		return errors.New(jsonResp)
	}
	counter := UniqueIDCounter{}
	json.Unmarshal([]byte(ConMaxAsbytes), &counter)
	counter.ContainerMaxID = counter.ContainerMaxID + 1
	counter.PalletMaxID = counter.PalletMaxID + 3
	jsonVal, _ := json.Marshal(counter)
	err = stub.PutState(UNIQUE_ID_COUNTER, []byte(jsonVal))
	if err != nil {
		return err
	}
	return nil
}

func (t *PharmaChaincode) SetCurrentOwnerTest(stub shim.ChaincodeStubInterface, ownerID string, containerID string) ([]byte, error) {
	err := setCurrentOwner(stub, ownerID, containerID)
	return []byte("success"), err
}

func (t *PharmaChaincode) GetContainerDetailsForOwner(stub shim.ChaincodeStubInterface, ownerID string) ([]byte, error) {

	fmt.Println("Fetching container details for Owner:" + ownerID)

	ConMaxAsbytes, err := stub.GetState(CONTAINER_OWNER)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for Container Owners \"}"
		return nil, errors.New(jsonResp)
	}
	ConOwners := ContainerOwners{}
	json.Unmarshal([]byte(ConMaxAsbytes), &ConOwners)

	var containerList []string
	var matchFound bool

	for index := range ConOwners.Owners {
		if ConOwners.Owners[index].OwnerId == ownerID {
			containerList = ConOwners.Owners[index].ContainerList
			matchFound = true
			break
		}
	}
	if matchFound {
		fmt.Println("MatchFound for Owner:" + ownerID)
		shipment := Shipment{}
	
		for _, containerID := range containerList {
			byteVal, _ := t.GetContainerDetails(stub, containerID)
			container := Container{}

			json.Unmarshal([]byte(byteVal), &container)
			shipment.ContainerList = append(shipment.ContainerList, container)
		}
		jsonVal, _ := json.Marshal(shipment)
		return jsonVal, nil
	} else {
		fmt.Println("Container details not found for Owner:" + ownerID)
		return nil, errors.New("Unable to get container details for Owner:" + ownerID)
	}
}
func (t *PharmaChaincode) GetOwner(stub shim.ChaincodeStubInterface) ([]byte, error) {

	ConMaxAsbytes, err := stub.GetState(CONTAINER_OWNER)
	fmt.Println("************Am in GET OWNER Method**********")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for Container Owners \"}"
		return nil, errors.New(jsonResp)
	}
	return ConMaxAsbytes, nil
}
func (t *PharmaChaincode) AcceptContainerbyLogistics(stub shim.ChaincodeStubInterface,containerID string, logisticsID string, receiverID string, remarks string) ([]byte, error) {

	fmt.Println("Accepting the  container by Logistics:" + logisticsID)
	fmt.Println("Accepting the  container by Logistics:" + containerID)
     valAsbytes, err := stub.GetState(containerID)
	 if len(valAsbytes) == 0 {
		 	jsonResp := "{\"Error\":\"Failed to get state for Container id since there is no such container \"}"
		return nil, errors.New(jsonResp)
	 }
	 fmt.Println("json value from the container****************")
	 fmt.Println(valAsbytes)
	 if err != nil{
		jsonResp := "{\"Error\":\"Failed to get state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	//timeLayOut := timePresent.Format(RFC1123)
	  shipment := Container{}	  
	json.Unmarshal([]byte(valAsbytes), &shipment)
	shipment.Recipient = receiverID
	conprov := shipment.Provenance  
    supplychain := conprov.Supplychain     
	chainActivity := ChainActivity{
		Sender:   shipment.Provenance.Sender,
		Receiver: logisticsID,
		Status:   STATUS_ACCEPTED,
		ActivityTimeStamp:time.Now().UTC()}  
	supplychain = append(supplychain, chainActivity) 
	conprov.Supplychain = supplychain
   conprov.TransitStatus = STATUS_ACCEPTED
   conprov.Sender = shipment.Provenance.Sender
   conprov.Receiver = logisticsID
   shipment.Provenance = conprov
   jsonVal, _ := json.Marshal(shipment)
   	err = stub.PutState(containerID, jsonVal)
    if err != nil{
		jsonResp := "{\"Error\":\"Failed to put state for Container id \"}"
		return nil, errors.New(jsonResp)
	}	
	fmt.Println(string(jsonVal))
	fmt.Println(string(shipment.Provenance.Sender))
	setCurrentOwner(stub, logisticsID, containerID)
	return nil, nil		
}
func (t *PharmaChaincode) RejectContainerbyLogistics(stub shim.ChaincodeStubInterface,containerID string, logisticsID string, receiverID string, remarks string) ([]byte, error) {

	fmt.Println("Rejecting the  container by Logistics:" + logisticsID + containerID)
     valAsbytes, err := stub.GetState(containerID)
	 if len(valAsbytes) == 0 {
		 	jsonResp := "{\"Error\":\"Failed to get state for Container id since there is no such container \"}"
		return nil, errors.New(jsonResp)
	 }
	 fmt.Println("json value from the container****************")
	 fmt.Println(valAsbytes)
	 if err != nil{
		jsonResp := "{\"Error\":\"Failed to get state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println(remarks)
	if len(remarks) == 0 {
		 	jsonResp := "{\"Error\":\"Failed to have the remarks  for Container id since there is no input remarks \"}"
		return nil, errors.New(jsonResp)
	 }
	 shipment := Container{}	  
	json.Unmarshal([]byte(valAsbytes), &shipment)
	shipment.Recipient = receiverID
	
	conprov := shipment.Provenance  
    supplychain := conprov.Supplychain     
	chainActivity := ChainActivity{
		Sender:   shipment.Provenance.Sender,
		Receiver: logisticsID,
		Status:   STATUS_REJECTED,
		ActivityTimeStamp:time.Now().UTC()}  
	supplychain = append(supplychain, chainActivity) 
	conprov.Supplychain = supplychain
   conprov.TransitStatus = STATUS_REJECTED
   conprov.Sender = shipment.Provenance.Sender
   conprov.Receiver = logisticsID
   shipment.Provenance = conprov
   jsonVal, _ := json.Marshal(shipment)
   	err = stub.PutState(containerID, jsonVal)
    if err != nil{
		jsonResp := "{\"Error\":\"Failed to put state for Container id \"}"
		return nil, errors.New(jsonResp)
	}	
	fmt.Println(string(jsonVal))
	fmt.Println("SENDER",shipment.Provenance.Sender)
		setCurrentOwner(stub, logisticsID, containerID)
	return nil, nil		
}

func (t *PharmaChaincode) AcceptContainerbyDistributor(stub shim.ChaincodeStubInterface,containerID string, receiverID string, remarks string) ([]byte, error) {
    fmt.Println("Running AcceptContainerbyDistributor ")
	fmt.Println("Accepting the  container by Logistics:" + containerID)
     valAsbytes, err := stub.GetState(containerID)
	 if len(valAsbytes) == 0 {
		 	jsonResp := "{\"Error\":\"Failed to get state for Container id since there is no such container \"}"
		return nil, errors.New(jsonResp)
	 }
	 fmt.Println("json value from the container****************")
	 fmt.Println(valAsbytes)
	 if err != nil{
		jsonResp := "{\"Error\":\"Failed to get state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	  shipment := Container{}	  
	json.Unmarshal([]byte(valAsbytes), &shipment)
	shipment.Recipient = receiverID
	conprov := shipment.Provenance  
    supplychain := conprov.Supplychain     
	chainActivity := ChainActivity{
		Sender:   shipment.Provenance.Sender,
		Receiver: receiverID,
		Status:   STATUS_ACCEPTED,
		ActivityTimeStamp:time.Now().UTC()}  
	supplychain = append(supplychain, chainActivity) 
	conprov.Supplychain = supplychain
   conprov.TransitStatus = STATUS_ACCEPTED
   //taking sender from the container to avoid inconsistency of sender from UI
   conprov.Sender = shipment.Provenance.Sender
   conprov.Receiver = receiverID
   shipment.Provenance = conprov
   jsonVal, _ := json.Marshal(shipment)
   	err = stub.PutState(containerID, jsonVal)
    if err != nil{
		jsonResp := "{\"Error\":\"Failed to put state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("JSON ACCEPTED BY Reciever")	
		fmt.Println(string(jsonVal))
	setCurrentOwner(stub, receiverID, containerID)
	return nil, nil		
}

func (t *PharmaChaincode) RejectContainerbyDistributor(stub shim.ChaincodeStubInterface,containerID string, receiverID string, remarks string) ([]byte, error) {
    fmt.Println("Running RejectContainerbyDistributor ")
	fmt.Println("Accepting the  container by Logistics:" + containerID)
     valAsbytes, err := stub.GetState(containerID)
	 if len(valAsbytes) == 0 {
		 	jsonResp := "{\"Error\":\"Failed to get state for Container id since there is no such container \"}"
		return nil, errors.New(jsonResp)
	 }
	 fmt.Println("json value from the container****************")
	 fmt.Println(valAsbytes)
	 if err != nil{
		jsonResp := "{\"Error\":\"Failed to get state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	 fmt.Println(remarks)
	if len(remarks) == 0 {
		 	jsonResp := "{\"Error\":\"Failed to have the remarks  for Container id since there is no input remarks \"}"
		return nil, errors.New(jsonResp)
	 }
	  shipment := Container{}
	json.Unmarshal([]byte(valAsbytes), &shipment)
	shipment.Recipient = receiverID
	conprov := shipment.Provenance  
    supplychain := conprov.Supplychain     
	chainActivity := ChainActivity{
		Sender:   shipment.Provenance.Sender,
		Receiver: receiverID,
		Status:   STATUS_REJECTED,		 
		// ActivityTimeStamp=timeLayOut}
		ActivityTimeStamp:time.Now().UTC()}  
	supplychain = append(supplychain, chainActivity) 
	conprov.Supplychain = supplychain
   conprov.TransitStatus = STATUS_REJECTED
   //taking sender from the container to avoid inconsistency of sender from UI
   conprov.Sender = shipment.Provenance.Sender
   conprov.Receiver = receiverID
   shipment.Provenance = conprov
   jsonVal, _ := json.Marshal(shipment)
   	err = stub.PutState(containerID, jsonVal)
    if err != nil{
		jsonResp := "{\"Error\":\"Failed to put state for Container id \"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("JSON ACCEPTED BY Reciever")	
		fmt.Println(string(jsonVal))
	setCurrentOwner(stub, receiverID, containerID)
	return nil, nil		
}

func (t *PharmaChaincode) GetUserAttribute(stub shim.ChaincodeStubInterface, attributeName string) ([]byte, error) {
	fmt.Println("***** Inside GetUserAttribute() func for attribute:" + attributeName)
	attributeValue, err := stub.ReadCertAttribute(attributeName)
	fmt.Println("attributeValue=" + string(attributeValue))
	
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get GetUserAttribute\"}"
		return nil, errors.New(jsonResp)
	}
	return attributeValue, nil
}

func setCurrentOwner(stub shim.ChaincodeStubInterface, ownerID string, containerID string) error {
	ConMaxAsbytes, err := stub.GetState(CONTAINER_OWNER)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for ContainerMaxNumber \"}"
		return errors.New(jsonResp)
	}
	ConOwners := ContainerOwners{}
	json.Unmarshal([]byte(ConMaxAsbytes), &ConOwners)

	var containerList []string
	var ownerIndex int
	var matchFound bool
	for index := range ConOwners.Owners {
		if ConOwners.Owners[index].OwnerId == ownerID {
			ownerIndex = index
			containerList = ConOwners.Owners[index].ContainerList
			matchFound = true
			break
		}
	}
	containerFound := false
	if matchFound {
		for index := range containerList {
			if containerList[index] == containerID {
				containerFound = true
				break
			}
		}
		if !containerFound {
			containerList = append(containerList, containerID)
			ConOwners.Owners[ownerIndex].ContainerList = containerList
		}
	} else {
		containerList := make([]string, 1)
		containerList[0] = containerID
		owner := Owner{OwnerId: ownerID, ContainerList: containerList}
		ConOwners.Owners = append(ConOwners.Owners, owner)
	}

	jsonVal, _ := json.Marshal(ConOwners)
	err = stub.PutState(CONTAINER_OWNER, []byte(jsonVal))
	if err != nil {
		return err
	}

	return nil
}