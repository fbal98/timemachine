package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type MessageQueue struct {
	Messages []string `json:"messages"`
	mu       sync.Mutex
	filePath string
}

func NewMessageQueue(filePath string) (*MessageQueue, error) {
	q := &MessageQueue{
		filePath: filePath,
	}
	
	// Load initial messages
	if err := q.load(); err != nil {
		return nil, err
	}
	
	return q, nil
}

func (q *MessageQueue) load() error {
	data, err := os.ReadFile(q.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, create empty queue
			q.Messages = []string{}
			return q.save()
		}
		return fmt.Errorf("error reading messages file: %v", err)
	}

	return json.Unmarshal(data, &q.Messages)
}

func (q *MessageQueue) save() error {
	data, err := json.MarshalIndent(q.Messages, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling messages: %v", err)
	}

	return os.WriteFile(q.filePath, data, 0644)
}

func (q *MessageQueue) Add(message string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.Messages = append(q.Messages, message)
	return q.save()
}

func (q *MessageQueue) Pop() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.Messages) == 0 {
		return "", fmt.Errorf("queue is empty")
	}

	// Get the first message
	message := q.Messages[0]
	
	// Remove it from the queue
	q.Messages = q.Messages[1:]
	
	// Save the updated queue
	if err := q.save(); err != nil {
		return "", fmt.Errorf("error saving queue after pop: %v", err)
	}

	return message, nil
}

func (q *MessageQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.Messages) == 0
}

func (q *MessageQueue) GetAll() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	// Return a copy to prevent external modifications
	messages := make([]string, len(q.Messages))
	copy(messages, q.Messages)
	return messages
}

type Config struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPass     string `json:"smtp_pass"`
	ToEmail      string `json:"to_email"`
	Subject      string `json:"subject"`
	MessagesFile string `json:"messages_file"`
	CronSchedule string `json:"cron_schedule"`
}

func sendEmail(config Config, message string) error {
	log.Printf("Attempting to send email to %s using SMTP server %s:%d", config.ToEmail, config.SMTPHost, config.SMTPPort)
	
	auth := smtp.PlainAuth("", config.SMTPUser, config.SMTPPass, config.SMTPHost)
	
	// Format email with headers for threading
	emailBody := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"References: <daily-reminder@timemachine>\r\n"+
		"In-Reply-To: <daily-reminder@timemachine>\r\n"+
		"\r\n"+
		"%s", config.SMTPUser, config.ToEmail, config.Subject, message)

	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	to := []string{config.ToEmail}

	log.Printf("Email content prepared, attempting SMTP connection...")
	err := smtp.SendMail(addr, auth, config.SMTPUser, to, []byte(emailBody))
	if err != nil {
		return fmt.Errorf("SMTP error: %v", err)
	}

	log.Printf("Email sent successfully to %s", config.ToEmail)
	return nil
}

func processQueueMessage(config Config, queue *MessageQueue) {
	if queue.IsEmpty() {
		log.Println("No messages left in the queue")
		return
	}
	
	message, err := queue.Pop()
	if err != nil {
		log.Printf("Failed to pop message from queue: %v", err)
		return
	}
	
	if err := sendEmail(config, message); err != nil {
		log.Printf("Failed to send email: %v", err)
		// Add the message back to the queue since we failed to send it
		if addErr := queue.Add(message); addErr != nil {
			log.Printf("Failed to add message back to queue: %v", addErr)
		}
	} else {
		log.Printf("Successfully sent and removed message from queue")
	}
}

