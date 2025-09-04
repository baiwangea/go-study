package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- Configuration ---
func loadConfig() (string, int64) {
	botToken := os.Getenv("TG_KEYS")
	if botToken == "" {
		log.Fatal("Environment variable TG_KEYS is not set.")
	}
	chatIDStr := os.Getenv("CHAT_ID")
	if chatIDStr == "" {
		log.Fatal("Environment variable CHAT_ID is not set.")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid CHAT_ID: %v. It must be an integer.", err)
	}
	return botToken, chatID
}

// --- Messaging Helpers ---

// escapeMarkdownV2 escapes characters for Telegram's MarkdownV2 format.
func escapeMarkdownV2(s string) string {
	// List of characters to escape for MarkdownV2
	escapeChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	var replacerArgs []string
	for _, char := range escapeChars {
		// In the Go code, we need \\ to represent a literal backslash
		replacerArgs = append(replacerArgs, char, "\\"+char)
	}
	return strings.NewReplacer(replacerArgs...).Replace(s)
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// --- Elegant Report Template ---

// sendWalletReport generates and sends a beautifully formatted wallet report.
func sendWalletReport(bot *tgbotapi.BotAPI, chatID int64) {
	// 1. Prepare the data
	address := "0x14Ceb52a53DC887941cdE7936D026d4Fdd842D23"
	txCount := 484
	totalUSDT := 250000.117184
	totalUSD := 250000.12
	avgTx := 516.529168
	startDate := "2025-07-16"
	endDate := "2025-08-28"
	generationTime := time.Now().Format("2006-01-02 15:04:05")

	// 2. Build the report string
	var report strings.Builder
	report.WriteString("📈 *BSC Wallet USDT Inflow Report*\n\n")
	report.WriteString(fmt.Sprintf("🏦 *Wallet Address:* `%s`\n", address))
	report.WriteString(fmt.Sprintf("📊 *Total Transactions:* %d\n", txCount))

	// Let the helper function handle the escaping of all data, both dynamic and static.
	report.WriteString(fmt.Sprintf("💰 *Total USDT:* %s USDT\n", escapeMarkdownV2(fmt.Sprintf("%.6f", totalUSDT))))
	report.WriteString(fmt.Sprintf("💵 *Total USD Value:* $%s\n", escapeMarkdownV2(fmt.Sprintf("%.2f", totalUSD))))
	report.WriteString(fmt.Sprintf("📊 *Average Transaction:* %s USDT\n", escapeMarkdownV2(fmt.Sprintf("%.6f", avgTx))))
	report.WriteString(fmt.Sprintf("📅 *Period:* %s to %s\n\n", escapeMarkdownV2(startDate), escapeMarkdownV2(endDate)))

	// FINAL FIX: Instead of manual escaping, we pass the raw string to our helper function.
	report.WriteString(fmt.Sprintf("⏰ *Generated At:* %s %s\n\n", escapeMarkdownV2(generationTime), escapeMarkdownV2("(UTC+8)")))
	report.WriteString(escapeMarkdownV2("📋 *Details in the attached Excel file*."))

	// 3. Send the message
	sendMessage(bot, chatID, report.String())
	log.Println("Sent a wallet report message.")
}

// sendReportWithAttachment demonstrates sending a report and then a file.
func sendReportWithAttachment(bot *tgbotapi.BotAPI, chatID int64) {
	// 1. Send the text report first
	sendWalletReport(bot, chatID)

	// 2. Create a dummy Excel file for demonstration
	fileName := "report_2025-08-28.xlsx"
	_ = os.WriteFile(fileName, []byte("This is a dummy excel file content."), 0644)

	// 3. Send the file
	doc := tgbotapi.FilePath(fileName)
	docMsg := tgbotapi.NewDocument(chatID, doc)
	_, err := bot.Send(docMsg)
	if err != nil {
		log.Printf("Error sending document: %v", err)
	}
	log.Println("Sent a report attachment.")
}

func main() {
	botToken, chatID := loadConfig()

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// --- Run the elegant report demonstration ---
	fmt.Println("Sending elegant report...")
	sendReportWithAttachment(bot, chatID)

	fmt.Println("\nAll messages sent. Check your Telegram chat!")
}
