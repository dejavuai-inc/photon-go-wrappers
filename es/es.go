package es

// #cgo CFLAGS: -I. -I/usr/include/Photon
// #cgo LDFLAGS: -L. -L/usr/lib/Photon -lPhotonAPI2 -Wl,-rpath,./ -Wl,-rpath,/usr/lib/Photon
// #include "PhotonAPI2.h"
import "C"

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

const (
	SEARCH_MODE_GRAY = 0
	SEARCH_MODE_RGB  = 1 // not supported yet
	SEARCH_MODE_BOTH = 2 // not supported yet

	// Rotation
	SEARCH_ROTATION_NONE   = 0
	SEARCH_ROTATION_MUL_90 = 1
	SEARCH_ROTATION_ARB    = 2

	// Sensitivity
	SEARCH_SENSITIVITY_NORM = 0
	SEARCH_SENSITIVITY_HIGH = 1

	// Search options/flags
	SEARCH_MIRRORED       = 1
	SEARCH_THOROUGH       = 2
	SEARCH_GET_IMAGE_INFO = 4

	// Search staging enumerator
	SEARCH_SINGLE_STAGE           = 0
	SEARCH_DOUBLE_STAGE_BASIC     = 1
	SEARCH_DOUBLE_STAGE_OPTIMIZED = 2
	SEARCH_DOUBLE_STAGE_DEFAULT   = SEARCH_DOUBLE_STAGE_BASIC

	// Resolution levels
	MIN_RESOLUTION_LEVEL     = 2
	MAX_RESOLUTION_LEVEL     = 12
	DEFAULT_RESOLUTION_LEVEL = 4
)

type GetThumbnailResponse struct {
	Response
	Width    uint `json:"width"`
	Height   uint `json:"height"`
	DataSize uint `json:"data_size"`
}

type Response struct {
	Error     string `json:"error"` // Move "error" field to Response
	ErrorInfo string `json:"error_info"`
}

type GetInfoResponse struct {
	SessionGUID   string `json:"session_guid"` // Add SessionGUID here
	DbGUID        string `json:"db_guid"`
	FileName      string `json:"file_name"`
	ColorMode     uint   `json:"color_mode"`
	Resolution    uint   `json:"resolution"`
	TileSize      uint   `json:"tile_size"`
	ImageCount    uint   `json:"image_count"`
	MaxImageCount uint   `json:"max_image_count"`
	ExtraInfo     string `json:"extra_info"` // Add ExtraInfo here
}

type FRect struct {
	MinX float32 `json:"minX"`
	MaxX float32 `json:"maxX"`
	MinY float32 `json:"minY"`
	MaxY float32 `json:"maxY"`
}

type SearchResult struct {
	ImageInfo  string  `json:"image_info"`
	MatchID    uint    `json:"match_id"`
	Scale      float64 `json:"scale"`
	Angle      float64 `json:"angle"`
	Mirrored   bool    `json:"mirrored"`
	Dxs        float64 `json:"dxs"`
	Dys        float64 `json:"dys"`
	Dxm        float64 `json:"dxm"`
	Dym        float64 `json:"dym"`
	Xform      int     `json:"xform"`
	Confidence float64 `json:"confidence"`
	Sr         FRect   `json:"sr"`
	Mr         FRect   `json:"mr"`
}

type SearchResponse struct {
	Response
	SessionGUID string         `json:"session_guid"` // Add SessionGUID here
	SearchTime  uint           `json:"search_time"`
	Results     []SearchResult `json:"results"`
}

type SearchResponses struct {
	Response
	Responses []SearchResponse `json:"responses"`
}

type GetInfoResponses struct {
	Error     string            `json:"error"` // Error field at the top level
	Responses []GetInfoResponse `json:"responses"`
}

type PhotonAPI2_StrInfo struct {
	Str    *C.char
	Length C.uint
}

type PhotonAPI2_BufInfo struct {
	data     *C.uchar
	dataSize C.uint
}

//export statusCallback
func statusCallback(status C.int, param unsafe.Pointer) {
	if status == C.SERVER_HEARTBEAT_LOST {
		fmt.Printf("\nStatus callback called: SERVER HEARTBEAT LOST\n")
		//C.Disconnect(ctx)
		os.Exit(1)
	} else {
		fmt.Printf("\nStatus callback called: SERVER HEARTBEAT IS OK\n")
	}
}

