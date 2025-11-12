package badge_printer_zebra

import (
	"fmt"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Encoder handles magnetic stripe and RFID encoding
type Encoder struct {
	logger *telemetry.Logger
}

// NewEncoder creates a new encoder
func NewEncoder(logger *telemetry.Logger) *Encoder {
	return &Encoder{
		logger: logger,
	}
}

// Encode encodes magnetic stripe or RFID data to the card
func (e *Encoder) Encode(conn Connection, encoding *pb.EncodingData) error {
	if encoding == nil {
		return nil
	}

	// Encode magnetic stripe
	if encoding.MagneticStripe != nil && encoding.MagneticStripe.Enabled {
		if err := e.encodeMagneticStripe(conn, encoding.MagneticStripe); err != nil {
			return fmt.Errorf("magnetic stripe encoding failed: %w", err)
		}
	}

	// Encode RFID
	if encoding.Rfid != nil && encoding.Rfid.Enabled {
		if err := e.encodeRFID(conn, encoding.Rfid); err != nil {
			return fmt.Errorf("RFID encoding failed: %w", err)
		}
	}

	return nil
}

// encodeMagneticStripe encodes magnetic stripe data
func (e *Encoder) encodeMagneticStripe(conn Connection, data *pb.MagneticStripeData) error {
	e.logger.Debug("encoding magnetic stripe",
		telemetry.Bool("track1", data.Track1Data != ""),
		telemetry.Bool("track2", data.Track2Data != ""),
		telemetry.Bool("track3", data.Track3Data != ""),
	)

	var zpl string

	// Start encoding command
	zpl += "^XA\n"

	// Set coercivity (HiCo or LoCo)
	if data.Coercivity == pb.MagneticStripeData_COERCIVITY_TYPE_HICO {
		zpl += "^MH\n" // High coercivity (2750 Oe)
	} else {
		zpl += "^ML\n" // Low coercivity (300 Oe)
	}

	// Encode Track 1 (alphanumeric, up to 79 characters)
	if data.Track1Data != "" {
		// ^M1 - Encode track 1
		zpl += fmt.Sprintf("^M1%s\n", data.Track1Data)
	}

	// Encode Track 2 (numeric, up to 40 characters)
	if data.Track2Data != "" {
		// ^M2 - Encode track 2
		zpl += fmt.Sprintf("^M2%s\n", data.Track2Data)
	}

	// Encode Track 3 (numeric, up to 107 characters)
	if data.Track3Data != "" {
		// ^M3 - Encode track 3
		zpl += fmt.Sprintf("^M3%s\n", data.Track3Data)
	}

	// End command
	zpl += "^XZ\n"

	// Send to printer
	_, err := conn.Write([]byte(zpl))
	if err != nil {
		return fmt.Errorf("failed to send magnetic stripe encoding: %w", err)
	}

	e.logger.Info("magnetic stripe encoded successfully")
	return nil
}

// encodeRFID encodes RFID chip data
func (e *Encoder) encodeRFID(conn Connection, data *pb.RFIDData) error {
	e.logger.Debug("encoding RFID",
		telemetry.String("type", data.Type.String()),
		telemetry.Int("blocks", len(data.Blocks)),
	)

	var zpl string

	// Start encoding command
	zpl += "^XA\n"

	// Set RFID tag type
	switch data.Type {
	case pb.RFIDData_RFID_TYPE_MIFARE_CLASSIC_1K:
		zpl += "^RS,MIFARE_1K\n"
	case pb.RFIDData_RFID_TYPE_MIFARE_CLASSIC_4K:
		zpl += "^RS,MIFARE_4K\n"
	case pb.RFIDData_RFID_TYPE_MIFARE_ULTRALIGHT:
		zpl += "^RS,MIFARE_UL\n"
	case pb.RFIDData_RFID_TYPE_MIFARE_DESFIRE:
		zpl += "^RS,DESFIRE\n"
	case pb.RFIDData_RFID_TYPE_HID_ICLASS:
		zpl += "^RS,ICLASS\n"
	case pb.RFIDData_RFID_TYPE_HID_PROX:
		zpl += "^RS,PROX\n"
	default:
		return fmt.Errorf("unsupported RFID type: %v", data.Type)
	}

	// Write UID if provided (for programmable tags)
	if len(data.Uid) > 0 {
		uidHex := fmt.Sprintf("%X", data.Uid)
		zpl += fmt.Sprintf("^RW,UID,%s\n", uidHex)
	}

	// Write data blocks
	for _, block := range data.Blocks {
		if len(block.Data) == 0 {
			continue
		}

		// Convert block data to hex
		dataHex := fmt.Sprintf("%X", block.Data)

		// ^RW command: write RFID data
		// Format: ^RW,BLOCK,block_number,data
		zpl += fmt.Sprintf("^RW,BLOCK,%d,%s\n", block.BlockNumber, dataHex)
	}

	// Set access key if provided (for secured tags)
	if len(data.AccessKey) > 0 {
		keyHex := fmt.Sprintf("%X", data.AccessKey)
		zpl += fmt.Sprintf("^RW,KEY,%s\n", keyHex)
	}

	// Verify RFID programming (optional)
	zpl += "^RV,ENABLE\n" // Enable verification

	// End command
	zpl += "^XZ\n"

	// Send to printer
	_, err := conn.Write([]byte(zpl))
	if err != nil {
		return fmt.Errorf("failed to send RFID encoding: %w", err)
	}

	e.logger.Info("RFID encoded successfully",
		telemetry.Int("blocks_written", len(data.Blocks)),
	)
	return nil
}
