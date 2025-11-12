package payment_tcp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"
)

// Connection represents a TCP connection to a payment terminal
type Connection struct {
	config        ConnectionConfig
	conn          net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
	mu            sync.Mutex
	connected     bool
	lastActivity  time.Time
	reconnecting  bool
}

// NewConnection creates a new payment terminal connection
func NewConnection(config ConnectionConfig) *Connection {
	// Apply defaults
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = DefaultReadTimeout
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = DefaultWriteTimeout
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultMaxRetries
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = DefaultRetryDelay
	}

	return &Connection{
		config:    config,
		connected: false,
	}
}

// Connect establishes a connection to the payment terminal
func (c *Connection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil // Already connected
	}

	address := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	var conn net.Conn
	var err error

	// Connect with timeout
	dialer := &net.Dialer{
		Timeout: c.config.Timeout,
	}

	if c.config.UseTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.config.TLSSkipVerify,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		// Plain TCP connection
		conn, err = dialer.Dial("tcp", address)
	}

	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	c.connected = true
	c.lastActivity = time.Now()

	return nil
}

// Disconnect closes the connection
func (c *Connection) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("error closing connection: %w", err)
		}
		c.conn = nil
	}

	c.connected = false
	c.reader = nil
	c.writer = nil

	return nil
}

// IsConnected returns true if the connection is established
func (c *Connection) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Send sends a message and waits for a response
func (c *Connection) Send(message []byte) ([]byte, error) {
	// Retry logic
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(c.config.RetryDelay)

			// Try to reconnect
			if err := c.Reconnect(); err != nil {
				lastErr = fmt.Errorf("reconnect failed: %w", err)
				continue
			}
		}

		// Send and receive
		response, err := c.sendInternal(message)
		if err != nil {
			lastErr = err
			continue
		}

		return response, nil
	}

	return nil, fmt.Errorf("send failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// sendInternal sends a message and waits for a response (no retry logic)
func (c *Connection) sendInternal(message []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	// Set write deadline
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout)); err != nil {
		return nil, fmt.Errorf("error setting write deadline: %w", err)
	}

	// Send message length (4 bytes, big-endian)
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte((len(message) >> 24) & 0xFF)
	lengthBytes[1] = byte((len(message) >> 16) & 0xFF)
	lengthBytes[2] = byte((len(message) >> 8) & 0xFF)
	lengthBytes[3] = byte(len(message) & 0xFF)

	if _, err := c.writer.Write(lengthBytes); err != nil {
		c.connected = false
		return nil, fmt.Errorf("error writing length: %w", err)
	}

	// Send message
	if _, err := c.writer.Write(message); err != nil {
		c.connected = false
		return nil, fmt.Errorf("error writing message: %w", err)
	}

	if err := c.writer.Flush(); err != nil {
		c.connected = false
		return nil, fmt.Errorf("error flushing: %w", err)
	}

	c.lastActivity = time.Now()

	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout)); err != nil {
		return nil, fmt.Errorf("error setting read deadline: %w", err)
	}

	// Read response length (4 bytes)
	responseLengthBytes := make([]byte, 4)
	if _, err := c.reader.Read(responseLengthBytes); err != nil {
		c.connected = false
		return nil, fmt.Errorf("error reading response length: %w", err)
	}

	responseLength := int(responseLengthBytes[0])<<24 |
		int(responseLengthBytes[1])<<16 |
		int(responseLengthBytes[2])<<8 |
		int(responseLengthBytes[3])

	// Validate response length
	if responseLength <= 0 || responseLength > 10000 {
		c.connected = false
		return nil, fmt.Errorf("invalid response length: %d", responseLength)
	}

	// Read response
	response := make([]byte, responseLength)
	if _, err := c.reader.Read(response); err != nil {
		c.connected = false
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	c.lastActivity = time.Now()

	return response, nil
}

// Reconnect closes the existing connection and establishes a new one
func (c *Connection) Reconnect() error {
	c.mu.Lock()
	reconnecting := c.reconnecting
	c.reconnecting = true
	c.mu.Unlock()

	// Avoid concurrent reconnections
	if reconnecting {
		// Wait for ongoing reconnection
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			c.mu.Lock()
			stillReconnecting := c.reconnecting
			connected := c.connected
			c.mu.Unlock()

			if !stillReconnecting || connected {
				return nil
			}
		}
		return fmt.Errorf("reconnection timeout")
	}

	defer func() {
		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()

	// Close existing connection
	if err := c.Disconnect(); err != nil {
		// Log error but continue with reconnection
	}

	// Establish new connection
	return c.Connect()
}

// Ping sends a network management message to check connectivity
func (c *Connection) Ping() error {
	// Build network management request (echo test)
	msg := NewISO8583Message(MTINetworkManagementRequest)
	msg.SetField(Field7_TransmissionDateTime, time.Now().Format("0102150405"))
	msg.SetField(Field11_STAN, fmt.Sprintf("%06d", 1))

	msgBytes, err := msg.Pack()
	if err != nil {
		return fmt.Errorf("error packing ping message: %w", err)
	}

	// Send ping
	response, err := c.Send(msgBytes)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Parse response
	responseMsg := NewISO8583Message("")
	if err := responseMsg.Unpack(response); err != nil {
		return fmt.Errorf("error parsing ping response: %w", err)
	}

	// Check MTI
	if responseMsg.MTI != MTINetworkManagementResponse {
		return fmt.Errorf("unexpected response MTI: %s", responseMsg.MTI)
	}

	// Check response code
	if responseCode, ok := responseMsg.GetField(Field39_ResponseCode); ok {
		if responseCode != "00" {
			return fmt.Errorf("ping failed with response code: %s", responseCode)
		}
	}

	return nil
}

// GetLastActivity returns the timestamp of the last successful communication
func (c *Connection) GetLastActivity() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastActivity
}

// GetConfig returns the connection configuration
func (c *Connection) GetConfig() ConnectionConfig {
	return c.config
}