func loadConfig() (Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return Config{}, fmt.Errorf("error loading .env file: %v", err)
	}

	config := Config{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     587, // Default port
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASS"),
		ToEmail:      os.Getenv("TO_EMAIL"),
		Subject:      os.Getenv("EMAIL_SUBJECT"),
		MessagesFile: os.Getenv("MESSAGES_FILE"),
		CronSchedule: os.Getenv("CRON_SCHEDULE"),
	}

	// Validate required fields
	if config.SMTPHost == "" {
		return config, fmt.Errorf("SMTP_HOST is required")
	}
	if config.SMTPUser == "" {
		return config, fmt.Errorf("SMTP_USER is required")
	}
	if config.SMTPPass == "" {
		return config, fmt.Errorf("SMTP_PASS is required")
	}
	if config.ToEmail == "" {
		return config, fmt.Errorf("TO_EMAIL is required")
	}

	// Set defaults if not provided
	if config.Subject == "" {
		config.Subject = "Daily Reminder"
	}
	if config.MessagesFile == "" {
		config.MessagesFile = "messages.json"
	}
	if config.CronSchedule == "" {
		config.CronSchedule = "0 7 * * *" // Default to 7 AM daily
	}

	return config, nil
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Time Machine - Add Message</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 20px auto;
            padding: 0 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 20px;
            border-radius: 5px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        form {
            display: flex;
            flex-direction: column;
            gap: 10px;
        }
        textarea {
            width: 100%;
            height: 100px;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            resize: vertical;
        }
        button {
            padding: 10px 20px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        button:hover {
            background-color: #0056b3;
        }
        .messages {
            margin-top: 20px;
        }
        .message {
            padding: 10px;
            border-bottom: 1px solid #eee;
        }
        .success {
            color: green;
            margin-bottom: 10px;
        }
        .error {
            color: red;
            margin-bottom: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Add New Message to Queue</h1>
        {{if .Success}}
            <div class="success">Message added successfully!</div>
        {{end}}
        {{if .Error}}
            <div class="error">{{.Error}}</div>
        {{end}}
        <form method="POST" action="/add">
            <textarea name="message" placeholder="Enter your message here..." required></textarea>
            <button type="submit">Add to Queue</button>
        </form>
        
        <div class="messages">
            <h2>Current Queue ({{len .Messages}} messages)</h2>
            {{range .Messages}}
                <div class="message">{{.}}</div>
            {{end}}
        </div>
    </div>
</body>
</html>
`

type PageData struct {
	Messages []string
	Success  bool
	Error    string
}

func startHTTPServer(queue *MessageQueue) {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Messages: queue.GetAll(),
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		message := r.FormValue("message")
		if message == "" {
			data := PageData{
				Messages: queue.GetAll(),
				Error:    "Message cannot be empty",
			}
			tmpl.Execute(w, data)
			return
		}

		err := queue.Add(message)
		if err != nil {
			data := PageData{
				Messages: queue.GetAll(),
				Error:    fmt.Sprintf("Failed to add message: %v", err),
			}
			tmpl.Execute(w, data)
			return
		}

		data := PageData{
			Messages: queue.GetAll(),
			Success:  true,
		}
		tmpl.Execute(w, data)
	})

	// Start HTTP server on port 8080
	go func() {
		log.Printf("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully:")
	log.Printf("SMTP Host: %s", config.SMTPHost)
	log.Printf("SMTP Port: %d", config.SMTPPort)
	log.Printf("SMTP User: %s", config.SMTPUser)
	log.Printf("To Email: %s", config.ToEmail)
	log.Printf("Messages File: %s", config.MessagesFile)
	log.Printf("Cron Schedule: %s", config.CronSchedule)

	// Initialize message queue
	queue, err := NewMessageQueue(config.MessagesFile)
	if err != nil {
		log.Fatalf("Failed to initialize message queue: %v", err)
	}
	log.Printf("Message queue initialized with %d messages", len(queue.Messages))

	// Start HTTP server
	startHTTPServer(queue)

	// Send first message immediately
	log.Println("Sending first message immediately...")
	processQueueMessage(config, queue)

	// Initialize cron scheduler with Muscat timezone
	muscat, err := time.LoadLocation("Asia/Muscat")
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}
	c := cron.New(cron.WithLocation(muscat))
	
	// Schedule task using the configured schedule
	_, err = c.AddFunc(config.CronSchedule, func() {
		processQueueMessage(config, queue)
	})

	if err != nil {
		log.Fatalf("Error scheduling cron job: %v", err)
	}

	c.Start()

	// Keep the program running
	log.Printf("Email scheduler started. Running on schedule: %s (Asia/Muscat)", config.CronSchedule)
	select {} // Block forever
} 