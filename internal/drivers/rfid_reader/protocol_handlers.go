package rfid_reader

import (
	"encoding/hex"
	"fmt"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// ==================================================
// HID Prox Handler (125 kHz)
// ==================================================

type proxHandler struct {
	logger *telemetry.Logger
}

// NewProxHandler creates a new HID Prox card handler
func NewProxHandler(logger *telemetry.Logger) CardProtocolHandler {
	return &proxHandler{logger: logger}
}

func (h *proxHandler) SupportsCard(uid []byte, atr []byte) bool {
	// Prox cards are 125 kHz, typically have 4-8 byte UIDs
	return len(uid) >= 4 && len(uid) <= 8
}

func (h *proxHandler) ReadCard(conn Connection, uid []byte, options *pb.ReadOptions) (*pb.CardData, error) {
	h.logger.Debug("reading HID Prox card",
		telemetry.String("uid", hex.EncodeToString(uid)),
	)

	// Parse Wiegand data from UID
	// Standard 26-bit Wiegand format:
	// - 8 bits facility code
	// - 16 bits card number
	// - 2 parity bits

	var facilityCode, cardNumber int32
	var wiegandBits int32 = 26

	if len(uid) >= 3 {
		// Extract facility code and card number
		// This is a simplified extraction - real implementation depends on format
		facilityCode = int32(uid[0])
		cardNumber = int32(uint16(uid[1])<<8 | uint16(uid[2]))
	}

	cardData := &pb.CardData{
		Type:         pb.CardType_CARD_TYPE_HID_PROX,
		Uid:          uid,
		SerialNumber: fmt.Sprintf("%d:%d", facilityCode, cardNumber),
		Technology:   pb.CardTechnology_CARD_TECHNOLOGY_PROX_125KHZ,
		ProtocolData: &pb.CardData_Prox{
			Prox: &pb.ProxCardData{
				FacilityCode: facilityCode,
				CardNumber:   cardNumber,
				WiegandBits:  wiegandBits,
				WiegandData:  uid,
				FormatName:   "26-bit Wiegand",
			},
		},
		Capabilities: &pb.CardCapabilities{
			Writable:         false,
			RequiresAuth:     false,
			SupportsNdef:     false,
			ReadOnly:         true,
		},
	}

	return cardData, nil
}

func (h *proxHandler) WriteCard(conn Connection, uid []byte, blocks []*pb.CardBlock, credentials *pb.AuthenticationCredentials) error {
	return fmt.Errorf("Prox cards are read-only")
}

func (h *proxHandler) AuthenticateCard(conn Connection, uid []byte, credentials *pb.AuthenticationCredentials) (bool, error) {
	// Prox cards don't support authentication
	return true, nil
}

func (h *proxHandler) GetCardType() pb.CardType {
	return pb.CardType_CARD_TYPE_HID_PROX
}

func (h *proxHandler) GetProtocolName() string {
	return "HID Prox (125 kHz)"
}

// ==================================================
// Mifare Classic Handler
// ==================================================

type mifareHandler struct {
	logger *telemetry.Logger
}

// NewMifareHandler creates a new Mifare Classic handler
func NewMifareHandler(logger *telemetry.Logger) CardProtocolHandler {
	return &mifareHandler{logger: logger}
}

func (h *mifareHandler) SupportsCard(uid []byte, atr []byte) bool {
	// Mifare cards have 4 or 7 byte UIDs
	return len(uid) == 4 || len(uid) == 7
}

func (h *mifareHandler) ReadCard(conn Connection, uid []byte, options *pb.ReadOptions) (*pb.CardData, error) {
	h.logger.Debug("reading Mifare Classic card",
		telemetry.String("uid", hex.EncodeToString(uid)),
	)

	// Determine card size (1K or 4K based on SAK)
	// For now, assume 1K (16 sectors, 4 blocks each = 64 blocks)
	numSectors := 16
	sizeBytes := 1024

	cardData := &pb.CardData{
		Type:         pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K,
		Uid:          uid,
		SerialNumber: hex.EncodeToString(uid),
		Technology:   pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ,
		ProtocolData: &pb.CardData_Mifare{
			Mifare: &pb.MifareCardData{
				SizeBytes:  int32(sizeBytes),
				NumSectors: int32(numSectors),
				Sak:        0x08, // Typical for Mifare Classic 1K
				Atqa:       0x0004,
			},
		},
		Capabilities: &pb.CardCapabilities{
			Writable:         true,
			RequiresAuth:     true,
			SupportsEncryption: true,
			SupportsNdef:     false,
			MaxDataSize:      int32(sizeBytes),
		},
	}

	// If full read requested and authentication provided, read sectors
	if options != nil && options.ReadFullData && len(options.AuthKeys) > 0 {
		sectors := make([]*pb.MifareSector, 0, numSectors)

		for sectorNum := 0; sectorNum < numSectors; sectorNum++ {
			// Authenticate and read sector
			// This is a placeholder - real implementation would:
			// 1. Send authentication command with key
			// 2. Read blocks in sector
			// 3. Parse access bits from sector trailer

			sector := &pb.MifareSector{
				SectorNumber: int32(sectorNum),
				Blocks:       make([]*pb.CardBlock, 0, 4),
			}

			// Each sector has 4 blocks (except last sector of 4K cards)
			for blockNum := 0; blockNum < 4; blockNum++ {
				// Placeholder block data
				block := &pb.CardBlock{
					BlockNumber: int32(sectorNum*4 + blockNum),
					Data:        make([]byte, 16),
					Access:      "read-only",
				}
				sector.Blocks = append(sector.Blocks, block)
			}

			sectors = append(sectors, sector)
		}

		cardData.GetMifare().Sectors = sectors
	}

	return cardData, nil
}

func (h *mifareHandler) WriteCard(conn Connection, uid []byte, blocks []*pb.CardBlock, credentials *pb.AuthenticationCredentials) error {
	h.logger.Debug("writing Mifare Classic card",
		telemetry.String("uid", hex.EncodeToString(uid)),
		telemetry.Int("blocks", len(blocks)),
	)

	// Real implementation would:
	// 1. Authenticate to each sector
	// 2. Write blocks
	// 3. Verify writes

	return fmt.Errorf("Mifare writing not yet fully implemented")
}

func (h *mifareHandler) AuthenticateCard(conn Connection, uid []byte, credentials *pb.AuthenticationCredentials) (bool, error) {
	if credentials == nil || len(credentials.MifareKeysA) == 0 {
		return false, fmt.Errorf("no Mifare keys provided")
	}

	// Try authentication with provided keys
	// Real implementation would send authentication commands to reader
	// For now, return success if keys are provided

	h.logger.Debug("authenticating Mifare card",
		telemetry.String("uid", hex.EncodeToString(uid)),
		telemetry.Int("keys", len(credentials.MifareKeysA)),
	)

	return true, nil
}

func (h *mifareHandler) GetCardType() pb.CardType {
	return pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K
}

func (h *mifareHandler) GetProtocolName() string {
	return "Mifare Classic"
}

// ==================================================
// HID iClass Handler
// ==================================================

type iclassHandler struct {
	logger *telemetry.Logger
}

// NewIClassHandler creates a new iClass handler
func NewIClassHandler(logger *telemetry.Logger) CardProtocolHandler {
	return &iclassHandler{logger: logger}
}

func (h *iclassHandler) SupportsCard(uid []byte, atr []byte) bool {
	// iClass cards have 8-byte UIDs (CSN)
	return len(uid) == 8
}

func (h *iclassHandler) ReadCard(conn Connection, uid []byte, options *pb.ReadOptions) (*pb.CardData, error) {
	h.logger.Debug("reading HID iClass card",
		telemetry.String("uid", hex.EncodeToString(uid)),
	)

	cardData := &pb.CardData{
		Type:         pb.CardType_CARD_TYPE_HID_ICLASS,
		Uid:          uid,
		SerialNumber: hex.EncodeToString(uid),
		Technology:   pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ,
		ProtocolData: &pb.CardData_Iclass{
			Iclass: &pb.IClassCardData{
				Applications: make([]*pb.IClassApplication, 0),
			},
		},
		Capabilities: &pb.CardCapabilities{
			Writable:         true,
			RequiresAuth:     true,
			SupportsEncryption: true,
			SupportsNdef:     false,
			MaxDataSize:      2048, // Typical iClass card
		},
	}

	return cardData, nil
}

func (h *iclassHandler) WriteCard(conn Connection, uid []byte, blocks []*pb.CardBlock, credentials *pb.AuthenticationCredentials) error {
	return fmt.Errorf("iClass writing not yet fully implemented")
}

func (h *iclassHandler) AuthenticateCard(conn Connection, uid []byte, credentials *pb.AuthenticationCredentials) (bool, error) {
	if credentials == nil || len(credentials.IclassKey) == 0 {
		return false, fmt.Errorf("no iClass key provided")
	}

	h.logger.Debug("authenticating iClass card",
		telemetry.String("uid", hex.EncodeToString(uid)),
	)

	return true, nil
}

func (h *iclassHandler) GetCardType() pb.CardType {
	return pb.CardType_CARD_TYPE_HID_ICLASS
}

func (h *iclassHandler) GetProtocolName() string {
	return "HID iClass"
}

// ==================================================
// NTAG Handler (NFC Tags)
// ==================================================

type ntagHandler struct {
	logger *telemetry.Logger
}

// NewNTAGHandler creates a new NTAG handler
func NewNTAGHandler(logger *telemetry.Logger) CardProtocolHandler {
	return &ntagHandler{logger: logger}
}

func (h *ntagHandler) SupportsCard(uid []byte, atr []byte) bool {
	// NTAG cards have 7-byte UIDs
	return len(uid) == 7
}

func (h *ntagHandler) ReadCard(conn Connection, uid []byte, options *pb.ReadOptions) (*pb.CardData, error) {
	h.logger.Debug("reading NTAG card",
		telemetry.String("uid", hex.EncodeToString(uid)),
	)

	// Determine NTAG type based on memory size
	// NTAG213: 180 bytes user memory
	// NTAG215: 504 bytes user memory
	// NTAG216: 888 bytes user memory

	ntagType := "NTAG213"
	totalMemory := 180
	userMemory := 144

	cardData := &pb.CardData{
		Type:         pb.CardType_CARD_TYPE_NTAG_213,
		Uid:          uid,
		SerialNumber: hex.EncodeToString(uid),
		Technology:   pb.CardTechnology_CARD_TECHNOLOGY_NFC,
		ProtocolData: &pb.CardData_Ntag{
			Ntag: &pb.NTAGCardData{
				NtagType:          ntagType,
				TotalMemory:       int32(totalMemory),
				UserMemory:        int32(userMemory),
				PasswordProtected: false,
				Counter:           0,
			},
		},
		Capabilities: &pb.CardCapabilities{
			Writable:     true,
			RequiresAuth: false,
			SupportsNdef: true,
			MaxDataSize:  int32(userMemory),
			ReadOnly:     false,
		},
	}

	// If full read requested, read NDEF records
	if options != nil && options.ReadFullData {
		// Placeholder for NDEF reading
		// Real implementation would:
		// 1. Read NDEF capability container
		// 2. Parse NDEF message
		// 3. Extract NDEF records

		cardData.NdefRecords = []*pb.NDEFRecord{}
	}

	return cardData, nil
}

func (h *ntagHandler) WriteCard(conn Connection, uid []byte, blocks []*pb.CardBlock, credentials *pb.AuthenticationCredentials) error {
	h.logger.Debug("writing NTAG card",
		telemetry.String("uid", hex.EncodeToString(uid)),
		telemetry.Int("blocks", len(blocks)),
	)

	// Real implementation would:
	// 1. Format NDEF message
	// 2. Write pages (4 bytes each)
	// 3. Update NDEF length

	return fmt.Errorf("NTAG writing not yet fully implemented")
}

func (h *ntagHandler) AuthenticateCard(conn Connection, uid []byte, credentials *pb.AuthenticationCredentials) (bool, error) {
	// NTAG cards can have password protection
	// For now, assume no password
	return true, nil
}

func (h *ntagHandler) GetCardType() pb.CardType {
	return pb.CardType_CARD_TYPE_NTAG_213
}

func (h *ntagHandler) GetProtocolName() string {
	return "NTAG (NFC)"
}

// ==================================================
// DESFire Handler
// ==================================================

type desfireHandler struct {
	logger *telemetry.Logger
}

// NewDESFireHandler creates a new DESFire handler
func NewDESFireHandler(logger *telemetry.Logger) CardProtocolHandler {
	return &desfireHandler{logger: logger}
}

func (h *desfireHandler) SupportsCard(uid []byte, atr []byte) bool {
	// DESFire cards have 7-byte UIDs and specific ATR
	return len(uid) == 7
}

func (h *desfireHandler) ReadCard(conn Connection, uid []byte, options *pb.ReadOptions) (*pb.CardData, error) {
	h.logger.Debug("reading DESFire card",
		telemetry.String("uid", hex.EncodeToString(uid)),
	)

	cardData := &pb.CardData{
		Type:         pb.CardType_CARD_TYPE_MIFARE_DESFIRE,
		Uid:          uid,
		SerialNumber: hex.EncodeToString(uid),
		Technology:   pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ,
		ProtocolData: &pb.CardData_Desfire{
			Desfire: &pb.DESFireCardData{
				Applications:      make([]*pb.DESFireApplication, 0),
				MasterKeySettings: []byte{},
				FreeMemory:        4096,
			},
		},
		Capabilities: &pb.CardCapabilities{
			Writable:           true,
			RequiresAuth:       true,
			SupportsEncryption: true,
			SupportsNdef:       true,
			MaxDataSize:        4096,
		},
	}

	return cardData, nil
}

func (h *desfireHandler) WriteCard(conn Connection, uid []byte, blocks []*pb.CardBlock, credentials *pb.AuthenticationCredentials) error {
	return fmt.Errorf("DESFire writing not yet fully implemented")
}

func (h *desfireHandler) AuthenticateCard(conn Connection, uid []byte, credentials *pb.AuthenticationCredentials) (bool, error) {
	// DESFire uses complex authentication (DES, 3DES, AES)
	return false, fmt.Errorf("DESFire authentication not yet fully implemented")
}

func (h *desfireHandler) GetCardType() pb.CardType {
	return pb.CardType_CARD_TYPE_MIFARE_DESFIRE
}

func (h *desfireHandler) GetProtocolName() string {
	return "Mifare DESFire"
}

// ==================================================
// FeliCa Handler (Sony)
// ==================================================

type felicaHandler struct {
	logger *telemetry.Logger
}

// NewFeliCaHandler creates a new FeliCa handler
func NewFeliCaHandler(logger *telemetry.Logger) CardProtocolHandler {
	return &felicaHandler{logger: logger}
}

func (h *felicaHandler) SupportsCard(uid []byte, atr []byte) bool {
	// FeliCa cards have 8-byte IDm
	return len(uid) == 8
}

func (h *felicaHandler) ReadCard(conn Connection, uid []byte, options *pb.ReadOptions) (*pb.CardData, error) {
	h.logger.Debug("reading FeliCa card",
		telemetry.String("uid", hex.EncodeToString(uid)),
	)

	cardData := &pb.CardData{
		Type:         pb.CardType_CARD_TYPE_FELICA,
		Uid:          uid,
		SerialNumber: hex.EncodeToString(uid),
		Technology:   pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ,
		ProtocolData: &pb.CardData_Felica{
			Felica: &pb.FeliCaCardData{
				Idm:        uid,
				Pmm:        make([]byte, 8),
				SystemCode: 0x88B4, // Common system code
				Services:   make([]*pb.FeliCaService, 0),
			},
		},
		Capabilities: &pb.CardCapabilities{
			Writable:     true,
			RequiresAuth: false,
			SupportsNdef: true,
			MaxDataSize:  1024,
		},
	}

	return cardData, nil
}

func (h *felicaHandler) WriteCard(conn Connection, uid []byte, blocks []*pb.CardBlock, credentials *pb.AuthenticationCredentials) error {
	return fmt.Errorf("FeliCa writing not yet fully implemented")
}

func (h *felicaHandler) AuthenticateCard(conn Connection, uid []byte, credentials *pb.AuthenticationCredentials) (bool, error) {
	// FeliCa typically doesn't require authentication for basic operations
	return true, nil
}

func (h *felicaHandler) GetCardType() pb.CardType {
	return pb.CardType_CARD_TYPE_FELICA
}

func (h *felicaHandler) GetProtocolName() string {
	return "FeliCa"
}
