package server

import (
	"bytes"
	"encoding/binary"
	"image"
	"log"
	"net"
	"os"
	"path/filepath"

	"color-blind-simulator-1/app/utils"

	"github.com/disintegration/imaging"
)

const (
	UDPPort = ":8081"
	MaxSize = 65507 // Maximum UDP packet size
)

// Operation types
const (
	OpFlip          = 1
	OpRotate        = 2
	OpRotateShear   = 3
	OpGrayscale     = 4
	OpBoxBlur       = 5
	OpGaussianBlur  = 6
	OpEdgeDetection = 7
	OpProtanopia    = 8
	OpDeuteranopia  = 9
	OpTritanopia    = 10
	OpProtanomaly   = 11
	OpDeuteranomaly = 12
	OpTritanomaly   = 13
	OpAchromatopsia = 14
	OpMonochromacy  = 15
	OpDaltonize     = 16
)

// StartUDPServer starts a UDP server for image processing
func StartUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", UDPPort)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Printf("UDP Server listening on %s", UDPPort)

	buffer := make([]byte, MaxSize)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		// Process the received data
		go handleUDPPacket(conn, remoteAddr, buffer[:n])
	}
}

func handleUDPPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// First 4 bytes are the operation type
	if len(data) < 4 {
		return
	}

	opType := binary.BigEndian.Uint32(data[:4])
	imageData := data[4:]

	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		return
	}

	// Decode the image
	src, err := imaging.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		return
	}

	// Process the image based on operation type
	var processedImage image.Image
	switch opType {
	case OpFlip:
		processedImage = utils.FlipImage(src)
	case OpRotate:
		processedImage = utils.RotateImage(src, 45) // Default 45-degree rotation
	case OpRotateShear:
		processedImage = utils.RotateImageWithShear(src, 45)
	case OpGrayscale:
		processedImage = utils.ConvertToGrayscale(src)
	case OpBoxBlur:
		processedImage = utils.ApplyBoxBlur(src)
	case OpGaussianBlur:
		processedImage = utils.ApplyGaussianBlur(src)
	case OpEdgeDetection:
		processedImage = utils.ApplyEdgeDetection(src)
	case OpProtanopia:
		processedImage = utils.SimulateColorBlindness(src, utils.ProtanopiaMatrix)
	case OpDeuteranopia:
		processedImage = utils.SimulateColorBlindness(src, utils.DeuteranopiaMatrix)
	case OpTritanopia:
		processedImage = utils.SimulateColorBlindness(src, utils.TritanopiaMatrix)
	case OpProtanomaly:
		processedImage = utils.SimulateColorBlindness(src, utils.ProtanomalyMatrix)
	case OpDeuteranomaly:
		processedImage = utils.SimulateColorBlindness(src, utils.DeuteranomalyMatrix)
	case OpTritanomaly:
		processedImage = utils.SimulateColorBlindness(src, utils.TritanomalyMatrix)
	case OpAchromatopsia:
		processedImage = utils.SimulateColorBlindness(src, utils.AchromatopsiaMatrix)
	case OpMonochromacy:
		processedImage = utils.SimulateColorBlindness(src, utils.MonochromacyMatrix)
	case OpDaltonize:
		processedImage = utils.Daltonize(src, utils.ProtanopiaMatrix)
	default:
		processedImage = src
	}

	// Save the processed image
	filename := filepath.Join(outputDir, "udp_processed.jpg")
	if err := imaging.Save(processedImage, filename); err != nil {
		log.Printf("Error saving processed image: %v", err)
		return
	}

	// Send acknowledgment
	response := []byte("Image processed successfully")
	conn.WriteToUDP(response, addr)
}
