package badge_printer_zebra

import (
	"fmt"
	"net"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// connectTCP creates a TCP connection to the printer
func connectTCP(address string, port int, logger *telemetry.Logger) (Connection, error) {
	addr := fmt.Sprintf("%s:%d", address, port)

	logger.Debug("connecting to printer via TCP",
		telemetry.String("address", addr),
	)

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("TCP connection failed: %w", err)
	}

	return &tcpConnection{
		conn:   conn,
		logger: logger,
	}, nil
}

// connectUSB creates a USB connection to the printer
func connectUSB(vendorID, productID uint16, logger *telemetry.Logger) (Connection, error) {
	logger.Debug("connecting to printer via USB",
		telemetry.Int("vendor_id", int(vendorID)),
		telemetry.Int("product_id", int(productID)),
	)

	conn := &usbConnection{
		vendorID:  vendorID,
		productID: productID,
		logger:    logger,
	}

	if err := conn.open(); err != nil {
		return nil, fmt.Errorf("USB connection failed: %w", err)
	}

	return conn, nil
}

// tcpConnection implements Connection for TCP
type tcpConnection struct {
	conn   net.Conn
	logger *telemetry.Logger
}

func (c *tcpConnection) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

func (c *tcpConnection) Read(buf []byte) (int, error) {
	return c.conn.Read(buf)
}

func (c *tcpConnection) Close() error {
	return c.conn.Close()
}

func (c *tcpConnection) IsConnected() bool {
	return c.conn != nil
}
