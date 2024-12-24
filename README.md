# Email Scheduler

A Go application that sends scheduled emails from a list of messages every day at 7 AM.

## Setup

1. Make sure you have Go installed on your system
2. Clone this repository
3. Copy `.env.example` to `.env` and update the values:
   - `SMTP_HOST`: SMTP server host
   - `SMTP_PORT`: SMTP server port
   - `SMTP_USER`: Your email address
   - `SMTP_PASS`: Your email app password
   - `TO_EMAIL`: Recipient email address
   - `EMAIL_SUBJECT`: Subject line for the emails
   - `MESSAGES_FILE`: Path to your messages JSON file

4. Create a `messages.json` file with an array of messages you want to send

## Running the Application

```bash
go run main.go
```

The application will start and send emails daily at 7 AM. Each email will be sent as a reply to the previous one, maintaining the email thread.

## Message Format

The `messages.json` file should contain an array of strings:

```json
[
    "Message 1",
    "Message 2",
    "Message 3"
]
```

Messages will be sent in order, and the application will loop back to the beginning when all messages have been sent. # timemachine
