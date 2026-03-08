package notify

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	// Register pure Go sqlite driver
	_ "modernc.org/sqlite"
)

// WhatsAppClient wraps the whatsmeow Client and provides high-level notification features.
type WhatsAppClient struct {
	Client *whatsmeow.Client
}

// InitWhatsApp initializes the WhatsApp client with a local SQLite session store.
// dbPath is the path to the sqlite file, e.g., "config/whatsapp_store.db"
func InitWhatsApp(dbPath string) (*WhatsAppClient, error) {
	// Ensure the directory for the db exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for whatsapp store: %w", err)
	}

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// modernc.org/sqlite registers the driver name as "sqlite"
	container, err := sqlstore.New(context.Background(), "sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath), dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to whatsapp session store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get first device store: %w", err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return &WhatsAppClient{
		Client: client,
	}, nil
}

// Authenticate handles the terminal QR code login process or simple connect if already logged in.
func (wa *WhatsAppClient) Authenticate() error {
	if wa.Client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := wa.Client.GetQRChannel(context.Background())
		err := wa.Client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect for login: %w", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code in the terminal
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("Please scan the QR code above with your WhatsApp app.")
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err := wa.Client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		fmt.Println("WhatsApp client connected successfully.")
	}
	return nil
}

// Disconnect cleans up the client connection
func (wa *WhatsAppClient) Disconnect() {
	wa.Client.Disconnect()
}
