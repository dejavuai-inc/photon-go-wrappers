package main


import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
        p "github.com/dejavuai-inc/photon-go-wrappers/es"
)

var ctx unsafe.Pointer

func main() {
	// Arguments
	// 1. Command-line argument parsing
	//if len(os.Args) < 5 { // Check if enough arguments are provided
	//	fmt.Println("Usage: program_name --server [IP or localhost:port] --image [image name]")
	//	return
	//}

	//serverArg := os.Args[1]
	//serverValue := os.Args[2]
	//imageArg := os.Args[3]
	//imageValue := os.Args[4]

	serverValue := "localhost:8100"
	imageValue := "/home/serge/Images/Nature/Legumes/P0302068.JPG"

	//if serverArg != "--server" || imageArg != "--image" {
	//	fmt.Println("Invalid argument names. Use --server and --image.")
	//	return
	//}

	// 2. Server argument validation
	serverParts := strings.Split(serverValue, ":")
	if len(serverParts) != 2 {
		fmt.Println("Invalid server format. Use XXX.XXX.XXX.XXX:YYYY or localhost:YYYY")
		return
	}

	serverIP := serverParts[0]
	serverPort, err := strconv.Atoi(serverParts[1])
	if err != nil || serverPort < 1 || serverPort > 65000 || (serverIP != "localhost" && !p.IsValidIP(serverIP)) {
		fmt.Println("Invalid server IP or port.")
		return
	}

	// 3. Image file existence check
	if _, err := os.Stat(imageValue); os.IsNotExist(err) {
		fmt.Println("Image file does not exist:", imageValue)
		return
	}

	//return

	// Init ************************************
	errgoint := p.Init()
	if errgoint != 0 {
		fmt.Println("ERROR : C.Init() ")
		p.Uninit()
		return
	}

	// Connect ************************************
	addr := "tcp://" + serverValue
	ctx = p.Connect(addr)
	fmt.Println("ctx : ", ctx)
	if ctx == nil {
		fmt.Println("ERROR : Could not connect to server!")
		p.Uninit()
		return
	}

	goStrInfo, _ := p.GetInfo(ctx)

	str := p.StringFromStrInfo(&goStrInfo)

	fmt.Println("str : ", str)

	//  ************************************

	var respGetInfo p.GetInfoResponses
	err1 := json.Unmarshal([]byte(str), &respGetInfo) // Deserialize JSON
	if err1 != nil {
		fmt.Println("Error deserializing JSON (1):", err)
		return
	}
	// Now you can access the data like this:
	fmt.Println("Overall Error:", respGetInfo.Error)

	for _, response := range respGetInfo.Responses {
		fmt.Println("Session GUID:", response.SessionGUID)
		fmt.Println("Database GUID:", response.DbGUID)
		fmt.Println("File Name:", response.FileName)
		fmt.Println("Color Mode:", response.ColorMode)
		fmt.Println("Resolution:", response.Resolution)
		fmt.Println("Tile Size:", response.TileSize)
		fmt.Println("Image Count:", response.ImageCount)
		fmt.Println("Max Image Count:", response.MaxImageCount)
		fmt.Println("Extra Info:", response.ExtraInfo)
		// ... and so on ...
	}

	databaseOpened := false

	for _, resp := range respGetInfo.Responses {
		if resp.DbGUID != "" { // Check if db_guid is not empty in Go
			databaseOpened = true
			break
		}
	}

	if !databaseOpened {
		CorrectExit(ctx, "No database is opened. Open a database first.")
		return
	}

	fmt.Println("Searching in all opened databases")

	rotStr := "1" // Default rotation setting

	var rotation int16 = p.SEARCH_ROTATION_NONE
	if rotStr == "1" {
		rotation = p.SEARCH_ROTATION_MUL_90
	} else if rotStr == "2" {
		rotation = p.SEARCH_ROTATION_ARB
	}

	resolution := 0 // Use database's resolution by default
	sensitivity := p.SEARCH_SENSITIVITY_NORM
	flags := p.SEARCH_GET_IMAGE_INFO // Option to return image info

	var imgData []byte
	var imageDataSize int64 = 0

	imgData, imageDataSize, err = LoadImage(imageValue)

	if err != nil {
		CorrectExit(ctx, "ERROR : LoadImage(nil, nil)")
		return
	}

	// Search
	fmt.Println(ctx)
	fmt.Println(imageValue)
	//fmt.Println("imageData : ", imageData)
	fmt.Println(imageDataSize)
	fmt.Println(resolution)
	fmt.Println(rotation)
	fmt.Println(sensitivity)
	fmt.Println(flags)
	//fmt.Println(cStrInfo)

	strInfo := p.Search(ctx, "", imgData, imageDataSize, resolution, rotation, sensitivity, flags)

	fmt.Println(strInfo)

	str = p.StringFromStrInfo(&strInfo)

	fmt.Println("Search str : ", str)

	//  ************************************

	var resp_Search p.SearchResponses
	err1 = json.Unmarshal([]byte(str), &resp_Search) // Deserialize JSON
	if err1 != nil {
		fmt.Println("Error deserializing JSON (2), resp_Search:", err1)
		return
	}

	// Iterate through search results
	for _, searchResponse := range resp_Search.Responses {
		sessionID := searchResponse.SessionGUID

		// Iterate through image IDs and collect unique match IDs
		numResults := len(searchResponse.Results)
		fmt.Printf("Number of search results: %d\n", numResults)

		if numResults == 0 {
			continue
		}

		uniqueMatchIDs := make(map[uint]bool) // Use a map for uniqueness
		for _, searchRes := range searchResponse.Results {
			uniqueMatchIDs[searchRes.MatchID] = true
		}

		// Request thumbnails
		for matchID := range uniqueMatchIDs {
			var thumbData []byte

			goStrInfo, thumbData = p.GetThumbnail(ctx, sessionID, matchID, "JPEG")

			str = p.StringFromStrInfo(&goStrInfo)

			var respGetThumb p.GetThumbnailResponse
			err := json.Unmarshal([]byte(str), &respGetThumb)
			if err != nil {
				fmt.Println("Error deserializing GetThumbnail response:", err)
				continue
			}

			if respGetThumb.Error != "OK" {
				fmt.Printf("GetThumbnail(): error=%s\n", respGetThumb.Error)
				continue
			}

			fmt.Printf("Received a thumbnail for matchID %d\n", matchID)

			thumbFileName := fmt.Sprintf("%s_%d.jpg", sessionID, matchID)
			fmt.Printf("Saving thumbnail: %s...\n", thumbFileName)
			if err := saveDataToFile(thumbFileName, thumbData); err != nil {
				fmt.Println("Error saving thumbnail:", err)
				continue
			}

			fmt.Println("Done!")
		}
	}

	// Disconnect  and  Uninit *********************
	CorrectExit(ctx, "The End")

}

func CorrectExit(ctx unsafe.Pointer, msg string) {
	fmt.Println(msg)
	p.Disconnect(ctx)
	p.Uninit()
}

func saveDataToFile(fileName string, data []byte) error {
	// Create or truncate the file (equivalent to FileMode.OpenWrite in C#)
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Write the data to the file
	bytesWritten, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	// Verify that all bytes were written
	if int64(bytesWritten) != int64(len(data)) {
		return fmt.Errorf("incomplete write: only %d bytes written", bytesWritten)
	}

	return nil // Success! No error to return
}

func LoadImage(fileName string) ([]byte, int64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, 0, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting file info: %v", err)
	}

	fileSize := fileInfo.Size() // Get file size in bytes

	data := make([]byte, fileSize)
	bytesRead, err := file.Read(data)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading file: %v", err)
	}
	if int64(bytesRead) != fileSize {
		return nil, 0, fmt.Errorf("incomplete read: only %d bytes read", bytesRead)
	}

	return data, fileSize, nil // Return data, size, and nil error
}