/*******************************************/
// Go Wrappers
/******************************************/

func Init() int {
	return int(C.Init())
}

func Uninit() int {
	return int(C.Uninit())
}

func Disconnect(ctx unsafe.Pointer) {
	C.Disconnect(ctx)
}

func Connect(address string) unsafe.Pointer {
	cAddress := C.CString(address)
	defer C.free(unsafe.Pointer(cAddress))
	return unsafe.Pointer(C.Connect(cAddress, C.PFPhotonAPI2_StatusCallback(C.statusCallback), nil))
}

func GetInfo(ctx unsafe.Pointer) (PhotonAPI2_StrInfo, error) {
	var info PhotonAPI2_StrInfo
	cInfo := (*C.struct_PhotonAPI2_StrInfo)(unsafe.Pointer(&info))
	errint := C.GetInfo(unsafe.Pointer(ctx), cInfo)
	if errint != 0 {
		return info, fmt.Errorf("GetInfo failed with error code: %d", errint)
	}
	return info, nil
}

func Search(ctx unsafe.Pointer, fileName string, imageData []byte, imageDataSize int64, resolution int, rotation int16, sensitivity int, flags int) PhotonAPI2_StrInfo {
	var info PhotonAPI2_StrInfo
	cInfo := (*C.struct_PhotonAPI2_StrInfo)(unsafe.Pointer(&info))

	cImageData := (*C.uchar)(C.CBytes(imageData))
	defer C.free(unsafe.Pointer(cImageData))

	errint := C.Search(unsafe.Pointer(ctx), C.CString(fileName), cImageData, C.uint(imageDataSize), C.uint(resolution), C.ushort(rotation), C.uint(sensitivity), C.uint(flags), cInfo)

	if errint != 0 {
		fmt.Printf("Search failed with error code: %d", errint)
		return info
	}
	return info
}

func GetThumbnail(ctx unsafe.Pointer, sessionGUID string, imageID uint, encoding string) (PhotonAPI2_StrInfo, []byte) {
	var strInfo PhotonAPI2_StrInfo
	var bufInfo PhotonAPI2_BufInfo

	cStrInfo := (*C.struct_PhotonAPI2_StrInfo)(unsafe.Pointer(&strInfo))
	cBufInfo := (*C.struct_PhotonAPI2_BufInfo)(unsafe.Pointer(&bufInfo))

	errint := C.GetThumbnail(unsafe.Pointer(ctx), C.CString(sessionGUID), C.uint(imageID), C.CString(encoding), cStrInfo, cBufInfo)
	if errint != 0 {
		fmt.Printf("GetThumbnail failed with error code: %d", errint)
		return strInfo, nil
	}

	// Convert PhotonAPI2_BufInfo to []byte
	imgData := BytesFromBufInfo(&bufInfo)

	return strInfo, imgData // Return strInfo, imgData as []byte
}

/*******************************************/
// Other functions
/******************************************/

// BytesFromBufInfo converts PhotonAPI2_BufInfo to a Go byte slice
func BytesFromBufInfo(bufInfo *PhotonAPI2_BufInfo) []byte {
	// Ensure the buffer has data
	if bufInfo.data == nil || bufInfo.dataSize == 0 {
		return nil
	}

	return (*[1 << 30]byte)(unsafe.Pointer(bufInfo.data))[:bufInfo.dataSize:bufInfo.dataSize]
}

func StringFromStrInfo(strInfo *PhotonAPI2_StrInfo) string {
	length := int(strInfo.Length) // Convert C.uint to int

	// Check if length is greater than 0
	if length > 0 {
		// Create a slice that points to the same memory as strInfo.Str,
		// but with the length specified
		strSlice := (*[1 << 30]byte)(unsafe.Pointer(strInfo.Str))[:length:length]

		// Check if the last byte in the slice is a null terminator
		if strSlice[length-1] == 0 {
			length--
		}
	}

	// Convert to a Go string without the null byte if present
	return C.GoStringN(strInfo.Str, C.int(length))
}

func IsValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return false
		}
	}

	return true
}
