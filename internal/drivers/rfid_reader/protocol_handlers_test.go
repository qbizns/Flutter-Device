package rfid_reader

import (
	"testing"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxHandler(t *testing.T) {
	logger := telemetry.NewLogger("test")
	handler := NewProxHandler(logger)

	t.Run("SupportsCard", func(t *testing.T) {
		// Valid Prox card UID (4-8 bytes)
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04}, nil))
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, nil))

		// Invalid (too short)
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02}, nil))

		// Invalid (too long)
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, nil))
	})

	t.Run("ReadCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x0A, 0x12, 0x34, 0x56}

		card, err := handler.ReadCard(conn, uid, nil)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.Equal(t, pb.CardType_CARD_TYPE_HID_PROX, card.Type)
		assert.Equal(t, uid, card.Uid)
		assert.Equal(t, pb.CardTechnology_CARD_TECHNOLOGY_PROX_125KHZ, card.Technology)
		assert.NotNil(t, card.GetProx())
		assert.Equal(t, int32(10), card.GetProx().FacilityCode)
		assert.False(t, card.Capabilities.Writable)
		assert.True(t, card.Capabilities.ReadOnly)
	})

	t.Run("WriteCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04}

		// Prox cards are read-only
		err := handler.WriteCard(conn, uid, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read-only")
	})

	t.Run("AuthenticateCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04}

		// Prox cards don't require authentication
		authenticated, err := handler.AuthenticateCard(conn, uid, nil)
		require.NoError(t, err)
		assert.True(t, authenticated)
	})

	t.Run("GetCardType", func(t *testing.T) {
		assert.Equal(t, pb.CardType_CARD_TYPE_HID_PROX, handler.GetCardType())
	})

	t.Run("GetProtocolName", func(t *testing.T) {
		assert.Equal(t, "HID Prox (125 kHz)", handler.GetProtocolName())
	})
}

func TestMifareHandler(t *testing.T) {
	logger := telemetry.NewLogger("test")
	handler := NewMifareHandler(logger)

	t.Run("SupportsCard", func(t *testing.T) {
		// Valid Mifare UID (4 bytes)
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04}, nil))

		// Valid Mifare UID (7 bytes)
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, nil))

		// Invalid lengths
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03}, nil))
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05}, nil))
	})

	t.Run("ReadCard_Basic", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04}

		card, err := handler.ReadCard(conn, uid, nil)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.Equal(t, pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K, card.Type)
		assert.Equal(t, uid, card.Uid)
		assert.Equal(t, pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ, card.Technology)
		assert.NotNil(t, card.GetMifare())
		assert.Equal(t, int32(1024), card.GetMifare().SizeBytes)
		assert.Equal(t, int32(16), card.GetMifare().NumSectors)
		assert.True(t, card.Capabilities.Writable)
		assert.True(t, card.Capabilities.RequiresAuth)
	})

	t.Run("ReadCard_FullData", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04}

		options := &pb.ReadOptions{
			ReadFullData: true,
			AuthKeys:     [][]byte{{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		}

		card, err := handler.ReadCard(conn, uid, options)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.NotEmpty(t, card.GetMifare().Sectors)
		assert.Len(t, card.GetMifare().Sectors, 16) // 16 sectors for 1K card
		assert.Len(t, card.GetMifare().Sectors[0].Blocks, 4) // 4 blocks per sector
	})

	t.Run("AuthenticateCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04}

		// No keys provided
		_, err := handler.AuthenticateCard(conn, uid, nil)
		assert.Error(t, err)

		// With keys
		creds := &pb.AuthenticationCredentials{
			MifareKeysA: [][]byte{{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		}
		authenticated, err := handler.AuthenticateCard(conn, uid, creds)
		require.NoError(t, err)
		assert.True(t, authenticated)
	})

	t.Run("GetCardType", func(t *testing.T) {
		assert.Equal(t, pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K, handler.GetCardType())
	})

	t.Run("GetProtocolName", func(t *testing.T) {
		assert.Equal(t, "Mifare Classic", handler.GetProtocolName())
	})
}

func TestIClassHandler(t *testing.T) {
	logger := telemetry.NewLogger("test")
	handler := NewIClassHandler(logger)

	t.Run("SupportsCard", func(t *testing.T) {
		// Valid iClass UID (8 bytes)
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, nil))

		// Invalid lengths
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04}, nil))
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, nil))
	})

	t.Run("ReadCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		card, err := handler.ReadCard(conn, uid, nil)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.Equal(t, pb.CardType_CARD_TYPE_HID_ICLASS, card.Type)
		assert.Equal(t, uid, card.Uid)
		assert.Equal(t, pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ, card.Technology)
		assert.NotNil(t, card.GetIclass())
		assert.True(t, card.Capabilities.Writable)
		assert.True(t, card.Capabilities.RequiresAuth)
	})

	t.Run("AuthenticateCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		// No key provided
		_, err := handler.AuthenticateCard(conn, uid, nil)
		assert.Error(t, err)

		// With key
		creds := &pb.AuthenticationCredentials{
			IclassKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		}
		authenticated, err := handler.AuthenticateCard(conn, uid, creds)
		require.NoError(t, err)
		assert.True(t, authenticated)
	})

	t.Run("GetCardType", func(t *testing.T) {
		assert.Equal(t, pb.CardType_CARD_TYPE_HID_ICLASS, handler.GetCardType())
	})

	t.Run("GetProtocolName", func(t *testing.T) {
		assert.Equal(t, "HID iClass", handler.GetProtocolName())
	})
}

