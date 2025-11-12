package access_control

import (
	"fmt"
	"net"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"github.com/tarm/serial"
)

// ==================================================
// TCP Connection
// ==================================================

// connectTCP creates a TCP connection to the controller
func connectTCP(address string, port int, logger *telemetry.Logger) (Connection, error) {
	addr := fmt.Sprintf("%s:%d", address, port)

	logger.Debug("connecting to access controller via TCP",
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
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *tcpConnection) IsConnected() bool {
	return c.conn != nil
}

func (c *tcpConnection) SetTimeout(timeout time.Duration) error {
	if c.conn != nil {
		if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
		return c.conn.SetWriteDeadline(time.Now().Add(timeout))
	}
	return fmt.Errorf("connection not established")
}

// ==================================================
// Serial Connection (RS-232/RS-485)
// ==================================================

// connectSerial creates a serial connection to the controller
func connectSerial(device string, baudRate int, logger *telemetry.Logger) (Connection, error) {
	logger.Debug("connecting to access controller via serial",
		telemetry.String("device", device),
		telemetry.Int("baud_rate", baudRate),
	)

	config := &serial.Config{
		Name:        device,
		Baud:        baudRate,
		ReadTimeout: time.Second * 5,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, fmt.Errorf("serial connection failed: %w", err)
	}

	return &serialConnection{
		port:   port,
		config: config,
		logger: logger,
	}, nil
}

// serialConnection implements Connection for serial/RS-485
type serialConnection struct {
	port   *serial.Port
	config *serial.Config
	logger *telemetry.Logger
}

func (c *serialConnection) Write(data []byte) (int, error) {
	if c.port == nil {
		return 0, fmt.Errorf("serial port not open")
	}
	return c.port.Write(data)
}

func (c *serialConnection) Read(buf []byte) (int, error) {
	if c.port == nil {
		return 0, fmt.Errorf("serial port not open")
	}
	return c.port.Read(buf)
}

func (c *serialConnection) Close() error {
	if c.port != nil {
		err := c.port.Close()
		c.port = nil
		return err
	}
	return nil
}

func (c *serialConnection) IsConnected() bool {
	return c.port != nil
}

func (c *serialConnection) SetTimeout(timeout time.Duration) error {
	if c.config != nil {
		c.config.ReadTimeout = timeout
		if c.port != nil {
			c.port.Close()
		}
		port, err := serial.OpenPort(c.config)
		if err != nil {
			return err
		}
		c.port = port
		return nil
	}
	return fmt.Errorf("serial configuration not available")
}