func TestNTAGHandler(t *testing.T) {
	logger := telemetry.NewLogger("test")
	handler := NewNTAGHandler(logger)

	t.Run("SupportsCard", func(t *testing.T) {
		// Valid NTAG UID (7 bytes)
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, nil))

		// Invalid lengths
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04}, nil))
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, nil))
	})

	t.Run("ReadCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x04, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}

		card, err := handler.ReadCard(conn, uid, nil)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.Equal(t, pb.CardType_CARD_TYPE_NTAG_213, card.Type)
		assert.Equal(t, uid, card.Uid)
		assert.Equal(t, pb.CardTechnology_CARD_TECHNOLOGY_NFC, card.Technology)
		assert.NotNil(t, card.GetNtag())
		assert.Equal(t, "NTAG213", card.GetNtag().NtagType)
		assert.True(t, card.Capabilities.Writable)
		assert.True(t, card.Capabilities.SupportsNdef)
		assert.False(t, card.Capabilities.ReadOnly)
	})

	t.Run("ReadCard_WithNDEF", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x04, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}

		options := &pb.ReadOptions{
			ReadFullData: true,
		}

		card, err := handler.ReadCard(conn, uid, options)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.NotNil(t, card.NdefRecords)
	})

	t.Run("AuthenticateCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x04, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}

		// NTAG cards typically don't require authentication
		authenticated, err := handler.AuthenticateCard(conn, uid, nil)
		require.NoError(t, err)
		assert.True(t, authenticated)
	})

	t.Run("GetCardType", func(t *testing.T) {
		assert.Equal(t, pb.CardType_CARD_TYPE_NTAG_213, handler.GetCardType())
	})

	t.Run("GetProtocolName", func(t *testing.T) {
		assert.Equal(t, "NTAG (NFC)", handler.GetProtocolName())
	})
}

func TestDESFireHandler(t *testing.T) {
	logger := telemetry.NewLogger("test")
	handler := NewDESFireHandler(logger)

	t.Run("SupportsCard", func(t *testing.T) {
		// Valid DESFire UID (7 bytes)
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, nil))

		// Invalid lengths
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04}, nil))
	})

	t.Run("ReadCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x04, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}

		card, err := handler.ReadCard(conn, uid, nil)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.Equal(t, pb.CardType_CARD_TYPE_MIFARE_DESFIRE, card.Type)
		assert.Equal(t, uid, card.Uid)
		assert.Equal(t, pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ, card.Technology)
		assert.NotNil(t, card.GetDesfire())
		assert.True(t, card.Capabilities.Writable)
		assert.True(t, card.Capabilities.RequiresAuth)
		assert.True(t, card.Capabilities.SupportsEncryption)
	})

	t.Run("GetCardType", func(t *testing.T) {
		assert.Equal(t, pb.CardType_CARD_TYPE_MIFARE_DESFIRE, handler.GetCardType())
	})

	t.Run("GetProtocolName", func(t *testing.T) {
		assert.Equal(t, "Mifare DESFire", handler.GetProtocolName())
	})
}

func TestFeliCaHandler(t *testing.T) {
	logger := telemetry.NewLogger("test")
	handler := NewFeliCaHandler(logger)

	t.Run("SupportsCard", func(t *testing.T) {
		// Valid FeliCa IDm (8 bytes)
		assert.True(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, nil))

		// Invalid lengths
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04}, nil))
		assert.False(t, handler.SupportsCard([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, nil))
	})

	t.Run("ReadCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		card, err := handler.ReadCard(conn, uid, nil)
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.Equal(t, pb.CardType_CARD_TYPE_FELICA, card.Type)
		assert.Equal(t, uid, card.Uid)
		assert.Equal(t, pb.CardTechnology_CARD_TECHNOLOGY_CONTACTLESS_13_56MHZ, card.Technology)
		assert.NotNil(t, card.GetFelica())
		assert.True(t, card.Capabilities.Writable)
		assert.True(t, card.Capabilities.SupportsNdef)
	})

	t.Run("AuthenticateCard", func(t *testing.T) {
		conn := newMockConnection()
		uid := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		// FeliCa doesn't require authentication for basic operations
		authenticated, err := handler.AuthenticateCard(conn, uid, nil)
		require.NoError(t, err)
		assert.True(t, authenticated)
	})

	t.Run("GetCardType", func(t *testing.T) {
		assert.Equal(t, pb.CardType_CARD_TYPE_FELICA, handler.GetCardType())
	})

	t.Run("GetProtocolName", func(t *testing.T) {
		assert.Equal(t, "FeliCa", handler.GetProtocolName())
	})
}

func BenchmarkProxHandler_ReadCard(b *testing.B) {
	logger := telemetry.NewLogger("test")
	handler := NewProxHandler(logger)
	conn := newMockConnection()
	uid := []byte{0x01, 0x02, 0x03, 0x04}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.ReadCard(conn, uid, nil)
	}
}

func BenchmarkMifareHandler_ReadCard(b *testing.B) {
	logger := telemetry.NewLogger("test")
	handler := NewMifareHandler(logger)
	conn := newMockConnection()
	uid := []byte{0x01, 0x02, 0x03, 0x04}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.ReadCard(conn, uid, nil)
	}
}
